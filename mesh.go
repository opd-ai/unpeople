package unpeople

import "math"

// ─── Vector / Color types ─────────────────────────────────────────────────────
// These types are layout-compatible with kaiju's matrix.Vec2/Vec3/Vec4/Vec4i/Color.
// Because Go's type system does not allow passing []unpeople.Vertex where
// []rendering.Vertex is expected, callers must copy or unsafely reinterpret the
// slice when integrating with the Kaiju rendering API directly.
// See the package-level ToKaijuVertices helper (to be added in Phase 6) for a
// safe conversion path once the Kaiju module is directly importable.

// Vec2 is a 2-component float32 vector (matches kaiju's matrix.Vec2).
type Vec2 [2]float32

// Vec3 is a 3-component float32 vector (matches kaiju's matrix.Vec3).
type Vec3 [3]float32

// Vec4 is a 4-component float32 vector (matches kaiju's matrix.Vec4).
type Vec4 [4]float32

// Vec4i is a 4-component int32 vector (matches kaiju's matrix.Vec4i).
type Vec4i [4]int32

// Color is an RGBA float32 colour (matches kaiju's matrix.Color).
type Color [4]float32

// ColorGray is the default material colour: mid-grey, fully opaque.
var ColorGray = Color{0.5, 0.5, 0.5, 1.0}

// ─── Skin Tone Colors ────────────────────────────────────────────────────────

// skinToneBase defines the base RGB values for each skin tone level.
// Values are designed to span a realistic range from very pale to very dark.
// All colors are in linear RGB space for proper blending.
var skinToneBase = [8]Color{
	{0.96, 0.90, 0.85, 1.0}, // SkinTonePale - very light, almost porcelain
	{0.91, 0.82, 0.74, 1.0}, // SkinToneFair - light with slight warmth
	{0.85, 0.73, 0.62, 1.0}, // SkinToneLight - light beige
	{0.76, 0.62, 0.50, 1.0}, // SkinToneMedium - medium beige
	{0.67, 0.53, 0.40, 1.0}, // SkinToneOlive - olive/tan
	{0.58, 0.44, 0.32, 1.0}, // SkinToneTan - warm tan
	{0.45, 0.32, 0.22, 1.0}, // SkinToneBrown - brown
	{0.30, 0.20, 0.14, 1.0}, // SkinToneDark - deep brown
}

// undertoneShift defines the RGB shift applied for warm/cool undertones.
var undertoneShift = [3]Color{
	{0.0, 0.0, 0.0, 0.0},   // SkinUndertoneNeutral - no shift
	{0.04, 0.01, -0.03, 0}, // SkinUndertoneWarm - slightly more red/yellow
	{-0.02, 0.01, 0.04, 0}, // SkinUndertoneCool - slightly more pink/blue
}

// ComputeSkinColor returns the vertex color for the given skin tone and undertone.
func ComputeSkinColor(tone SkinTone, undertone SkinUndertone) Color {
	base := skinToneBase[tone]
	shift := undertoneShift[undertone]

	// Apply undertone shift and clamp to valid range
	r := clampFloat32(base[0]+shift[0], 0.0, 1.0)
	g := clampFloat32(base[1]+shift[1], 0.0, 1.0)
	b := clampFloat32(base[2]+shift[2], 0.0, 1.0)

	return Color{r, g, b, 1.0}
}

// clampFloat32 restricts a float32 value to the given range.
func clampFloat32(v, minV, maxV float32) float32 {
	if v < minV {
		return minV
	}
	if v > maxV {
		return maxV
	}
	return v
}

// ─── Vertex ──────────────────────────────────────────────────────────────────

// Vertex is layout-compatible with kaiju's rendering.Vertex (same field order
// and types).  To pass a []Vertex to kaiju's rendering.NewMesh, callers must
// perform an explicit element-wise copy into a []rendering.Vertex slice, or use
// an unsafe reinterpret cast after verifying struct sizes match.
type Vertex struct {
	Position     Vec3
	Normal       Vec3
	Tangent      Vec4
	UV0          Vec2
	Color        Color
	JointIds     Vec4i
	JointWeights Vec4
	MorphTarget  Vec3
}

// ─── Mesh ────────────────────────────────────────────────────────────────────

// Mesh holds the generated geometry in a format layout-compatible with kaiju's
// rendering.NewMesh(key, vertices, indices).  An explicit type conversion is
// required before passing to the Kaiju API; see the Vertex comment above.
type Mesh struct {
	// Key is a unique string that identifies this mesh variant inside the
	// Kaiju engine's mesh cache.
	Key string
	// Vertices is the vertex buffer. Each element maps 1-to-1 with Kaiju's
	// rendering.Vertex struct.
	Vertices []Vertex
	// Indices is the index buffer (triangles: every 3 entries form one face).
	Indices []uint32
}

// ─── meshBuilder ─────────────────────────────────────────────────────────────

// meshBuilder accumulates geometry from multiple body-part primitives,
// adjusting the index offset on each append so that all parts share a single
// flat vertex/index buffer.
type meshBuilder struct {
	vertices []Vertex
	indices  []uint32
}

func (b *meshBuilder) append(verts []Vertex, idxs []uint32) {
	base := uint32(len(b.vertices))
	b.vertices = append(b.vertices, verts...)

	// Pre-grow the indices slice to avoid repeated allocations in the loop
	if cap(b.indices)-len(b.indices) < len(idxs) {
		newIndices := make([]uint32, len(b.indices), len(b.indices)+len(idxs))
		copy(newIndices, b.indices)
		b.indices = newIndices
	}
	for _, idx := range idxs {
		b.indices = append(b.indices, base+idx)
	}
}

func (b *meshBuilder) build(key string) *Mesh {
	return &Mesh{
		Key:      key,
		Vertices: b.vertices,
		Indices:  b.indices,
	}
}

// ─── Vec3 helpers ─────────────────────────────────────────────────────────────

func vec3Add(a, b Vec3) Vec3 {
	return Vec3{a[0] + b[0], a[1] + b[1], a[2] + b[2]}
}

func vec3Sub(a, b Vec3) Vec3 {
	return Vec3{a[0] - b[0], a[1] - b[1], a[2] - b[2]}
}

func vec3Scale(a Vec3, s float32) Vec3 {
	return Vec3{a[0] * s, a[1] * s, a[2] * s}
}

func vec3Cross(a, b Vec3) Vec3 {
	return Vec3{
		a[1]*b[2] - a[2]*b[1],
		a[2]*b[0] - a[0]*b[2],
		a[0]*b[1] - a[1]*b[0],
	}
}

func vec3Len(a Vec3) float32 {
	return float32(math.Sqrt(float64(a[0]*a[0] + a[1]*a[1] + a[2]*a[2])))
}

func vec3Normalize(a Vec3) Vec3 {
	l := vec3Len(a)
	if l < 1e-7 {
		return Vec3{0, 1, 0}
	}
	return Vec3{a[0] / l, a[1] / l, a[2] / l}
}

// ─── Bilinear Sampling Helpers ────────────────────────────────────────────────

// colorSampler is a function that samples a color at integer pixel coordinates.
type colorSampler func(x, y int) Color

// sampleBilinear performs bilinear interpolation on a grid with the given dimensions.
// The sampler function retrieves pixel colors at integer coordinates.
// UV coordinates are in [0,1] range. interpolateAlpha controls whether the alpha
// channel is interpolated (true) or fixed to 1.0 (false).
func sampleBilinear(u, v float32, width, height int, sampler colorSampler, interpolateAlpha bool) Color {
	u = clampFloat32(u, 0, 1)
	v = clampFloat32(v, 0, 1)

	px := u * float32(width-1)
	py := v * float32(height-1)

	x0 := int(px)
	y0 := int(py)
	x1 := minInt(x0+1, width-1)
	y1 := minInt(y0+1, height-1)
	fx := px - float32(x0)
	fy := py - float32(y0)

	c00 := sampler(x0, y0)
	c10 := sampler(x1, y0)
	c01 := sampler(x0, y1)
	c11 := sampler(x1, y1)

	r := lerpFloat32(lerpFloat32(c00[0], c10[0], fx), lerpFloat32(c01[0], c11[0], fx), fy)
	g := lerpFloat32(lerpFloat32(c00[1], c10[1], fx), lerpFloat32(c01[1], c11[1], fx), fy)
	b := lerpFloat32(lerpFloat32(c00[2], c10[2], fx), lerpFloat32(c01[2], c11[2], fx), fy)

	var a float32
	if interpolateAlpha {
		a = lerpFloat32(lerpFloat32(c00[3], c10[3], fx), lerpFloat32(c01[3], c11[3], fx), fy)
	} else {
		a = 1.0
	}

	return Color{r, g, b, a}
}

// lerpFloat32 linearly interpolates between a and b by t.
func lerpFloat32(a, b, t float32) float32 {
	return a + (b-a)*t
}

// minInt returns the smaller of two integers.
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
