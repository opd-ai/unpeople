//go:build kaiju

// Package kaiju provides integration with the Kaiju game engine for unpeople
// humanoid mesh generation.
//
// This package implements a drop-in Generator that works directly with Kaiju's
// asset pipeline and produces rendering.Mesh objects compatible with Kaiju's
// rendering system.
//
// Build with -tags kaiju to enable this integration:
//
//	go build -tags kaiju ./...
//
// Example usage:
//
//	gen := kaiju.NewKaijuGenerator()
//	params := unpeople.DefaultParams()
//	params.Seed = 42
//	mesh, err := gen.Generate(params)
//	// mesh is *rendering.Mesh ready for Kaiju
package kaiju

import (
	"github.com/opd-ai/unpeople"
	"kaijuengine.com/rendering"
)

// KaijuGenerator wraps unpeople.Generator for direct Kaiju integration.
// It produces rendering.Mesh objects compatible with Kaiju's rendering system.
type KaijuGenerator struct {
	gen *unpeople.Generator
}

// NewKaijuGenerator creates a new generator that produces Kaiju-compatible meshes.
func NewKaijuGenerator() *KaijuGenerator {
	return &KaijuGenerator{
		gen: unpeople.NewGenerator(),
	}
}

// Generate produces a Kaiju rendering.Mesh from the given parameters.
// The returned mesh can be used directly with Kaiju's rendering pipeline.
func (g *KaijuGenerator) Generate(params unpeople.Params) (*rendering.Mesh, error) {
	mesh, err := g.gen.Generate(params)
	if err != nil {
		return nil, err
	}
	return ToKaijuMesh(mesh), nil
}

// ToKaijuMesh converts an unpeople.Mesh to a Kaiju rendering.Mesh.
// The vertex layout is compatible - this performs a type-safe conversion
// without copying vertex data where possible.
func ToKaijuMesh(src *unpeople.Mesh) *rendering.Mesh {
	vertices := make([]rendering.Vertex, len(src.Vertices))
	for i, v := range src.Vertices {
		vertices[i] = rendering.Vertex{
			Position:     rendering.Vec3(v.Position),
			Normal:       rendering.Vec3(v.Normal),
			Tangent:      rendering.Vec4(v.Tangent),
			UV0:          rendering.Vec2(v.UV0),
			UV1:          rendering.Vec2(v.UV1),
			Color:        rendering.Color(v.Color),
			JointIds:     rendering.Vec4i(v.JointIds),
			JointWeights: rendering.Vec4(v.JointWeights),
			MorphTarget:  rendering.Vec3(v.MorphTarget),
		}
	}

	return &rendering.Mesh{
		Vertices: vertices,
		Indices:  src.Indices,
		Key:      src.Key,
	}
}

// ToUnpeopleMesh converts a Kaiju rendering.Mesh back to an unpeople.Mesh.
// This is useful for round-trip operations or serialization.
func ToUnpeopleMesh(src *rendering.Mesh) *unpeople.Mesh {
	vertices := make([]unpeople.Vertex, len(src.Vertices))
	for i, v := range src.Vertices {
		vertices[i] = unpeople.Vertex{
			Position:     unpeople.Vec3(v.Position),
			Normal:       unpeople.Vec3(v.Normal),
			Tangent:      unpeople.Vec4(v.Tangent),
			UV0:          unpeople.Vec2(v.UV0),
			UV1:          unpeople.Vec2(v.UV1),
			Color:        unpeople.Color(v.Color),
			JointIds:     unpeople.Vec4i(v.JointIds),
			JointWeights: unpeople.Vec4(v.JointWeights),
			MorphTarget:  unpeople.Vec3(v.MorphTarget),
		}
	}

	return &unpeople.Mesh{
		Vertices: vertices,
		Indices:  src.Indices,
		Key:      src.Key,
	}
}
