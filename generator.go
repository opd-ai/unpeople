package unpeople

import (
	"fmt"
	"math/rand"
)

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
//
// The returned *Mesh is never nil when err is nil; its Vertices and Indices
// slices are non-empty and all index values are within the vertex range.
func (g *Generator) Generate(p Params) (*Mesh, error) {
	if err := p.Validate(); err != nil {
		return nil, fmt.Errorf("unpeople: invalid params: %w", err)
	}

	// Seeded, deterministic PRNG – must be created fresh each call so that
	// the same seed always produces the same sequence.
	rng := rand.New(rand.NewSource(p.Seed)) //nolint:gosec // intentional seeded RNG

	layout := computeBodyLayout(&p, rng)

	// Mesh key encodes the primary visual parameters; the Kaiju engine uses
	// this string for its mesh cache.
	key := fmt.Sprintf(
		"humanoid_sp%d_ht%d_bl%d_pr%d_ph%d_ag%d_po%d_se%d",
		p.Species, p.Height, p.Build, p.Proportions,
		p.Phenotype, p.Age, p.Posture, p.Seed,
	)

	return buildMesh(layout, key), nil
}
