package unpeople

// computeBodyLayout derives a fully-specified bodyLayout from Params.
// The seeded rng is used only for small stochastic details that the caller
// cannot observe individually (e.g. posture micro-offsets); every observable
// aspect of the layout is derived deterministically from the named parameters.
func computeBodyLayout(p *Params, rng *splitmix64) bodyLayout {
	l := defaultBodyLayout()

	applySpecies(&l, p.Species)
	applyHeight(&l, p.Height)
	applyBuild(&l, p.Build, p.Species)
	applyProportions(&l, p.Proportions)
	applyPhenotype(&l, p.Phenotype)
	applyAge(&l, p.Age, p.Species)
	applyShoulderWidth(&l, p.ShoulderWidth)
	applyHipWidth(&l, p.HipWidth)
	applyLimbLength(&l, p.LimbLength)
	applyNeckLength(&l, p.NeckLength)
	applyHandSize(&l, p.HandSize, p.FingerLength)
	applyFootSize(&l, p.FootSize)
	applyFacialFeatures(&l, p.FaceShape, p.Jaw, p.Brow, p.Ears)
	applyPosture(&l, p.Posture, p.Age, rng)

	return l
}

// ─── Species ─────────────────────────────────────────────────────────────────

// applyElfTraits applies slender, tall proportions for elves.
func applyElfTraits(l *bodyLayout) {
	scaleAll(l, 1.05)
	l.headRX *= 0.93
	l.headRZ *= 0.93
	l.hipsRX *= 0.92
	l.upperLegRadius *= 0.88
	l.lowerLegRadius *= 0.85
}

// applyDwarfTraits applies stocky, compact proportions for dwarves.
func applyDwarfTraits(l *bodyLayout) {
	scaleHeight(l, 0.77)
	l.chestRX *= 1.25
	l.hipsRX *= 1.15
	l.upperArmRadius *= 1.20
	l.upperLegRadius *= 1.20
	l.lowerLegRadius *= 1.15
}

// applySmallSpeciesTraits applies traits for gnome, halfling, goblin, kobold.
func applySmallSpeciesTraits(l *bodyLayout, s Species) {
	switch s {
	case SpeciesGnome:
		scaleHeight(l, 0.62)
		l.headRX *= 1.15
		l.headRY *= 1.15
		l.headRZ *= 1.15
	case SpeciesHalfling:
		scaleHeight(l, 0.68)
		l.headRX *= 1.05
		l.chestRX *= 1.05
		l.hipsRX *= 1.08
	case SpeciesGoblin:
		scaleHeight(l, 0.70)
		l.upperArmRadius *= 0.75
		l.forearmRadius *= 0.70
		l.headRX *= 1.10
	case SpeciesKobold:
		scaleHeight(l, 0.65)
		l.headRX *= 0.88
		l.headRY *= 0.90
	}
}

// applyLargeSpeciesTraits applies traits for orc, troll, ogre.
func applyLargeSpeciesTraits(l *bodyLayout, s Species) {
	switch s {
	case SpeciesOrc:
		l.chestRX *= 1.25
		l.chestRZ *= 1.20
		l.hipsRX *= 1.15
		l.upperArmRadius *= 1.30
		l.upperLegRadius *= 1.20
		l.neckRadius *= 1.30
	case SpeciesTroll:
		scaleAll(l, 1.35)
		l.upperArmRadius *= 1.40
		l.upperLegRadius *= 1.40
		l.lowerLegRadius *= 1.35
		l.neckRadius *= 1.50
	case SpeciesOgre:
		scaleAll(l, 1.50)
		l.chestRX *= 1.50
		l.hipsRX *= 1.40
		l.upperArmRadius *= 1.60
		l.upperLegRadius *= 1.50
		l.neckRadius *= 1.60
		l.headRX *= 1.20
		l.headRY *= 1.15
	}
}

func applySpecies(l *bodyLayout, s Species) {
	switch s {
	case SpeciesElf:
		applyElfTraits(l)
	case SpeciesDwarf:
		applyDwarfTraits(l)
	case SpeciesGnome, SpeciesHalfling, SpeciesGoblin, SpeciesKobold:
		applySmallSpeciesTraits(l, s)
	case SpeciesOrc, SpeciesTroll, SpeciesOgre:
		applyLargeSpeciesTraits(l, s)
	}
}

// ─── Height ──────────────────────────────────────────────────────────────────

func applyHeight(l *bodyLayout, h Height) {
	switch h {
	case HeightGiant:
		scaleAll(l, 2.00)
	case HeightTall:
		scaleAll(l, 1.15)
	case HeightMedium:
		// default
	case HeightShort:
		scaleAll(l, 0.85)
	case HeightTiny:
		scaleAll(l, 0.60)
	}
}

// ─── Build ───────────────────────────────────────────────────────────────────

// buildMultipliers holds species-specific adjustment factors for build effects.
type buildMultipliers struct {
	chest, limb float32
}

// buildInteractionTable maps (Build, Species) pairs to adjustment multipliers.
// These address awkward combinations like Orc+Fragile or Troll+Lean where
// base build multipliers produce unnatural proportions.
var buildInteractionTable = map[Build]map[Species]buildMultipliers{
	BuildFragile: {
		SpeciesOrc:   {1.15, 1.12}, // Orcs are naturally bulky
		SpeciesTroll: {1.20, 1.15},
		SpeciesOgre:  {1.25, 1.18},
		SpeciesDwarf: {1.10, 1.08}, // Dwarves are stocky
	},
	BuildLean: {
		SpeciesOrc:   {1.08, 1.06},
		SpeciesTroll: {1.12, 1.10},
		SpeciesOgre:  {1.15, 1.12},
	},
	BuildMuscular: {
		SpeciesOrc:   {0.90, 0.92}, // Already-bulky species need less expansion
		SpeciesTroll: {0.90, 0.92},
		SpeciesOgre:  {0.90, 0.92},
	},
}

// speciesBuildInteraction returns species-aware multipliers for build effects.
// chestMult adjusts torso reduction/expansion, limbMult adjusts limb reduction.
func speciesBuildInteraction(s Species, b Build) (chestMult, limbMult float32) {
	if speciesMap, ok := buildInteractionTable[b]; ok {
		if mults, ok := speciesMap[s]; ok {
			return mults.chest, mults.limb
		}
	}
	return 1.0, 1.0
}

func applyBuild(l *bodyLayout, b Build, s Species) {
	// Get species-aware adjustment multipliers
	chestMult, limbMult := speciesBuildInteraction(s, b)

	switch b {
	case BuildMuscular:
		l.chestRX *= 1.28 * chestMult
		l.chestRZ *= 1.20 * chestMult
		l.abdomenRX *= 1.10 * chestMult
		l.upperArmRadius *= 1.30 * limbMult
		l.forearmRadius *= 1.20 * limbMult
		l.upperLegRadius *= 1.25 * limbMult
		l.lowerLegRadius *= 1.20 * limbMult
		l.neckRadius *= 1.15 * chestMult
	case BuildAthletic:
		l.chestRX *= 1.12
		l.upperArmRadius *= 1.12
		l.upperLegRadius *= 1.10
	case BuildAverage:
		// default
	case BuildLean:
		l.chestRX *= 0.90 * chestMult
		l.abdomenRX *= 0.88 * chestMult
		l.hipsRX *= 0.92 * chestMult
		l.upperArmRadius *= 0.88 * limbMult
		l.upperLegRadius *= 0.88 * limbMult
		l.lowerLegRadius *= 0.90 * limbMult
	case BuildStocky:
		l.chestRX *= 1.15
		l.abdomenRX *= 1.15
		l.hipsRX *= 1.18
		l.upperLegRadius *= 1.15
		l.lowerLegRadius *= 1.10
	case BuildFragile:
		l.chestRX *= 0.80 * chestMult
		l.chestRZ *= 0.85 * chestMult
		l.abdomenRX *= 0.82 * chestMult
		l.hipsRX *= 0.85 * chestMult
		l.upperArmRadius *= 0.75 * limbMult
		l.forearmRadius *= 0.72 * limbMult
		l.upperLegRadius *= 0.78 * limbMult
		l.lowerLegRadius *= 0.75 * limbMult
		l.neckRadius *= 0.85 * chestMult
	}
}

// ─── Proportions ─────────────────────────────────────────────────────────────

func applyProportions(l *bodyLayout, p Proportions) {
	switch p {
	case ProportionsHeroic:
		// Wide shoulders, long legs, narrow hips
		l.chestRX *= 1.15
		l.hipsRX *= 0.92
		scaleLimbs(l, 1.08) // Elongate legs for heroic proportions
	case ProportionsRealistic:
		// default
	case ProportionsStylized:
		l.headRX *= 1.08
		l.headRY *= 1.10
		l.headRZ *= 1.08
	case ProportionsCaricature:
		l.headRX *= 1.25
		l.headRY *= 1.30
		l.headRZ *= 1.25
		l.chestRX *= 0.88
		l.upperArmRadius *= 0.88
		// Reduce hand/foot size for caricature effect
		l.handHW *= 0.85
		l.handHH *= 0.85
		l.handHD *= 0.85
		l.footHW *= 0.85
		l.footHD *= 0.85
	}
}

// ─── Phenotype ───────────────────────────────────────────────────────────────

func applyPhenotype(l *bodyLayout, p Phenotype) {
	switch p {
	case PhenotypeMasculine:
		l.chestRX *= 1.12
		l.chestRZ *= 1.08
		l.hipsRX *= 0.92
		l.neckRadius *= 1.10
		l.upperArmRadius *= 1.08
	case PhenotypeAndrogynous:
		// default
	case PhenotypeFeminine:
		l.chestRX *= 0.95
		l.hipsRX *= 1.12
		l.abdomenRX *= 0.88
		l.neckRadius *= 0.88
		l.upperArmRadius *= 0.90
		l.headRX *= 0.97
		l.headRY *= 0.97
	}
}

// ─── Age ─────────────────────────────────────────────────────────────────────

// childHeadMultiplier returns species-specific head scaling for child proportions.
// Smaller species have proportionally larger heads as children (neoteny).
func childHeadMultiplier(s Species) float32 {
	switch s {
	case SpeciesGnome, SpeciesHalfling:
		return 1.08
	case SpeciesKobold, SpeciesGoblin:
		return 1.06
	case SpeciesOgre, SpeciesTroll:
		return 0.95 // Large species have proportionally smaller child heads
	default:
		return 1.0
	}
}

// applyElderlyAge applies body modifications for elderly age stages.
func applyElderlyAge(l *bodyLayout, a Age) {
	switch a {
	case AgeDecrepit:
		scaleAll(l, 0.94)
		l.chestRX *= 0.85
		l.upperArmRadius *= 0.78
		l.upperLegRadius *= 0.80
	case AgeElderly:
		l.chestRX *= 0.90
		l.upperArmRadius *= 0.85
		l.upperLegRadius *= 0.85
	case AgeOld:
		l.chestRX *= 0.95
	}
}

// applyYoungAge applies body modifications for young age stages.
func applyYoungAge(l *bodyLayout, a Age, childHeadMult float32) {
	switch a {
	case AgeYouth:
		scaleAll(l, 0.95)
		l.headRX *= 1.03
		l.headRY *= 1.03
	case AgeTeen:
		scaleAll(l, 0.88)
		l.headRX *= 1.05
		l.headRY *= 1.07
	case AgeChild:
		scaleAll(l, 0.70)
		l.headRX *= 1.15 * childHeadMult
		l.headRY *= 1.18 * childHeadMult
		l.headRZ *= 1.15 * childHeadMult
	case AgeToddler:
		scaleAll(l, 0.45)
		l.headRX *= 1.35 * childHeadMult
		l.headRY *= 1.40 * childHeadMult
		l.headRZ *= 1.35 * childHeadMult
	}
}

// applyAge modifies the body layout based on the character's age stage.
// Species-specific head scaling is applied for AgeChild and AgeToddler.
func applyAge(l *bodyLayout, a Age, s Species) {
	switch a {
	case AgeDecrepit, AgeElderly, AgeOld:
		applyElderlyAge(l, a)
	case AgeAdult:
		// default - no modifications
	case AgeYouth, AgeTeen, AgeChild, AgeToddler:
		applyYoungAge(l, a, childHeadMultiplier(s))
	}
}

// ─── Shoulder Width ──────────────────────────────────────────────────────────

func applyShoulderWidth(l *bodyLayout, sw ShoulderWidth) {
	var xDelta float32
	switch sw {
	case ShoulderWidthBroad:
		l.chestRX *= 1.15
		xDelta = 0.03
	case ShoulderWidthAverage:
		// default
	case ShoulderWidthNarrow:
		l.chestRX *= 0.88
		xDelta = -0.025
	}
	if xDelta != 0 {
		shiftArmPositionsX(l, xDelta)
	}
}

// ─── Hip Width ───────────────────────────────────────────────────────────────

func applyHipWidth(l *bodyLayout, hw HipWidth) {
	var xDelta float32
	switch hw {
	case HipWidthWide:
		l.hipsRX *= 1.18
		xDelta = 0.020
	case HipWidthAverage:
		// default
	case HipWidthNarrow:
		l.hipsRX *= 0.85
		xDelta = -0.015
	}
	if xDelta != 0 {
		shiftLegPositionsX(l, xDelta)
	}
}

// ─── Limb Length ─────────────────────────────────────────────────────────────

func applyLimbLength(l *bodyLayout, ll LimbLength) {
	var scale float32
	switch ll {
	case LimbLengthLong:
		scale = 1.12
	case LimbLengthProportional:
		return
	case LimbLengthShort:
		scale = 0.85
	}
	scaleLimbs(l, scale)
}

// ─── Neck Length ─────────────────────────────────────────────────────────────

func applyNeckLength(l *bodyLayout, nl NeckLength) {
	switch nl {
	case NeckLengthLong:
		l.neckTop[1] += 0.04
		l.headCenter[1] += 0.04
	case NeckLengthMedium:
		// default
	case NeckLengthShort:
		l.neckTop[1] -= 0.03
		l.headCenter[1] -= 0.03
	case NeckLengthThick:
		l.neckRadius *= 1.35
	}
}

// ─── Hand Size ───────────────────────────────────────────────────────────────

func applyHandSize(l *bodyLayout, hs HandSize, fl FingerLength) {
	switch hs {
	case HandSizeLarge:
		l.handHW *= 1.20
		l.handHH *= 1.20
		l.handHD *= 1.15
		l.fingerRadius *= 1.15
	case HandSizeMedium:
		// default
	case HandSizeSmall:
		l.handHW *= 0.82
		l.handHH *= 0.82
		l.handHD *= 0.85
		l.fingerRadius *= 0.85
	}

	// Set finger length multiplier based on FingerLength param
	switch fl {
	case FingerLengthLong:
		l.handHH *= 1.15
		l.fingerLengthMult = 1.20
	case FingerLengthAverage:
		l.fingerLengthMult = 1.0
	case FingerLengthShort:
		l.handHH *= 0.85
		l.fingerLengthMult = 0.80
	}

	// After any change to handHH, recompute the Y centre so that the top of
	// the hand box stays flush with the wrist (forearm bottom attachment).
	l.handCenterL[1] = l.forearmBottomL[1] - l.handHH
	l.handCenterR[1] = l.forearmBottomR[1] - l.handHH
}

// ─── Foot Size ───────────────────────────────────────────────────────────────

func applyFootSize(l *bodyLayout, fs FootSize) {
	switch fs {
	case FootSizeLarge:
		l.footHW *= 1.20
		l.footHD *= 1.22
	case FootSizeMedium:
		// default
	case FootSizeSmall:
		l.footHW *= 0.82
		l.footHD *= 0.82
	}
}

// ─── Facial Features ─────────────────────────────────────────────────────────

// applyFaceShape modifies head geometry based on face shape.
func applyFaceShape(l *bodyLayout, fs FaceShape) {
	switch fs {
	case FaceShapeRound:
		l.headRX *= 1.08
		l.headRZ *= 1.08
		l.headRY *= 0.97
	case FaceShapeSquare:
		l.headRX *= 1.05
		l.headRZ *= 0.95
	case FaceShapeHeart:
		l.headRX *= 1.03
	case FaceShapeDiamond:
		l.headRX *= 0.97
		l.headRY *= 1.05
	case FaceShapeOblong:
		l.headRY *= 1.12
		l.headRX *= 0.92
	}
}

// applyJaw modifies head geometry based on jaw type.
func applyJaw(l *bodyLayout, j Jaw) {
	switch j {
	case JawProminent:
		l.headCenter[1] -= 0.005
		l.headRY *= 1.04
	case JawAngular:
		l.headRX *= 1.03
	case JawRounded:
		l.headRY *= 0.97
	}
}

// applyBrow modifies head geometry based on brow type.
func applyBrow(l *bodyLayout, br Brow) {
	if br == BrowHeavy {
		l.headRY *= 1.02
	}
}

// applyEarScale modifies ear scale based on ear type.
func applyEarScale(l *bodyLayout, e Ears) {
	switch e {
	case EarsLarge:
		l.earScale *= 1.30
	case EarsPointed:
		l.earScale *= 1.15
	case EarsSmall:
		l.earScale *= 0.70
	}
}

// updateEarAttachments repositions ear attachment points based on head geometry.
func updateEarAttachments(l *bodyLayout) {
	l.earAttachL[0] = -l.headRX
	l.earAttachL[1] = l.headCenter[1] + l.headRY*0.15
	l.earAttachL[2] = l.headCenter[2]
	l.earAttachR[0] = l.headRX
	l.earAttachR[1] = l.headCenter[1] + l.headRY*0.15
	l.earAttachR[2] = l.headCenter[2]
}

// applyFacialFeatures modifies the body layout based on facial feature parameters.
func applyFacialFeatures(l *bodyLayout, fs FaceShape, j Jaw, br Brow, e Ears) {
	applyFaceShape(l, fs)
	applyJaw(l, j)
	applyBrow(l, br)
	applyEarScale(l, e)
	updateEarAttachments(l)
}

// ─── Posture ─────────────────────────────────────────────────────────────────

// applyPosture adjusts the body layout based on stance. For elderly characters
// (AgeDecrepit or AgeElderly) with PostureUpright, a subtle automatic slouch
// is applied to reflect age-related posture changes.
func applyPosture(l *bodyLayout, p Posture, a Age, rng *splitmix64) {
	jitter := (rng.Float32() - 0.5) * 0.003

	// Auto-adjust posture for elderly characters when upright is requested
	effectivePosture := p
	if p == PostureUpright {
		switch a {
		case AgeDecrepit:
			effectivePosture = PostureSlouched // Automatic mild slouch
		case AgeElderly:
			// Apply a subtle forward lean (less than full slouch)
			shiftUpperBodyForward(l, 0.015+jitter)
			return
		}
	}

	switch effectivePosture {
	case PostureUpright, PostureRigid:
		// default (rigid = slightly stiffer but no visible geometry change)
	case PostureSlouched:
		shiftUpperBodyForward(l, 0.030+jitter)
	case PostureHunched:
		shiftUpperBodyForward(l, 0.065+jitter)
	}
}

// shiftUpperBodyForward tilts the torso/neck/head/shoulders forward by dz.
func shiftUpperBodyForward(l *bodyLayout, dz float32) {
	l.chestBottom[2] += dz
	l.chestTop[2] += dz
	l.abdomenTop[2] += dz * 0.50
	l.neckBottom[2] += dz
	l.neckTop[2] += dz
	l.headCenter[2] += dz
	l.upperArmTopL[2] += dz
	l.upperArmTopR[2] += dz
	l.upperArmBottomL[2] += dz * 0.50
	l.upperArmBottomR[2] += dz * 0.50
}

// ─── Uniform scale helpers ────────────────────────────────────────────────────

// allPositionFields returns pointers to all Vec3 position fields in the layout.
// Used by scaleAll and scaleHeight to avoid duplicating field lists.
func allPositionFields(l *bodyLayout) []*Vec3 {
	return []*Vec3{
		&l.headCenter,
		&l.neckBottom, &l.neckTop,
		&l.chestBottom, &l.chestTop,
		&l.abdomenBottom, &l.abdomenTop,
		&l.hipsBottom, &l.hipsTop,
		&l.upperArmTopL, &l.upperArmBottomL, &l.upperArmTopR, &l.upperArmBottomR,
		&l.forearmTopL, &l.forearmBottomL, &l.forearmTopR, &l.forearmBottomR,
		&l.handCenterL, &l.handCenterR,
		&l.upperLegTopL, &l.upperLegBottomL, &l.upperLegTopR, &l.upperLegBottomR,
		&l.lowerLegTopL, &l.lowerLegBottomL, &l.lowerLegTopR, &l.lowerLegBottomR,
		&l.footCenterL, &l.footCenterR,
		&l.earAttachL, &l.earAttachR,
	}
}

// allUniformRadii returns pointers to all radius/dimension fields that should
// scale uniformly (XYZ). Excludes height-only fields like handHH, footHH, headRY.
func allUniformRadii(l *bodyLayout) []*float32 {
	return []*float32{
		&l.headRX, &l.headRY, &l.headRZ,
		&l.neckRadius,
		&l.chestRX, &l.chestRZ,
		&l.abdomenRX, &l.abdomenRZ,
		&l.hipsRX, &l.hipsRZ,
		&l.upperArmRadius, &l.forearmRadius,
		&l.handHW, &l.handHH, &l.handHD,
		&l.upperLegRadius, &l.lowerLegRadius,
		&l.footHW, &l.footHH, &l.footHD,
		&l.earScale,
	}
}

// heightOnlyRadii returns pointers to radius/dimension fields that scale only
// in the vertical (Y) direction during scaleHeight operations.
func heightOnlyRadii(l *bodyLayout) []*float32 {
	return []*float32{&l.headRY, &l.handHH, &l.footHH}
}

// scaleAll uniformly scales all positions and radii around the world origin
// (feet remain near Y≈0).
func scaleAll(l *bodyLayout, s float32) {
	l.totalHeight *= s
	for _, v := range allPositionFields(l) {
		scaleV3(v, s)
	}
	for _, r := range allUniformRadii(l) {
		*r *= s
	}
}

// scaleHeight scales only the Y component of all position vectors and vertical
// radii (used by species that are shorter but not narrower).
func scaleHeight(l *bodyLayout, s float32) {
	l.totalHeight *= s
	for _, v := range allPositionFields(l) {
		scaleV3Y(v, s)
	}
	for _, r := range heightOnlyRadii(l) {
		*r *= s
	}
}

// scaleLimbs rescales limb segment lengths proportionally (arms and legs)
// while keeping attachment points (shoulders, hip sockets) fixed.
func scaleLimbs(l *bodyLayout, s float32) {
	// Scale arms: shoulder → elbow → wrist → hand
	scaleArmChain(
		&l.upperArmTopL, &l.upperArmBottomL, &l.forearmTopL, &l.forearmBottomL, &l.handCenterL,
		&l.upperArmTopR, &l.upperArmBottomR, &l.forearmTopR, &l.forearmBottomR, &l.handCenterR,
		s, l.handHH, true,
	)

	// Scale legs: hip → knee → ankle → foot
	scaleLegChain(
		&l.upperLegTopL, &l.upperLegBottomL, &l.lowerLegTopL, &l.lowerLegBottomL, &l.footCenterL,
		&l.upperLegTopR, &l.upperLegBottomR, &l.lowerLegTopR, &l.lowerLegBottomR, &l.footCenterR,
		s, l.footHH,
	)
}

// scaleArmChain scales a two-segment arm chain (upper arm + forearm) while
// keeping shoulder positions fixed. Hands are repositioned below the wrist.
func scaleArmChain(
	upperTopL, upperBottomL, lowerTopL, lowerBottomL, endCenterL *Vec3,
	upperTopR, upperBottomR, lowerTopR, lowerBottomR, endCenterR *Vec3,
	s, endOffset float32, endBelow bool,
) {
	// Left arm
	upperVecL := vec3Sub(*upperBottomL, *upperTopL)
	lowerVecL := vec3Sub(*lowerBottomL, *lowerTopL)
	*upperBottomL = vec3Add(*upperTopL, vec3Scale(upperVecL, s))
	*lowerTopL = *upperBottomL
	*lowerBottomL = vec3Add(*lowerTopL, vec3Scale(lowerVecL, s))
	if endBelow {
		*endCenterL = Vec3{(*lowerBottomL)[0], (*lowerBottomL)[1] - endOffset, (*lowerBottomL)[2]}
	} else {
		*endCenterL = Vec3{(*lowerBottomL)[0], (*lowerBottomL)[1] + endOffset, (*lowerBottomL)[2]}
	}

	// Right arm
	upperVecR := vec3Sub(*upperBottomR, *upperTopR)
	lowerVecR := vec3Sub(*lowerBottomR, *lowerTopR)
	*upperBottomR = vec3Add(*upperTopR, vec3Scale(upperVecR, s))
	*lowerTopR = *upperBottomR
	*lowerBottomR = vec3Add(*lowerTopR, vec3Scale(lowerVecR, s))
	if endBelow {
		*endCenterR = Vec3{(*lowerBottomR)[0], (*lowerBottomR)[1] - endOffset, (*lowerBottomR)[2]}
	} else {
		*endCenterR = Vec3{(*lowerBottomR)[0], (*lowerBottomR)[1] + endOffset, (*lowerBottomR)[2]}
	}
}

// scaleLegChain scales a two-segment leg chain (upper leg + lower leg) while
// keeping hip positions fixed. Feet are repositioned above the ankle.
func scaleLegChain(
	upperTopL, upperBottomL, lowerTopL, lowerBottomL, footCenterL *Vec3,
	upperTopR, upperBottomR, lowerTopR, lowerBottomR, footCenterR *Vec3,
	s, footHH float32,
) {
	// Left leg
	upperVecL := vec3Sub(*upperBottomL, *upperTopL)
	lowerVecL := vec3Sub(*lowerBottomL, *lowerTopL)
	*upperBottomL = vec3Add(*upperTopL, vec3Scale(upperVecL, s))
	*lowerTopL = *upperBottomL
	*lowerBottomL = vec3Add(*lowerTopL, vec3Scale(lowerVecL, s))
	*footCenterL = Vec3{(*lowerBottomL)[0], (*lowerBottomL)[1] + footHH, (*footCenterL)[2]}

	// Right leg
	upperVecR := vec3Sub(*upperBottomR, *upperTopR)
	lowerVecR := vec3Sub(*lowerBottomR, *lowerTopR)
	*upperBottomR = vec3Add(*upperTopR, vec3Scale(upperVecR, s))
	*lowerTopR = *upperBottomR
	*lowerBottomR = vec3Add(*lowerTopR, vec3Scale(lowerVecR, s))
	*footCenterR = Vec3{(*lowerBottomR)[0], (*lowerBottomR)[1] + footHH, (*footCenterR)[2]}
}

// ─── Vec3 helpers ─────────────────────────────────────────────────────────────

func scaleV3(v *Vec3, s float32) {
	v[0] *= s
	v[1] *= s
	v[2] *= s
}

func scaleV3Y(v *Vec3, s float32) {
	v[1] *= s
}

// ─── Limb Position Shift Helpers ──────────────────────────────────────────────

// shiftLimbPairX applies a lateral shift to a left/right pair of joint positions.
// Left position shifts by +delta, right by -delta.
func shiftLimbPairX(left, right *Vec3, delta float32) {
	left[0] += delta
	right[0] -= delta
}

// shiftArmPositionsX laterally shifts all arm joint positions by the given delta.
// Left arm positions move by +delta, right arm positions move by -delta.
func shiftArmPositionsX(l *bodyLayout, delta float32) {
	shiftLimbPairX(&l.upperArmTopL, &l.upperArmTopR, delta)
	shiftLimbPairX(&l.upperArmBottomL, &l.upperArmBottomR, delta)
	shiftLimbPairX(&l.forearmTopL, &l.forearmTopR, delta)
	shiftLimbPairX(&l.forearmBottomL, &l.forearmBottomR, delta)
	shiftLimbPairX(&l.handCenterL, &l.handCenterR, delta)
}

// shiftLegPositionsX laterally shifts all leg joint positions by the given delta.
// Left leg positions move by +delta, right leg positions move by -delta.
func shiftLegPositionsX(l *bodyLayout, delta float32) {
	shiftLimbPairX(&l.upperLegTopL, &l.upperLegTopR, delta)
	shiftLimbPairX(&l.upperLegBottomL, &l.upperLegBottomR, delta)
	shiftLimbPairX(&l.lowerLegTopL, &l.lowerLegTopR, delta)
	shiftLimbPairX(&l.lowerLegBottomL, &l.lowerLegBottomR, delta)
	shiftLimbPairX(&l.footCenterL, &l.footCenterR, delta)
}
