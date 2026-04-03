// Package unpeople – procedural skin texture generation
//
// This file implements noise-driven procedural skin texture generation.
// Textures include freckles, blemishes, and age spots that are deterministically
// generated based on parameters. The texture is baked to an atlas matching the
// UV layout from atlas.go.
package unpeople

import "math"

// ─── Texture Types ───────────────────────────────────────────────────────────

// Texture represents a procedurally generated skin texture atlas.
// Pixels are RGBA colors in row-major order.
type Texture struct {
	Width  int
	Height int
	Pixels []Color
}

// SkinTextureParams configures procedural skin texture generation.
type SkinTextureParams struct {
	// Base skin color (from SkinTone/SkinUndertone)
	BaseColor Color
	// Age determines presence of age spots and skin texture variation
	Age Age
	// Seed for deterministic noise generation
	Seed int64
	// FreckleIntensity controls freckle density (0 = none, 1 = heavy)
	FreckleIntensity float32
	// BlemishIntensity controls blemish density (0 = none, 1 = heavy)
	BlemishIntensity float32
}

// ─── Default Texture Parameters ──────────────────────────────────────────────

// DefaultSkinTextureParams returns texture parameters appropriate for the
// character's skin tone, age, and other traits.
func DefaultSkinTextureParams(tone SkinTone, undertone SkinUndertone, age Age, seed int64) SkinTextureParams {
	baseColor := ComputeSkinColor(tone, undertone)

	// Freckle intensity based on skin tone (lighter = more visible freckles)
	freckleIntensity := computeFreckleIntensity(tone)

	// Blemish intensity based on age (teens have more, elderly have spots)
	blemishIntensity := computeBlemishIntensity(age)

	return SkinTextureParams{
		BaseColor:        baseColor,
		Age:              age,
		Seed:             seed,
		FreckleIntensity: freckleIntensity,
		BlemishIntensity: blemishIntensity,
	}
}

// computeFreckleIntensity returns freckle visibility based on skin tone.
func computeFreckleIntensity(tone SkinTone) float32 {
	switch tone {
	case SkinTonePale:
		return 0.6
	case SkinToneFair:
		return 0.5
	case SkinToneLight:
		return 0.35
	case SkinToneMedium:
		return 0.2
	case SkinToneOlive:
		return 0.15
	case SkinToneTan:
		return 0.1
	case SkinToneBrown:
		return 0.05
	case SkinToneDark:
		return 0.02
	default:
		return 0.2
	}
}

// computeBlemishIntensity returns blemish/age-spot intensity based on age.
func computeBlemishIntensity(age Age) float32 {
	switch age {
	case AgeToddler, AgeChild:
		return 0.05 // Clear skin
	case AgeTeen:
		return 0.25 // Acne-prone
	case AgeYouth:
		return 0.15
	case AgeAdult:
		return 0.1
	case AgeOld:
		return 0.2 // Starting to get age spots
	case AgeElderly:
		return 0.35 // Liver spots, age marks
	case AgeDecrepit:
		return 0.5 // Heavy age spotting
	default:
		return 0.1
	}
}

// ─── Texture Generation ──────────────────────────────────────────────────────

// textureGenContext holds precomputed data for texture generation.
type textureGenContext struct {
	params    SkinTextureParams
	freckles  []featurePoint
	blemishes []featurePoint
	uvAtlas   UVAtlas
}

// newTextureGenContext initializes the texture generation context.
func newTextureGenContext(params SkinTextureParams, rng *splitmix64) *textureGenContext {
	return &textureGenContext{
		params:    params,
		freckles:  generateFrecklePositions(rng, params.FreckleIntensity, 256),
		blemishes: generateBlemishPositions(rng, params.BlemishIntensity, params.Age, 128),
		uvAtlas:   defaultUVAtlas(),
	}
}

// isFreckleRegion returns true if the UV coordinate is in a freckle-visible area.
func (ctx *textureGenContext) isFreckleRegion(u, v float32) bool {
	return isInRegion(u, v, ctx.uvAtlas.Face) || isInRegion(u, v, ctx.uvAtlas.Head) ||
		isInRegion(u, v, ctx.uvAtlas.UpperArmL) || isInRegion(u, v, ctx.uvAtlas.UpperArmR)
}

// isBlemishRegion returns true if the UV coordinate is in a blemish-visible area.
func (ctx *textureGenContext) isBlemishRegion(u, v float32) bool {
	return isInRegion(u, v, ctx.uvAtlas.Face) || isInRegion(u, v, ctx.uvAtlas.Head) ||
		isInRegion(u, v, ctx.uvAtlas.HandL) || isInRegion(u, v, ctx.uvAtlas.HandR)
}

// isVeinRegion returns true if veins should be visible at this UV coordinate.
func (ctx *textureGenContext) isVeinRegion(u, v float32) bool {
	return ctx.params.Age >= AgeOld &&
		(isInRegion(u, v, ctx.uvAtlas.HandL) || isInRegion(u, v, ctx.uvAtlas.HandR))
}

// computePixelColor generates the skin color for a single pixel.
func (ctx *textureGenContext) computePixelColor(u, v float32) Color {
	color := ctx.params.BaseColor

	noiseVal := perlinNoise2D(u*8, v*8, ctx.params.Seed) * 0.03
	color = colorShift(color, noiseVal)

	if ctx.isFreckleRegion(u, v) {
		freckleVal := sampleFreckles(u, v, ctx.freckles)
		color = applyFreckles(color, freckleVal, ctx.params.FreckleIntensity)
	}

	if ctx.isBlemishRegion(u, v) {
		blemishVal := sampleBlemishes(u, v, ctx.blemishes, ctx.params.Age)
		color = applyBlemishes(color, blemishVal, ctx.params.Age)
	}

	if ctx.isVeinRegion(u, v) {
		veinVal := generateVeinPattern(u, v, ctx.params.Seed)
		color = applyVeins(color, veinVal, ctx.params.Age)
	}

	return color
}

// GenerateSkinTexture creates a procedural skin texture atlas with the given
// dimensions. The texture includes subtle color variations, freckles (for
// lighter skin tones), and age-related blemishes.
func GenerateSkinTexture(params SkinTextureParams, width, height int) *Texture {
	rng := newSplitmix64(params.Seed)
	ctx := newTextureGenContext(params, rng)

	tex := &Texture{
		Width:  width,
		Height: height,
		Pixels: make([]Color, width*height),
	}

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			u := float32(x) / float32(width-1)
			v := float32(y) / float32(height-1)
			tex.Pixels[y*width+x] = ctx.computePixelColor(u, v)
		}
	}

	return tex
}

// ─── Feature Positions ───────────────────────────────────────────────────────

type featurePoint struct {
	u, v   float32
	radius float32
	dark   float32 // Darkness multiplier
}

// generateFrecklePositions creates a set of freckle locations.
func generateFrecklePositions(rng *splitmix64, intensity float32, maxCount int) []featurePoint {
	count := int(float32(maxCount) * intensity)
	points := make([]featurePoint, count)
	for i := 0; i < count; i++ {
		points[i] = featurePoint{
			u:      rng.Float32(),
			v:      rng.Float32(),
			radius: 0.003 + rng.Float32()*0.007, // 0.3% to 1% of texture
			dark:   0.6 + rng.Float32()*0.3,     // 60-90% darkening
		}
	}
	return points
}

// generateBlemishPositions creates age spots and blemish locations.
func generateBlemishPositions(rng *splitmix64, intensity float32, age Age, maxCount int) []featurePoint {
	count := int(float32(maxCount) * intensity)
	points := make([]featurePoint, count)

	// Age spots are larger and darker for elderly
	sizeMultiplier := float32(1.0)
	if age >= AgeElderly {
		sizeMultiplier = 1.5
	}

	for i := 0; i < count; i++ {
		points[i] = featurePoint{
			u:      rng.Float32(),
			v:      rng.Float32(),
			radius: (0.005 + rng.Float32()*0.015) * sizeMultiplier,
			dark:   0.7 + rng.Float32()*0.25,
		}
	}
	return points
}

// ─── Feature Sampling ────────────────────────────────────────────────────────

// computeFeatureContrib calculates the contribution of a feature point at a UV coordinate.
func computeFeatureContrib(u, v float32, f featurePoint) float32 {
	du := u - f.u
	dv := v - f.v
	distSq := du*du + dv*dv
	radiusSq := f.radius * f.radius

	if distSq >= radiusSq {
		return 0
	}

	t := 1.0 - distSq/radiusSq
	return t * t * f.dark
}

// sampleFreckles computes freckle contribution at a UV coordinate.
func sampleFreckles(u, v float32, freckles []featurePoint) float32 {
	var maxContrib float32
	for _, f := range freckles {
		if contrib := computeFeatureContrib(u, v, f); contrib > maxContrib {
			maxContrib = contrib
		}
	}
	return maxContrib
}

// computeBlemishFalloff calculates the falloff value for a blemish based on age.
func computeBlemishFalloff(t float32, age Age) float32 {
	if age >= AgeElderly {
		return t * t * t // Sharper falloff for age spots
	}
	return t * t
}

// computeBlemishContrib calculates the contribution of a blemish at a UV coordinate.
func computeBlemishContrib(u, v float32, b featurePoint, age Age) float32 {
	du := u - b.u
	dv := v - b.v
	distSq := du*du + dv*dv
	radiusSq := b.radius * b.radius

	if distSq >= radiusSq {
		return 0
	}

	t := 1.0 - distSq/radiusSq
	t = computeBlemishFalloff(t, age)
	return t * b.dark
}

// sampleBlemishes computes blemish contribution at a UV coordinate.
func sampleBlemishes(u, v float32, blemishes []featurePoint, age Age) float32 {
	var maxContrib float32
	for _, b := range blemishes {
		if contrib := computeBlemishContrib(u, v, b, age); contrib > maxContrib {
			maxContrib = contrib
		}
	}
	return maxContrib
}

// ─── Feature Application ─────────────────────────────────────────────────────

// applyFreckles darkens the color based on freckle contribution.
func applyFreckles(color Color, freckleVal, intensity float32) Color {
	if freckleVal < 0.01 {
		return color
	}
	// Freckles are darker, slightly warmer spots
	darken := 1.0 - freckleVal*intensity*0.3
	return Color{
		color[0] * darken,
		color[1] * darken * 0.95, // Slightly less green
		color[2] * darken * 0.90, // Less blue (warmer)
		color[3],
	}
}

// applyBlemishes applies age spots or blemishes.
func applyBlemishes(color Color, blemishVal float32, age Age) Color {
	if blemishVal < 0.01 {
		return color
	}

	// Different coloration based on age
	if age >= AgeElderly {
		// Liver spots: brownish
		return Color{
			color[0] * (1.0 - blemishVal*0.25),
			color[1] * (1.0 - blemishVal*0.3),
			color[2] * (1.0 - blemishVal*0.35),
			color[3],
		}
	} else if age == AgeTeen {
		// Acne: slightly reddish
		return Color{
			clampFloat32(color[0]*(1.0+blemishVal*0.1), 0, 1),
			color[1] * (1.0 - blemishVal*0.1),
			color[2] * (1.0 - blemishVal*0.1),
			color[3],
		}
	}
	// Generic blemish: slight darkening
	darken := 1.0 - blemishVal*0.15
	return Color{color[0] * darken, color[1] * darken, color[2] * darken, color[3]}
}

// applyVeins adds subtle vein visibility for elderly characters.
func applyVeins(color Color, veinVal float32, age Age) Color {
	if veinVal < 0.01 {
		return color
	}

	// Vein intensity increases with age
	intensity := float32(0.05)
	if age == AgeElderly {
		intensity = 0.1
	} else if age == AgeDecrepit {
		intensity = 0.15
	}

	// Veins add blue-purple tint
	return Color{
		color[0] * (1.0 - veinVal*intensity*0.3),
		color[1] * (1.0 - veinVal*intensity*0.2),
		clampFloat32(color[2]*(1.0+veinVal*intensity*0.15), 0, 1),
		color[3],
	}
}

// ─── Noise Functions ─────────────────────────────────────────────────────────

// perlinNoise2D generates 2D Perlin-like noise for organic texture variation.
// Returns a value in [-1, 1].
func perlinNoise2D(x, y float32, seed int64) float32 {
	// Simple value noise with interpolation
	ix := int(x)
	iy := int(y)
	fx := x - float32(ix)
	fy := y - float32(iy)

	// Get random values at grid corners
	v00 := hashToFloat(ix, iy, seed)
	v10 := hashToFloat(ix+1, iy, seed)
	v01 := hashToFloat(ix, iy+1, seed)
	v11 := hashToFloat(ix+1, iy+1, seed)

	// Smooth interpolation
	fx = smoothstep(fx)
	fy = smoothstep(fy)

	// Bilinear interpolation
	v0 := v00 + (v10-v00)*fx
	v1 := v01 + (v11-v01)*fx
	return v0 + (v1-v0)*fy
}

// hashToFloat converts grid coordinates to a pseudo-random float in [-1, 1].
func hashToFloat(x, y int, seed int64) float32 {
	// Simple hash combining coordinates and seed
	h := uint64(x)*374761393 + uint64(y)*668265263 + uint64(seed)
	h = (h ^ (h >> 13)) * 1274126177
	h = h ^ (h >> 16)
	return float32(h%1000001)/500000.0 - 1.0
}

// smoothstep provides smooth interpolation curve.
func smoothstep(t float32) float32 {
	return t * t * (3.0 - 2.0*t)
}

// generateVeinPattern creates a vein-like branching pattern.
func generateVeinPattern(u, v float32, seed int64) float32 {
	// Use multi-scale noise to create vein-like structures
	val1 := perlinNoise2D(u*5, v*5, seed)
	val2 := perlinNoise2D(u*10, v*10, seed+1) * 0.5
	val3 := perlinNoise2D(u*20, v*20, seed+2) * 0.25

	combined := val1 + val2 + val3

	// Threshold to create line-like structures
	threshold := float32(0.3)
	if combined > threshold {
		return (combined - threshold) / (1.0 - threshold)
	}
	return 0
}

// ─── Utility Functions ───────────────────────────────────────────────────────

// colorShift adjusts a color by a noise value.
func colorShift(color Color, shift float32) Color {
	return Color{
		clampFloat32(color[0]+shift, 0, 1),
		clampFloat32(color[1]+shift, 0, 1),
		clampFloat32(color[2]+shift, 0, 1),
		color[3],
	}
}

// isInRegion checks if UV coordinates fall within a UV region.
func isInRegion(u, v float32, region UVRegion) bool {
	return u >= region.UMin && u <= region.UMax &&
		v >= region.VMin && v <= region.VMax
}

// ─── Full Texture Atlas Generation ───────────────────────────────────────────

// GenerateSkinTextureFromParams creates a skin texture using character params.
func GenerateSkinTextureFromParams(p Params, width, height int) *Texture {
	params := DefaultSkinTextureParams(p.SkinTone, p.SkinUndertone, p.Age, p.Seed)
	return GenerateSkinTexture(params, width, height)
}

// MeshWithTextures bundles a mesh with all texture maps (albedo and normal).
type MeshWithTextures struct {
	Mesh          *Mesh
	Material      Material
	AlbedoTexture *Texture
	NormalTexture *NormalMap
}

// GenerateWithTextures produces a complete mesh with all texture maps.
// This is the most comprehensive output format for fully-textured characters.
func (g *Generator) GenerateWithTextures(p Params) (*MeshWithTextures, error) {
	mesh, err := g.Generate(p)
	if err != nil {
		return nil, err
	}

	skinColor := ComputeSkinColor(p.SkinTone, p.SkinUndertone)
	material := BuildMaterial(skinColor, p.Age, p.Build)

	// Generate textures at 512x512 (reasonable default)
	albedo := GenerateSkinTextureFromParams(p, 512, 512)
	normal := GenerateMusculatureAtlas(p.Build, p.Seed, 512, 512)

	return &MeshWithTextures{
		Mesh:          mesh,
		Material:      material,
		AlbedoTexture: albedo,
		NormalTexture: normal,
	}, nil
}

// ─── Texture Accessors ───────────────────────────────────────────────────────

// At returns the color at the given pixel coordinates.
func (t *Texture) At(x, y int) Color {
	if x < 0 || x >= t.Width || y < 0 || y >= t.Height {
		return Color{0.5, 0.5, 0.5, 1.0}
	}
	return t.Pixels[y*t.Width+x]
}

// SampleBilinear samples the texture with bilinear interpolation.
// UV coordinates are in [0,1] range.
func (t *Texture) SampleBilinear(u, v float32) Color {
	return sampleBilinear(u, v, t.Width, t.Height, t.At, true)
}

// ─── Export Utilities ────────────────────────────────────────────────────────

// ToRGBA8 converts the texture to 8-bit RGBA bytes (for export/display).
// The returned slice has length Width * Height * 4 (RGBA order).
func (t *Texture) ToRGBA8() []byte {
	data := make([]byte, t.Width*t.Height*4)
	for i, c := range t.Pixels {
		data[i*4+0] = uint8(clampFloat32(c[0], 0, 1) * 255)
		data[i*4+1] = uint8(clampFloat32(c[1], 0, 1) * 255)
		data[i*4+2] = uint8(clampFloat32(c[2], 0, 1) * 255)
		data[i*4+3] = uint8(clampFloat32(c[3], 0, 1) * 255)
	}
	return data
}

// ToRGBA8 converts the normal map to 8-bit RGBA bytes (for export/display).
func (nm *NormalMap) ToRGBA8() []byte {
	data := make([]byte, nm.Width*nm.Height*4)
	for i, c := range nm.Pixels {
		data[i*4+0] = uint8(clampFloat32(c[0], 0, 1) * 255)
		data[i*4+1] = uint8(clampFloat32(c[1], 0, 1) * 255)
		data[i*4+2] = uint8(clampFloat32(c[2], 0, 1) * 255)
		data[i*4+3] = uint8(clampFloat32(c[3], 0, 1) * 255)
	}
	return data
}

// HasFeatures returns true if this texture has visible skin features
// (freckles, blemishes, etc.) beyond just the base color.
func (p SkinTextureParams) HasFeatures() bool {
	return p.FreckleIntensity > 0.05 || p.BlemishIntensity > 0.05
}

// ─── Noise Helpers ───────────────────────────────────────────────────────────

// absFloat32 returns the absolute value of a float32.
func absFloat32(x float32) float32 {
	if x < 0 {
		return -x
	}
	return x
}

// fractalNoise generates multi-octave noise for richer texture detail.
func fractalNoise(u, v float32, octaves int, seed int64) float32 {
	var total float32
	amplitude := float32(1.0)
	frequency := float32(1.0)
	maxValue := float32(0.0)

	for i := 0; i < octaves; i++ {
		total += perlinNoise2D(u*frequency, v*frequency, seed+int64(i)) * amplitude
		maxValue += amplitude
		amplitude *= 0.5
		frequency *= 2.0
	}

	return total / maxValue
}

// voronoiNoise generates cell-like noise for spots and other features.
func voronoiNoise(u, v float32, seed int64) float32 {
	// Grid cell
	iu := int(math.Floor(float64(u)))
	iv := int(math.Floor(float64(v)))
	fu := u - float32(iu)
	fv := v - float32(iv)

	// Search neighborhood for closest feature point
	minDist := float32(2.0)
	for dx := -1; dx <= 1; dx++ {
		for dy := -1; dy <= 1; dy++ {
			// Random point in neighboring cell
			px := float32(dx) + hashToFloat(iu+dx, iv+dy, seed)*0.5 + 0.5
			py := float32(dy) + hashToFloat(iu+dx, iv+dy, seed+1)*0.5 + 0.5
			d := (fu-px)*(fu-px) + (fv-py)*(fv-py)
			if d < minDist {
				minDist = d
			}
		}
	}
	return float32(math.Sqrt(float64(minDist)))
}
