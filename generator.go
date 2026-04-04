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

// buildOptions bundles optional mesh generation flags
type buildOptions struct {
	hasHairSlot bool
	faceShape   FaceShape
	jaw         Jaw
	brow        Brow
	skinColor   Color // Computed skin color for all vertices
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

	// Convert bools to int for key encoding
	hairSlot := 0
	if p.HasHairSlot {
		hairSlot = 1
	}
	mergeVerts := 0
	if p.MergeVertices {
		mergeVerts = 1
	}

	// Mesh key encodes all geometry-affecting parameters so that the Kaiju
	// engine's mesh cache never reuses a mesh for a different parameter set.
	// Skin tone affects vertex colors, so it must be part of the key.
	key := fmt.Sprintf(
		"humanoid_sp%d_ht%d_bl%d_pr%d_ph%d_ag%d_po%d"+
			"_fs%d_jw%d_br%d_er%d"+
			"_sw%d_hw%d_ll%d_nl%d"+
			"_hs%d_fl%d_ft%d_hr%d_sk%d_ut%d_se%d_mv%d",
		p.Species, p.Height, p.Build, p.Proportions, p.Phenotype, p.Age, p.Posture,
		p.FaceShape, p.Jaw, p.Brow, p.Ears,
		p.ShoulderWidth, p.HipWidth, p.LimbLength, p.NeckLength,
		p.HandSize, p.FingerLength, p.FootSize, hairSlot,
		p.SkinTone, p.SkinUndertone, p.Seed, mergeVerts,
	)

	opts := buildOptions{
		hasHairSlot: p.HasHairSlot,
		faceShape:   p.FaceShape,
		jaw:         p.Jaw,
		brow:        p.Brow,
		skinColor:   ComputeSkinColor(p.SkinTone, p.SkinUndertone),
	}

	mesh := buildMesh(layout, key, opts)

	// Apply vertex merging if requested
	if p.MergeVertices {
		epsilon := scaledEpsilon(layout.totalHeight)
		mesh = MergeNearbyVertices(mesh, epsilon)
	}

	return mesh, nil
}
