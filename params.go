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
	SpeciesHuman    Species = iota
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
	HeightGiant  Height = iota
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
	ProportionsHeroic     Proportions = iota
	ProportionsRealistic
	ProportionsStylized
	ProportionsCaricature
)

// ─── Phenotype ───────────────────────────────────────────────────────────────

// Phenotype describes the sexual dimorphism of the silhouette.
type Phenotype int

const (
	PhenotypeMasculine   Phenotype = iota
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
	PostureUpright  Posture = iota
	PostureSlouched
	PostureHunched
	PostureRigid
)

// ─── FaceShape ───────────────────────────────────────────────────────────────

// FaceShape describes the overall silhouette of the head/face.
type FaceShape int

const (
	FaceShapeOval    FaceShape = iota
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
	BrowHeavy  Brow = iota
	BrowNormal
	BrowLight
	BrowArched
)

// ─── Ears ────────────────────────────────────────────────────────────────────

// Ears describes the size and shape of the ears.
type Ears int

const (
	EarsSmall   Ears = iota
	EarsMedium
	EarsLarge
	EarsPointed
	EarsRounded
)

// ─── ShoulderWidth ───────────────────────────────────────────────────────────

// ShoulderWidth describes the breadth of the shoulders.
type ShoulderWidth int

const (
	ShoulderWidthBroad   ShoulderWidth = iota
	ShoulderWidthAverage
	ShoulderWidthNarrow
)

// ─── HipWidth ────────────────────────────────────────────────────────────────

// HipWidth describes the breadth of the hips / pelvis.
type HipWidth int

const (
	HipWidthWide    HipWidth = iota
	HipWidthAverage
	HipWidthNarrow
)

// ─── LimbLength ──────────────────────────────────────────────────────────────

// LimbLength describes the proportional length of the arms and legs.
type LimbLength int

const (
	LimbLengthLong         LimbLength = iota
	LimbLengthProportional
	LimbLengthShort
)

// ─── NeckLength ──────────────────────────────────────────────────────────────

// NeckLength describes the length and girth of the neck.
type NeckLength int

const (
	NeckLengthLong   NeckLength = iota
	NeckLengthMedium
	NeckLengthShort
	NeckLengthThick
)

// ─── HandSize ────────────────────────────────────────────────────────────────

// HandSize describes the overall size of the hands.
type HandSize int

const (
	HandSizeLarge  HandSize = iota
	HandSizeMedium
	HandSizeSmall
)

// ─── FingerLength ────────────────────────────────────────────────────────────

// FingerLength describes the proportional length of the fingers.
type FingerLength int

const (
	FingerLengthLong    FingerLength = iota
	FingerLengthAverage
	FingerLengthShort
)

// ─── FootSize ────────────────────────────────────────────────────────────────

// FootSize describes the overall size of the feet.
type FootSize int

const (
	FootSizeLarge  FootSize = iota
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
	}
}

// Validate returns an error if any parameter is outside its defined range.
func (p *Params) Validate() error {
	if p.Species < SpeciesHuman || p.Species > SpeciesOgre {
		return errors.New("invalid Species value")
	}
	if p.Height < HeightGiant || p.Height > HeightTiny {
		return errors.New("invalid Height value")
	}
	if p.Build < BuildMuscular || p.Build > BuildFragile {
		return errors.New("invalid Build value")
	}
	if p.Proportions < ProportionsHeroic || p.Proportions > ProportionsCaricature {
		return errors.New("invalid Proportions value")
	}
	if p.Phenotype < PhenotypeMasculine || p.Phenotype > PhenotypeFeminine {
		return errors.New("invalid Phenotype value")
	}
	if p.Age < AgeDecrepit || p.Age > AgeToddler {
		return errors.New("invalid Age value")
	}
	if p.Posture < PostureUpright || p.Posture > PostureRigid {
		return errors.New("invalid Posture value")
	}
	if p.FaceShape < FaceShapeOval || p.FaceShape > FaceShapeOblong {
		return errors.New("invalid FaceShape value")
	}
	if p.Jaw < JawProminent || p.Jaw > JawRounded {
		return errors.New("invalid Jaw value")
	}
	if p.Brow < BrowHeavy || p.Brow > BrowArched {
		return errors.New("invalid Brow value")
	}
	if p.Ears < EarsSmall || p.Ears > EarsRounded {
		return errors.New("invalid Ears value")
	}
	if p.ShoulderWidth < ShoulderWidthBroad || p.ShoulderWidth > ShoulderWidthNarrow {
		return errors.New("invalid ShoulderWidth value")
	}
	if p.HipWidth < HipWidthWide || p.HipWidth > HipWidthNarrow {
		return errors.New("invalid HipWidth value")
	}
	if p.LimbLength < LimbLengthLong || p.LimbLength > LimbLengthShort {
		return errors.New("invalid LimbLength value")
	}
	if p.NeckLength < NeckLengthLong || p.NeckLength > NeckLengthThick {
		return errors.New("invalid NeckLength value")
	}
	if p.HandSize < HandSizeLarge || p.HandSize > HandSizeSmall {
		return errors.New("invalid HandSize value")
	}
	if p.FingerLength < FingerLengthLong || p.FingerLength > FingerLengthShort {
		return errors.New("invalid FingerLength value")
	}
	if p.FootSize < FootSizeLarge || p.FootSize > FootSizeSmall {
		return errors.New("invalid FootSize value")
	}
	return nil
}
