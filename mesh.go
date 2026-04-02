package unpeople

import "math"

// ─── Vector / Color types ─────────────────────────────────────────────────────
// These mirror the layout of kaiju's matrix.Vec2/Vec3/Vec4/Vec4i/Color so that
// a Mesh produced here can be consumed directly by kaiju's rendering.NewMesh.

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

// ─── Vertex ──────────────────────────────────────────────────────────────────

// Vertex matches the memory layout of kaiju's rendering.Vertex exactly.
// Kaiju's rendering.NewMesh accepts a []Vertex and []uint32 with this layout.
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

// Mesh holds the generated geometry ready for use with kaiju's
// rendering.NewMesh(key, vertices, indices).
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
