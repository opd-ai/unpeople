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

// characterExample defines a test character configuration.
type characterExample struct {
	label  string
	params unpeople.Params
}

// makeElfExample creates a tall athletic elf configuration.
func makeElfExample() characterExample {
	p := unpeople.DefaultParams()
	p.Seed = 100
	p.Species = unpeople.SpeciesElf
	p.Height = unpeople.HeightTall
	p.Build = unpeople.BuildAthletic
	p.Phenotype = unpeople.PhenotypeFeminine
	p.Ears = unpeople.EarsPointed
	p.FaceShape = unpeople.FaceShapeOval
	p.LimbLength = unpeople.LimbLengthLong
	return characterExample{"Tall Athletic Elf – pointed ears", p}
}

// makeDwarfExample creates a stocky dwarf configuration.
func makeDwarfExample() characterExample {
	p := unpeople.DefaultParams()
	p.Seed = 200
	p.Species = unpeople.SpeciesDwarf
	p.Height = unpeople.HeightShort
	p.Build = unpeople.BuildStocky
	p.Phenotype = unpeople.PhenotypeMasculine
	p.ShoulderWidth = unpeople.ShoulderWidthBroad
	p.FaceShape = unpeople.FaceShapeSquare
	p.Jaw = unpeople.JawProminent
	return characterExample{"Stocky Dwarf – broad shoulders", p}
}

// makeTrollExample creates a giant muscular troll configuration.
func makeTrollExample() characterExample {
	p := unpeople.DefaultParams()
	p.Seed = 300
	p.Species = unpeople.SpeciesTroll
	p.Height = unpeople.HeightGiant
	p.Build = unpeople.BuildMuscular
	p.Proportions = unpeople.ProportionsHeroic
	p.NeckLength = unpeople.NeckLengthThick
	p.HandSize = unpeople.HandSizeLarge
	p.FootSize = unpeople.FootSizeLarge
	return characterExample{"Giant Muscular Troll", p}
}

// makeGnomeExample creates a toddler gnome configuration.
func makeGnomeExample() characterExample {
	p := unpeople.DefaultParams()
	p.Seed = 400
	p.Species = unpeople.SpeciesGnome
	p.Age = unpeople.AgeToddler
	p.Proportions = unpeople.ProportionsCaricature
	p.Height = unpeople.HeightTiny
	p.Build = unpeople.BuildFragile
	return characterExample{"Toddler Gnome – caricature", p}
}

// makeOrcExample creates an elderly frail orc configuration.
func makeOrcExample() characterExample {
	p := unpeople.DefaultParams()
	p.Seed = 500
	p.Species = unpeople.SpeciesOrc
	p.Age = unpeople.AgeElderly
	p.Build = unpeople.BuildFragile
	p.Posture = unpeople.PostureHunched
	p.LimbLength = unpeople.LimbLengthShort
	return characterExample{"Elderly Frail Orc – hunched", p}
}

// makeHalflingExample creates an athletic halfling configuration.
func makeHalflingExample() characterExample {
	p := unpeople.DefaultParams()
	p.Seed = 600
	p.Species = unpeople.SpeciesHalfling
	p.Build = unpeople.BuildAthletic
	p.Proportions = unpeople.ProportionsHeroic
	p.HipWidth = unpeople.HipWidthNarrow
	p.ShoulderWidth = unpeople.ShoulderWidthBroad
	return characterExample{"Athletic Halfling – heroic proportions", p}
}

// makeKoboldExample creates a lean feminine kobold configuration.
func makeKoboldExample() characterExample {
	p := unpeople.DefaultParams()
	p.Seed = 700
	p.Species = unpeople.SpeciesKobold
	p.Build = unpeople.BuildLean
	p.Phenotype = unpeople.PhenotypeFeminine
	p.Proportions = unpeople.ProportionsStylized
	p.NeckLength = unpeople.NeckLengthLong
	p.FingerLength = unpeople.FingerLengthLong
	return characterExample{"Lean Feminine Kobold – stylized", p}
}

// buildExamples returns all character examples to generate.
func buildExamples() []characterExample {
	return []characterExample{
		{"Adult Human (default)", unpeople.DefaultParams()},
		makeElfExample(),
		makeDwarfExample(),
		makeTrollExample(),
		makeGnomeExample(),
		makeOrcExample(),
		makeHalflingExample(),
		makeKoboldExample(),
	}
}

// printHeader outputs the table header.
func printHeader() {
	fmt.Printf("%-45s  %7s  %8s  %5s\n", "Character", "Vertices", "Triangles", "Key")
	fmt.Printf("%-45s  %7s  %8s  %5s\n",
		"─────────────────────────────────────────────",
		"───────", "─────────", "───")
}

// printResult outputs a single character's mesh statistics.
func printResult(label string, m *unpeople.Mesh) {
	fmt.Printf("%-45s  %7d  %8d  %s\n",
		label,
		len(m.Vertices),
		len(m.Indices)/3,
		m.Key,
	)
}

func main() {
	g := unpeople.NewGenerator()
	examples := buildExamples()

	printHeader()

	for _, ex := range examples {
		m, err := g.Generate(ex.params)
		if err != nil {
			log.Printf("ERROR %s: %v\n", ex.label, err)
			continue
		}
		printResult(ex.label, m)
	}
}
