package unpeople

import "fmt"

// Generator creates deterministic procedural humanoid meshes.
// A single Generator instance is safe for concurrent use from multiple
// goroutines because it holds no mutable state.
type Generator struct{}

// NewGenerator constructs a Generator ready for use.
func NewGenerator() *Generator {
	return &Generator{}
}

// Generate produces a humanoid Mesh from the supplied parameters.
//
// Determinism guarantee: given the same Params (including Seed), Generate
// always returns a Mesh with an identical vertex and index buffer.
// Determinism is guaranteed across Go versions because the internal PRNG
// (splitmix64) is implemented locally in this package with a fixed algorithm.
//
// The returned *Mesh is never nil when err is nil; its Vertices and Indices
// slices are non-empty and all index values are within the vertex range.
func (g *Generator) Generate(p Params) (*Mesh, error) {
	if err := p.Validate(); err != nil {
		return nil, fmt.Errorf("unpeople: invalid params: %w", err)
	}

	// Package-local fixed-algorithm PRNG; must be created fresh each call so
	// that the same seed always produces the same sequence.
	rng := newSplitmix64(p.Seed)

	layout := computeBodyLayout(&p, rng)

	// Mesh key encodes all geometry-affecting parameters so that the Kaiju
	// engine's mesh cache never reuses a mesh for a different parameter set.
	key := fmt.Sprintf(
		"humanoid_sp%d_ht%d_bl%d_pr%d_ph%d_ag%d_po%d"+
			"_fs%d_jw%d_br%d_er%d"+
			"_sw%d_hw%d_ll%d_nl%d"+
			"_hs%d_fl%d_ft%d_se%d",
		p.Species, p.Height, p.Build, p.Proportions, p.Phenotype, p.Age, p.Posture,
		p.FaceShape, p.Jaw, p.Brow, p.Ears,
		p.ShoulderWidth, p.HipWidth, p.LimbLength, p.NeckLength,
		p.HandSize, p.FingerLength, p.FootSize, p.Seed,
	)

	return buildMesh(layout, key), nil
}
