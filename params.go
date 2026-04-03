// Package unpeople provides deterministic procedural generation of humanoid
// character meshes. Given a seed and a set of descriptive parameters, the
// Generator always produces an identical mesh, making it suitable for use in
// open-world games where characters must be reproducible from a saved seed.
//
// The output Mesh type is layout-compatible with the Kaiju engine's
// rendering.Vertex / rendering.Mesh structures so that it can be passed
// directly to kaiju's rendering pipeline.
package unpeople

import "errors"

// ─── Species ─────────────────────────────────────────────────────────────────

// Species identifies the humanoid species of the generated character.
type Species int

const (
	SpeciesHuman Species = iota
	SpeciesElf
	SpeciesDwarf
	SpeciesGnome
	SpeciesHalfling
	SpeciesGoblin
	SpeciesKobold
	SpeciesOrc
	SpeciesTroll
	SpeciesOgre
)

// ─── Height ──────────────────────────────────────────────────────────────────

// Height describes the overall stature of the character.
type Height int

const (
	HeightGiant Height = iota
	HeightTall
	HeightMedium
	HeightShort
	HeightTiny
)

// ─── Build ───────────────────────────────────────────────────────────────────

// Build describes the muscular / body-mass profile.
type Build int

const (
	BuildMuscular Build = iota
	BuildAthletic
	BuildAverage
	BuildLean
	BuildStocky
	BuildFragile // "Frail" in the spec
)

// ─── Proportions ─────────────────────────────────────────────────────────────

// Proportions describes the overall stylistic exaggeration of body ratios.
type Proportions int

const (
	ProportionsHeroic Proportions = iota
	ProportionsRealistic
	ProportionsStylized
	ProportionsCaricature
)

// ─── Phenotype ───────────────────────────────────────────────────────────────

// Phenotype describes the sexual dimorphism of the silhouette.
type Phenotype int

const (
	PhenotypeMasculine Phenotype = iota
	PhenotypeAndrogynous
	PhenotypeFeminine
)

// ─── Age ─────────────────────────────────────────────────────────────────────

// Age describes the developmental / ageing stage of the character.
type Age int

const (
	AgeDecrepit Age = iota
	AgeElderly
	AgeOld
	AgeAdult
	AgeYouth
	AgeTeen
	AgeChild
	AgeToddler
)

// ─── Posture ─────────────────────────────────────────────────────────────────

// Posture describes the resting stance of the character.
type Posture int

const (
	PostureUpright Posture = iota
	PostureSlouched
	PostureHunched
	PostureRigid
)

// ─── FaceShape ───────────────────────────────────────────────────────────────

// FaceShape describes the overall silhouette of the head/face.
type FaceShape int

const (
	FaceShapeOval FaceShape = iota
	FaceShapeRound
	FaceShapeSquare
	FaceShapeHeart
	FaceShapeDiamond
	FaceShapeOblong
)

// ─── Jaw ─────────────────────────────────────────────────────────────────────

// Jaw describes the prominence and shape of the jawline.
type Jaw int

const (
	JawProminent Jaw = iota
	JawAverage
	JawSubtle
	JawAngular
	JawRounded
)

// ─── Brow ────────────────────────────────────────────────────────────────────

// Brow describes the thickness and shape of the brow ridge.
type Brow int

const (
	BrowHeavy Brow = iota
	BrowNormal
	BrowLight
	BrowArched
)

// ─── Ears ────────────────────────────────────────────────────────────────────

// Ears describes the size and shape of the ears.
type Ears int

const (
	EarsSmall Ears = iota
	EarsMedium
	EarsLarge
	EarsPointed
	EarsRounded
)

// ─── ShoulderWidth ───────────────────────────────────────────────────────────

// ShoulderWidth describes the breadth of the shoulders.
type ShoulderWidth int

const (
	ShoulderWidthBroad ShoulderWidth = iota
	ShoulderWidthAverage
	ShoulderWidthNarrow
)

// ─── HipWidth ────────────────────────────────────────────────────────────────

// HipWidth describes the breadth of the hips / pelvis.
type HipWidth int

const (
	HipWidthWide HipWidth = iota
	HipWidthAverage
	HipWidthNarrow
)

// ─── LimbLength ──────────────────────────────────────────────────────────────

// LimbLength describes the proportional length of the arms and legs.
type LimbLength int

const (
	LimbLengthLong LimbLength = iota
	LimbLengthProportional
	LimbLengthShort
)

// ─── NeckLength ──────────────────────────────────────────────────────────────

// NeckLength describes the length and girth of the neck.
type NeckLength int

const (
	NeckLengthLong NeckLength = iota
	NeckLengthMedium
	NeckLengthShort
	NeckLengthThick
)

// ─── HandSize ────────────────────────────────────────────────────────────────

// HandSize describes the overall size of the hands.
type HandSize int

const (
	HandSizeLarge HandSize = iota
	HandSizeMedium
	HandSizeSmall
)

// ─── FingerLength ────────────────────────────────────────────────────────────

// FingerLength describes the proportional length of the fingers.
type FingerLength int

const (
	FingerLengthLong FingerLength = iota
	FingerLengthAverage
	FingerLengthShort
)

// ─── FootSize ────────────────────────────────────────────────────────────────

// FootSize describes the overall size of the feet.
type FootSize int

const (
	FootSizeLarge FootSize = iota
	FootSizeMedium
	FootSizeSmall
)

// ─── Params ──────────────────────────────────────────────────────────────────

// Params is the complete set of inputs to the Generator. Every field has a
// well-defined set of valid enum values; call Validate before generating.
type Params struct {
	// Seed controls the PRNG so that identical Params always yield the same mesh.
	Seed int64

	// Body structure
	Species     Species
	Height      Height
	Build       Build
	Proportions Proportions

	// Physical traits
	Phenotype Phenotype
	Age       Age
	Posture   Posture

	// Facial features
	FaceShape FaceShape
	Jaw       Jaw
	Brow      Brow
	Ears      Ears

	// Body details
	ShoulderWidth ShoulderWidth
	HipWidth      HipWidth
	LimbLength    LimbLength
	NeckLength    NeckLength

	// Hands & feet
	HandSize     HandSize
	FingerLength FingerLength
	FootSize     FootSize

	// Optional mesh components
	HasHairSlot bool // If true, generate skull cap placeholder for hair attachment
}

// DefaultParams returns a Params representing a generic Adult Human with
// average proportions and no extremes in any trait.
func DefaultParams() Params {
	return Params{
		Seed:          0,
		Species:       SpeciesHuman,
		Height:        HeightMedium,
		Build:         BuildAverage,
		Proportions:   ProportionsRealistic,
		Phenotype:     PhenotypeAndrogynous,
		Age:           AgeAdult,
		Posture:       PostureUpright,
		FaceShape:     FaceShapeOval,
		Jaw:           JawAverage,
		Brow:          BrowNormal,
		Ears:          EarsMedium,
		ShoulderWidth: ShoulderWidthAverage,
		HipWidth:      HipWidthAverage,
		LimbLength:    LimbLengthProportional,
		NeckLength:    NeckLengthMedium,
		HandSize:      HandSizeMedium,
		FingerLength:  FingerLengthAverage,
		FootSize:      FootSizeMedium,
		HasHairSlot:   true,
	}
}

// Validate returns an error if any parameter is outside its defined range.
func (p *Params) Validate() error {
	// Table-driven validation to reduce cyclomatic complexity
	checks := []struct {
		val  int
		min  int
		max  int
		name string
	}{
		{int(p.Species), int(SpeciesHuman), int(SpeciesOgre), "Species"},
		{int(p.Height), int(HeightGiant), int(HeightTiny), "Height"},
		{int(p.Build), int(BuildMuscular), int(BuildFragile), "Build"},
		{int(p.Proportions), int(ProportionsHeroic), int(ProportionsCaricature), "Proportions"},
		{int(p.Phenotype), int(PhenotypeMasculine), int(PhenotypeFeminine), "Phenotype"},
		{int(p.Age), int(AgeDecrepit), int(AgeToddler), "Age"},
		{int(p.Posture), int(PostureUpright), int(PostureRigid), "Posture"},
		{int(p.FaceShape), int(FaceShapeOval), int(FaceShapeOblong), "FaceShape"},
		{int(p.Jaw), int(JawProminent), int(JawRounded), "Jaw"},
		{int(p.Brow), int(BrowHeavy), int(BrowArched), "Brow"},
		{int(p.Ears), int(EarsSmall), int(EarsRounded), "Ears"},
		{int(p.ShoulderWidth), int(ShoulderWidthBroad), int(ShoulderWidthNarrow), "ShoulderWidth"},
		{int(p.HipWidth), int(HipWidthWide), int(HipWidthNarrow), "HipWidth"},
		{int(p.LimbLength), int(LimbLengthLong), int(LimbLengthShort), "LimbLength"},
		{int(p.NeckLength), int(NeckLengthLong), int(NeckLengthThick), "NeckLength"},
		{int(p.HandSize), int(HandSizeLarge), int(HandSizeSmall), "HandSize"},
		{int(p.FingerLength), int(FingerLengthLong), int(FingerLengthShort), "FingerLength"},
		{int(p.FootSize), int(FootSizeLarge), int(FootSizeSmall), "FootSize"},
	}
	for _, c := range checks {
		if c.val < c.min || c.val > c.max {
			return errors.New("invalid " + c.name + " value")
		}
	}
	return nil
}
