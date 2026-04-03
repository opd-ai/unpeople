// Package main provides tests for the example CLI demonstration.
package main

import (
	"testing"

	"github.com/opd-ai/unpeople"
)

// TestExampleGeneration verifies that all example character configurations
// generate valid meshes without errors.
func TestExampleGeneration(t *testing.T) {
	g := unpeople.NewGenerator()

	examples := []struct {
		label  string
		params unpeople.Params
	}{
		{
			"Adult Human (default)",
			unpeople.DefaultParams(),
		},
		{
			"Tall Athletic Elf – pointed ears",
			func() unpeople.Params {
				p := unpeople.DefaultParams()
				p.Seed = 100
				p.Species = unpeople.SpeciesElf
				p.Height = unpeople.HeightTall
				p.Build = unpeople.BuildAthletic
				p.Phenotype = unpeople.PhenotypeFeminine
				p.Ears = unpeople.EarsPointed
				p.FaceShape = unpeople.FaceShapeOval
				p.LimbLength = unpeople.LimbLengthLong
				return p
			}(),
		},
		{
			"Stocky Dwarf – broad shoulders",
			func() unpeople.Params {
				p := unpeople.DefaultParams()
				p.Seed = 200
				p.Species = unpeople.SpeciesDwarf
				p.Height = unpeople.HeightShort
				p.Build = unpeople.BuildStocky
				p.Phenotype = unpeople.PhenotypeMasculine
				p.ShoulderWidth = unpeople.ShoulderWidthBroad
				p.FaceShape = unpeople.FaceShapeSquare
				p.Jaw = unpeople.JawProminent
				return p
			}(),
		},
		{
			"Giant Muscular Troll",
			func() unpeople.Params {
				p := unpeople.DefaultParams()
				p.Seed = 300
				p.Species = unpeople.SpeciesTroll
				p.Height = unpeople.HeightGiant
				p.Build = unpeople.BuildMuscular
				p.Proportions = unpeople.ProportionsHeroic
				p.NeckLength = unpeople.NeckLengthThick
				p.HandSize = unpeople.HandSizeLarge
				p.FootSize = unpeople.FootSizeLarge
				return p
			}(),
		},
		{
			"Toddler Gnome – caricature",
			func() unpeople.Params {
				p := unpeople.DefaultParams()
				p.Seed = 400
				p.Species = unpeople.SpeciesGnome
				p.Age = unpeople.AgeToddler
				p.Proportions = unpeople.ProportionsCaricature
				p.Height = unpeople.HeightTiny
				p.Build = unpeople.BuildFragile
				return p
			}(),
		},
		{
			"Elderly Frail Orc – hunched",
			func() unpeople.Params {
				p := unpeople.DefaultParams()
				p.Seed = 500
				p.Species = unpeople.SpeciesOrc
				p.Age = unpeople.AgeElderly
				p.Build = unpeople.BuildFragile
				p.Posture = unpeople.PostureHunched
				p.LimbLength = unpeople.LimbLengthShort
				return p
			}(),
		},
		{
			"Athletic Halfling – heroic proportions",
			func() unpeople.Params {
				p := unpeople.DefaultParams()
				p.Seed = 600
				p.Species = unpeople.SpeciesHalfling
				p.Build = unpeople.BuildAthletic
				p.Proportions = unpeople.ProportionsHeroic
				p.HipWidth = unpeople.HipWidthNarrow
				p.ShoulderWidth = unpeople.ShoulderWidthBroad
				return p
			}(),
		},
		{
			"Lean Feminine Kobold – stylized",
			func() unpeople.Params {
				p := unpeople.DefaultParams()
				p.Seed = 700
				p.Species = unpeople.SpeciesKobold
				p.Build = unpeople.BuildLean
				p.Phenotype = unpeople.PhenotypeFeminine
				p.Proportions = unpeople.ProportionsStylized
				p.NeckLength = unpeople.NeckLengthLong
				p.FingerLength = unpeople.FingerLengthLong
				return p
			}(),
		},
	}

	for _, ex := range examples {
		t.Run(ex.label, func(t *testing.T) {
			m, err := g.Generate(ex.params)
			if err != nil {
				t.Fatalf("Generate failed: %v", err)
			}
			if len(m.Vertices) == 0 {
				t.Error("Expected non-empty vertices")
			}
			if len(m.Indices) == 0 {
				t.Error("Expected non-empty indices")
			}
			if m.Key == "" {
				t.Error("Expected non-empty mesh key")
			}
			// Verify indices are in bounds
			for i, idx := range m.Indices {
				if int(idx) >= len(m.Vertices) {
					t.Errorf("Index %d out of bounds at position %d: %d >= %d",
						idx, i, idx, len(m.Vertices))
				}
			}
		})
	}
}

// TestExampleBuildReproducibility ensures the example produces consistent output
// when run multiple times with the same parameters.
func TestExampleBuildReproducibility(t *testing.T) {
	g := unpeople.NewGenerator()
	p := unpeople.DefaultParams()
	p.Seed = 12345

	m1, err := g.Generate(p)
	if err != nil {
		t.Fatalf("First generation failed: %v", err)
	}

	m2, err := g.Generate(p)
	if err != nil {
		t.Fatalf("Second generation failed: %v", err)
	}

	if len(m1.Vertices) != len(m2.Vertices) {
		t.Errorf("Vertex count mismatch: %d vs %d", len(m1.Vertices), len(m2.Vertices))
	}
	if len(m1.Indices) != len(m2.Indices) {
		t.Errorf("Index count mismatch: %d vs %d", len(m1.Indices), len(m2.Indices))
	}
	if m1.Key != m2.Key {
		t.Errorf("Key mismatch: %s vs %s", m1.Key, m2.Key)
	}
}
