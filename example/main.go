// Command example demonstrates the unpeople humanoid mesh generator.
//
// It generates several characters with varied parameter combinations and
// prints the resulting mesh statistics.  Use this as a reference for how to
// integrate the generator into a Kaiju-based project.
package main

import (
	"fmt"
	"log"

	"github.com/opd-ai/unpeople"
)

func main() {
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

	fmt.Printf("%-45s  %7s  %8s  %5s\n", "Character", "Vertices", "Triangles", "Key")
	fmt.Printf("%-45s  %7s  %8s  %5s\n",
		"─────────────────────────────────────────────",
		"───────", "─────────", "───")

	for _, ex := range examples {
		m, err := g.Generate(ex.params)
		if err != nil {
			log.Printf("ERROR %s: %v\n", ex.label, err)
			continue
		}
		fmt.Printf("%-45s  %7d  %8d  %s\n",
			ex.label,
			len(m.Vertices),
			len(m.Indices)/3,
			m.Key,
		)
	}
}
