//go:build !kaiju

// Package kaiju provides integration with the Kaiju game engine for unpeople
// humanoid mesh generation.
//
// This stub exists so the package can be imported without the kaiju build tag.
// The actual implementation requires building with -tags kaiju and having
// the kaijuengine.com/rendering package available.
//
// Build with -tags kaiju to enable full Kaiju integration:
//
//	go build -tags kaiju ./...
package kaiju

import "github.com/opd-ai/unpeople"

// KaijuGenerator is a placeholder type when built without the kaiju tag.
// Use -tags kaiju to get the full implementation.
type KaijuGenerator struct {
	gen *unpeople.Generator
}

// NewKaijuGenerator creates a new generator.
// Note: This is a stub implementation. Build with -tags kaiju for full support.
func NewKaijuGenerator() *KaijuGenerator {
	return &KaijuGenerator{
		gen: unpeople.NewGenerator(),
	}
}

// Generate produces an unpeople.Mesh from the given parameters.
// Note: This is a stub implementation. Build with -tags kaiju to get
// *rendering.Mesh output directly.
func (g *KaijuGenerator) Generate(params unpeople.Params) (*unpeople.Mesh, error) {
	return g.gen.Generate(params)
}
