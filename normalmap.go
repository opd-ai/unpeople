// Package unpeople – procedural normal map generation
//
// This file implements procedural normal map generation for musculature detail.
// Normal maps simulate surface detail (muscle definition) without adding geometry.
// The intensity of muscle detail is driven by the Build parameter.
package unpeople

import "math"

// ─── Normal Map Types ────────────────────────────────────────────────────────

// NormalMap represents a procedurally generated normal map for musculature.
// Values are in tangent space: R=X, G=Y, B=Z where (0.5, 0.5, 1.0) is flat.
type NormalMap struct {
	Width  int
	Height int
	Pixels []Color // RGBA pixels in row-major order
}

// MuscleDefinition describes the intensity of musculature detail for normal maps.
type MuscleDefinition int

const (
	MuscleDefinitionNone MuscleDefinition = iota
	MuscleDefinitionSubtle
	MuscleDefinitionModerate
	MuscleDefinitionPronounced
	MuscleDefinitionExtreme
)

// MusculatureParams configures the procedural normal map generation.
type MusculatureParams struct {
	// Definition controls the overall intensity of muscle detail
	Definition MuscleDefinition
	// BodyPart specifies which body part's musculature to generate
	BodyPart MusculatureBodyPart
	// Seed for deterministic noise generation
	Seed int64
}

// MusculatureBodyPart identifies body regions with distinct muscle patterns.
type MusculatureBodyPart int

const (
	MusculatureChest MusculatureBodyPart = iota
	MusculatureAbdomen
	MusculatureUpperArm
	MusculatureForearm
	MusculatureUpperLeg
	MusculatureLowerLeg
	MusculatureBack
	MusculatureShoulder
)

// ─── Build to Muscle Definition Mapping ──────────────────────────────────────

// BuildToMuscleDefinition maps a Build enum to the appropriate muscle definition.
// Muscular builds show pronounced definition; average and lean show moderate;
// fragile builds show minimal muscle detail.
func BuildToMuscleDefinition(b Build) MuscleDefinition {
	switch b {
	case BuildMuscular:
		return MuscleDefinitionPronounced
	case BuildAthletic:
		return MuscleDefinitionModerate
	case BuildAverage:
		return MuscleDefinitionSubtle
	case BuildLean:
		return MuscleDefinitionModerate // Lean shows definition due to low body fat
	case BuildStocky:
		return MuscleDefinitionSubtle // Muscle obscured by mass
	case BuildFragile:
		return MuscleDefinitionNone
	default:
		return MuscleDefinitionSubtle
	}
}

// ─── Normal Map Generation ───────────────────────────────────────────────────

// GenerateMusculatureNormalMap creates a procedural normal map for the specified
// body part with muscle detail driven by the definition intensity.
//
// The generated normal map uses tangent space encoding where:
// - R (X): horizontal displacement (-1 to +1 mapped to 0-1)
// - G (Y): vertical displacement (-1 to +1 mapped to 0-1)
// - B (Z): surface normal (always positive, typically 0.5-1.0)
// - A: unused (set to 1.0)
func GenerateMusculatureNormalMap(params MusculatureParams, width, height int) *NormalMap {
	rng := newSplitmix64(params.Seed)
	nm := &NormalMap{
		Width:  width,
		Height: height,
		Pixels: make([]Color, width*height),
	}

	// Get muscle pattern for the body part
	pattern := getMusclePattern(params.BodyPart)

	// Definition intensity multiplier (0.0 to 1.0)
	intensity := definitionIntensity(params.Definition)

	// Generate normal map
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			u := float32(x) / float32(width-1)
			v := float32(y) / float32(height-1)

			// Compute muscle bump at this UV coordinate
			nx, ny := computeMuscleBump(u, v, pattern, intensity, rng)

			// Encode to normal map color
			nm.Pixels[y*width+x] = encodeNormal(nx, ny)
		}
	}

	return nm
}

// definitionIntensity returns a 0-1 multiplier for muscle definition.
func definitionIntensity(d MuscleDefinition) float32 {
	switch d {
	case MuscleDefinitionNone:
		return 0.0
	case MuscleDefinitionSubtle:
		return 0.15
	case MuscleDefinitionModerate:
		return 0.35
	case MuscleDefinitionPronounced:
		return 0.55
	case MuscleDefinitionExtreme:
		return 0.75
	default:
		return 0.15
	}
}

// encodeNormal converts a tangent-space normal to RGBA color.
// nx and ny are the X and Y components (-1 to 1), Z is computed.
func encodeNormal(nx, ny float32) Color {
	// Compute Z component (normal always points somewhat outward)
	nz := float32(math.Sqrt(float64(1.0 - nx*nx - ny*ny)))
	if nz < 0.0 {
		nz = 0.0
	}

	// Map from [-1,1] to [0,1]
	return Color{
		0.5 + nx*0.5,
		0.5 + ny*0.5,
		0.5 + nz*0.5,
		1.0,
	}
}

// ─── Muscle Patterns ─────────────────────────────────────────────────────────

// musclePattern defines the muscle groups and their bump shapes for a body part.
type musclePattern struct {
	groups []muscleGroup
}

// muscleGroup represents a single muscle or muscle group as a bump function.
type muscleGroup struct {
	// Center position in UV space
	centerU, centerV float32
	// Size of the muscle bump (radius in UV space)
	radiusU, radiusV float32
	// Bump direction (affects normal orientation)
	directionX, directionY float32
	// Relative intensity (0-1)
	intensity float32
}

// getMusclePattern returns the predefined muscle pattern for a body part.
func getMusclePattern(part MusculatureBodyPart) musclePattern {
	switch part {
	case MusculatureChest:
		return musclePattern{groups: []muscleGroup{
			// Pectoralis major (left)
			{0.35, 0.55, 0.20, 0.18, -0.3, 0.2, 1.0},
			// Pectoralis major (right)
			{0.65, 0.55, 0.20, 0.18, 0.3, 0.2, 1.0},
			// Sternum center line (subtle depression)
			{0.50, 0.55, 0.05, 0.30, 0.0, 0.0, -0.3},
		}}
	case MusculatureAbdomen:
		return musclePattern{groups: []muscleGroup{
			// Rectus abdominis (6-pack) - upper pair
			{0.42, 0.75, 0.08, 0.10, -0.15, 0.1, 0.9},
			{0.58, 0.75, 0.08, 0.10, 0.15, 0.1, 0.9},
			// Middle pair
			{0.42, 0.55, 0.08, 0.10, -0.15, 0.0, 0.85},
			{0.58, 0.55, 0.08, 0.10, 0.15, 0.0, 0.85},
			// Lower pair
			{0.42, 0.35, 0.08, 0.10, -0.15, -0.1, 0.8},
			{0.58, 0.35, 0.08, 0.10, 0.15, -0.1, 0.8},
			// Obliques (left)
			{0.25, 0.50, 0.10, 0.25, -0.4, 0.0, 0.6},
			// Obliques (right)
			{0.75, 0.50, 0.10, 0.25, 0.4, 0.0, 0.6},
			// Linea alba (center line)
			{0.50, 0.50, 0.03, 0.35, 0.0, 0.0, -0.2},
		}}
	case MusculatureUpperArm:
		return musclePattern{groups: []muscleGroup{
			// Biceps brachii
			{0.50, 0.45, 0.25, 0.30, 0.0, 0.3, 1.0},
			// Triceps (back)
			{0.50, 0.75, 0.22, 0.25, 0.0, -0.2, 0.8},
			// Deltoid transition
			{0.50, 0.15, 0.30, 0.12, 0.0, 0.4, 0.7},
		}}
	case MusculatureForearm:
		return musclePattern{groups: []muscleGroup{
			// Brachioradialis
			{0.40, 0.30, 0.15, 0.20, -0.2, 0.15, 0.8},
			// Extensor group
			{0.60, 0.40, 0.18, 0.25, 0.2, 0.1, 0.7},
			// Flexor group
			{0.40, 0.60, 0.18, 0.30, -0.15, -0.1, 0.75},
		}}
	case MusculatureUpperLeg:
		return musclePattern{groups: []muscleGroup{
			// Quadriceps - rectus femoris
			{0.50, 0.50, 0.20, 0.35, 0.0, 0.15, 1.0},
			// Vastus lateralis (outer)
			{0.70, 0.45, 0.15, 0.30, 0.25, 0.1, 0.85},
			// Vastus medialis (inner)
			{0.30, 0.45, 0.15, 0.30, -0.25, 0.1, 0.85},
			// Hamstrings (back)
			{0.50, 0.80, 0.22, 0.18, 0.0, -0.2, 0.7},
		}}
	case MusculatureLowerLeg:
		return musclePattern{groups: []muscleGroup{
			// Gastrocnemius (calf) - medial head
			{0.40, 0.35, 0.18, 0.25, -0.15, 0.2, 1.0},
			// Gastrocnemius - lateral head
			{0.60, 0.35, 0.18, 0.25, 0.15, 0.2, 1.0},
			// Tibialis anterior
			{0.35, 0.60, 0.12, 0.25, -0.2, -0.1, 0.6},
			// Soleus
			{0.50, 0.55, 0.20, 0.20, 0.0, 0.1, 0.5},
		}}
	case MusculatureBack:
		return musclePattern{groups: []muscleGroup{
			// Trapezius (upper)
			{0.50, 0.85, 0.35, 0.12, 0.0, 0.3, 0.8},
			// Latissimus dorsi (left)
			{0.30, 0.50, 0.18, 0.30, -0.3, -0.1, 0.9},
			// Latissimus dorsi (right)
			{0.70, 0.50, 0.18, 0.30, 0.3, -0.1, 0.9},
			// Erector spinae (spine)
			{0.45, 0.50, 0.06, 0.40, -0.1, 0.0, 0.6},
			{0.55, 0.50, 0.06, 0.40, 0.1, 0.0, 0.6},
			// Rhomboids
			{0.35, 0.70, 0.10, 0.12, -0.2, 0.15, 0.5},
			{0.65, 0.70, 0.10, 0.12, 0.2, 0.15, 0.5},
		}}
	case MusculatureShoulder:
		return musclePattern{groups: []muscleGroup{
			// Deltoid - anterior
			{0.35, 0.50, 0.15, 0.20, -0.25, 0.2, 0.9},
			// Deltoid - lateral
			{0.50, 0.45, 0.18, 0.22, 0.0, 0.35, 1.0},
			// Deltoid - posterior
			{0.65, 0.50, 0.15, 0.20, 0.25, 0.2, 0.9},
		}}
	default:
		// Flat normal map
		return musclePattern{groups: nil}
	}
}

// computeMuscleBump computes the normal displacement at UV coordinates.
// Returns the X and Y components of the normal in tangent space.
func computeMuscleBump(u, v float32, pattern musclePattern, intensity float32, rng *splitmix64) (nx, ny float32) {
	if intensity < 0.001 || len(pattern.groups) == 0 {
		return 0, 0
	}

	// Add subtle noise for organic variation
	noise := (rng.Float32()*2.0 - 1.0) * 0.02 * intensity

	// Accumulate bump from all muscle groups
	for _, mg := range pattern.groups {
		// Distance from muscle center (normalized by muscle radius)
		du := (u - mg.centerU) / mg.radiusU
		dv := (v - mg.centerV) / mg.radiusV
		distSq := du*du + dv*dv

		// Gaussian falloff
		if distSq < 4.0 { // Only compute if within 2 standard deviations
			falloff := float32(math.Exp(float64(-distSq * 0.5)))
			contribution := falloff * mg.intensity * intensity

			// Add directional bump
			nx += mg.directionX * contribution
			ny += mg.directionY * contribution
		}
	}

	// Add noise
	nx += noise
	ny += noise * 0.7 // Slightly less noise in Y

	// Clamp to valid range
	maxDisp := float32(0.7) // Max normal displacement
	nx = clampFloat32(nx, -maxDisp, maxDisp)
	ny = clampFloat32(ny, -maxDisp, maxDisp)

	return nx, ny
}

// ─── Full Body Normal Map Atlas ──────────────────────────────────────────────

// GenerateMusculatureAtlas generates a complete normal map atlas for the humanoid
// body, with each body part in its corresponding UV region from the atlas.
// The atlas dimensions should match the UV atlas layout.
func GenerateMusculatureAtlas(build Build, seed int64, width, height int) *NormalMap {
	definition := BuildToMuscleDefinition(build)

	atlas := &NormalMap{
		Width:  width,
		Height: height,
		Pixels: make([]Color, width*height),
	}

	// Initialize with flat normal (0.5, 0.5, 1.0)
	flatNormal := Color{0.5, 0.5, 1.0, 1.0}
	for i := range atlas.Pixels {
		atlas.Pixels[i] = flatNormal
	}

	// Get the UV atlas regions
	uvAtlas := defaultUVAtlas()

	// Generate and composite each body part
	bodyParts := []struct {
		part   MusculatureBodyPart
		region UVRegion
	}{
		{MusculatureChest, uvAtlas.Chest},
		{MusculatureAbdomen, uvAtlas.Abdomen},
		{MusculatureUpperArm, uvAtlas.UpperArmL},
		{MusculatureUpperArm, uvAtlas.UpperArmR},
		{MusculatureForearm, uvAtlas.ForearmL},
		{MusculatureForearm, uvAtlas.ForearmR},
		{MusculatureUpperLeg, uvAtlas.UpperLegL},
		{MusculatureUpperLeg, uvAtlas.UpperLegR},
		{MusculatureLowerLeg, uvAtlas.LowerLegL},
		{MusculatureLowerLeg, uvAtlas.LowerLegR},
	}

	for _, bp := range bodyParts {
		params := MusculatureParams{
			Definition: definition,
			BodyPart:   bp.part,
			Seed:       seed + int64(bp.part),
		}
		compositeBodyPart(atlas, bp.region, params)
	}

	return atlas
}

// compositeBodyPart renders a body part's normal map into the atlas.
func compositeBodyPart(atlas *NormalMap, region UVRegion, params MusculatureParams) {
	// Determine pixel bounds from UV region
	x0 := int(float32(atlas.Width) * region.UMin)
	x1 := int(float32(atlas.Width) * region.UMax)
	y0 := int(float32(atlas.Height) * region.VMin)
	y1 := int(float32(atlas.Height) * region.VMax)

	// Region dimensions
	regionWidth := x1 - x0
	regionHeight := y1 - y0
	if regionWidth <= 0 || regionHeight <= 0 {
		return
	}

	// Get muscle pattern and intensity
	pattern := getMusclePattern(params.BodyPart)
	intensity := definitionIntensity(params.Definition)
	rng := newSplitmix64(params.Seed)

	// Generate normals for this region
	for py := y0; py < y1; py++ {
		for px := x0; px < x1; px++ {
			// Local UV within the region [0,1]
			localU := float32(px-x0) / float32(regionWidth-1)
			localV := float32(py-y0) / float32(regionHeight-1)

			nx, ny := computeMuscleBump(localU, localV, pattern, intensity, rng)
			atlas.Pixels[py*atlas.Width+px] = encodeNormal(nx, ny)
		}
	}
}

// ─── Accessors ───────────────────────────────────────────────────────────────

// At returns the normal color at the given pixel coordinates.
func (nm *NormalMap) At(x, y int) Color {
	if x < 0 || x >= nm.Width || y < 0 || y >= nm.Height {
		return Color{0.5, 0.5, 1.0, 1.0} // Flat normal for out-of-bounds
	}
	return nm.Pixels[y*nm.Width+x]
}

// SampleBilinear samples the normal map with bilinear interpolation.
// UV coordinates are in [0,1] range. Alpha is always 1.0 for normal maps.
func (nm *NormalMap) SampleBilinear(u, v float32) Color {
	return sampleBilinear(u, v, nm.Width, nm.Height, nm.At, false)
}
