package unpeople

// computeBodyLayout derives a fully-specified bodyLayout from Params.
// The seeded rng is used only for small stochastic details that the caller
// cannot observe individually (e.g. posture micro-offsets); every observable
// aspect of the layout is derived deterministically from the named parameters.
func computeBodyLayout(p *Params, rng *splitmix64) bodyLayout {
	l := defaultBodyLayout()

	applySpecies(&l, p.Species)
	applyHeight(&l, p.Height)
	applyBuild(&l, p.Build)
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

func applySpecies(l *bodyLayout, s Species) {
	switch s {
	case SpeciesElf:
		scaleAll(l, 1.05)
		l.headRX *= 0.93
		l.headRZ *= 0.93
		l.hipsRX *= 0.92
		l.upperLegRadius *= 0.88
		l.lowerLegRadius *= 0.85
	case SpeciesDwarf:
		scaleHeight(l, 0.77)
		l.chestRX *= 1.25
		l.hipsRX *= 1.15
		l.upperArmRadius *= 1.20
		l.upperLegRadius *= 1.20
		l.lowerLegRadius *= 1.15
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

func applyBuild(l *bodyLayout, b Build) {
	switch b {
	case BuildMuscular:
		l.chestRX *= 1.28
		l.chestRZ *= 1.20
		l.abdomenRX *= 1.10
		l.upperArmRadius *= 1.30
		l.forearmRadius *= 1.20
		l.upperLegRadius *= 1.25
		l.lowerLegRadius *= 1.20
		l.neckRadius *= 1.15
	case BuildAthletic:
		l.chestRX *= 1.12
		l.upperArmRadius *= 1.12
		l.upperLegRadius *= 1.10
	case BuildAverage:
		// default
	case BuildLean:
		l.chestRX *= 0.90
		l.abdomenRX *= 0.88
		l.hipsRX *= 0.92
		l.upperArmRadius *= 0.88
		l.upperLegRadius *= 0.88
		l.lowerLegRadius *= 0.90
	case BuildStocky:
		l.chestRX *= 1.15
		l.abdomenRX *= 1.15
		l.hipsRX *= 1.18
		l.upperLegRadius *= 1.15
		l.lowerLegRadius *= 1.10
	case BuildFragile:
		l.chestRX *= 0.80
		l.chestRZ *= 0.85
		l.abdomenRX *= 0.82
		l.hipsRX *= 0.85
		l.upperArmRadius *= 0.75
		l.forearmRadius *= 0.72
		l.upperLegRadius *= 0.78
		l.lowerLegRadius *= 0.75
		l.neckRadius *= 0.85
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

// applyAge modifies the body layout based on the character's age stage.
// Species-specific head scaling is applied for AgeChild and AgeToddler.
func applyAge(l *bodyLayout, a Age, s Species) {
	// Species-specific head scale multiplier for child proportions.
	// Smaller species (Gnome, Halfling, Kobold, Goblin) have proportionally
	// larger heads as children to match real-world neoteny patterns.
	childHeadMult := float32(1.0)
	switch s {
	case SpeciesGnome, SpeciesHalfling:
		childHeadMult = 1.08
	case SpeciesKobold, SpeciesGoblin:
		childHeadMult = 1.06
	case SpeciesOgre, SpeciesTroll:
		childHeadMult = 0.95 // Large species have proportionally smaller child heads
	}

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
	case AgeAdult:
		// default
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
	if xDelta == 0 {
		return
	}
	// Shift all arm / hand X positions
	l.upperArmTopL[0] += xDelta
	l.upperArmBottomL[0] += xDelta
	l.forearmTopL[0] += xDelta
	l.forearmBottomL[0] += xDelta
	l.handCenterL[0] += xDelta

	l.upperArmTopR[0] -= xDelta
	l.upperArmBottomR[0] -= xDelta
	l.forearmTopR[0] -= xDelta
	l.forearmBottomR[0] -= xDelta
	l.handCenterR[0] -= xDelta
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
	if xDelta == 0 {
		return
	}
	l.upperLegTopL[0] += xDelta
	l.upperLegBottomL[0] += xDelta
	l.lowerLegTopL[0] += xDelta
	l.lowerLegBottomL[0] += xDelta
	l.footCenterL[0] += xDelta

	l.upperLegTopR[0] -= xDelta
	l.upperLegBottomR[0] -= xDelta
	l.lowerLegTopR[0] -= xDelta
	l.lowerLegBottomR[0] -= xDelta
	l.footCenterR[0] -= xDelta
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
	case HandSizeMedium:
		// default
	case HandSizeSmall:
		l.handHW *= 0.82
		l.handHH *= 0.82
		l.handHD *= 0.85
	}
	switch fl {
	case FingerLengthLong:
		l.handHH *= 1.15
	case FingerLengthAverage:
		// default
	case FingerLengthShort:
		l.handHH *= 0.85
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

func applyFacialFeatures(l *bodyLayout, fs FaceShape, j Jaw, br Brow, e Ears) {
	switch fs {
	case FaceShapeOval:
		// default
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

	switch j {
	case JawProminent:
		l.headCenter[1] -= 0.005
		l.headRY *= 1.04
	case JawAngular:
		l.headRX *= 1.03
	case JawRounded:
		l.headRY *= 0.97
	case JawAverage, JawSubtle:
		// default
	}

	switch br {
	case BrowHeavy:
		l.headRY *= 1.02
	case BrowNormal, BrowLight, BrowArched:
		// default
	}

	// Apply ear-specific modifications
	switch e {
	case EarsLarge:
		l.earScale *= 1.30
	case EarsPointed:
		l.earScale *= 1.15 // Pointed ears are slightly larger
	case EarsSmall:
		l.earScale *= 0.70
	case EarsMedium, EarsRounded:
		// default scale
	}

	// Update ear attachment points based on current head geometry.
	// Ears attach at the lateral extent of the head at roughly eye level.
	l.earAttachL[0] = -l.headRX
	l.earAttachL[1] = l.headCenter[1] + l.headRY*0.15
	l.earAttachL[2] = l.headCenter[2]

	l.earAttachR[0] = l.headRX
	l.earAttachR[1] = l.headCenter[1] + l.headRY*0.15
	l.earAttachR[2] = l.headCenter[2]
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
	// ── Arms ──────────────────────────────────────────────────────────────
	upperArmVecL := vec3Sub(l.upperArmBottomL, l.upperArmTopL)
	upperArmVecR := vec3Sub(l.upperArmBottomR, l.upperArmTopR)
	foreArmVecL := vec3Sub(l.forearmBottomL, l.forearmTopL)
	foreArmVecR := vec3Sub(l.forearmBottomR, l.forearmTopR)

	l.upperArmBottomL = vec3Add(l.upperArmTopL, vec3Scale(upperArmVecL, s))
	l.upperArmBottomR = vec3Add(l.upperArmTopR, vec3Scale(upperArmVecR, s))

	// Elbow = new upper arm bottom
	l.forearmTopL = l.upperArmBottomL
	l.forearmTopR = l.upperArmBottomR
	l.forearmBottomL = vec3Add(l.forearmTopL, vec3Scale(foreArmVecL, s))
	l.forearmBottomR = vec3Add(l.forearmTopR, vec3Scale(foreArmVecR, s))

	// Wrist / hand
	l.handCenterL = Vec3{
		l.forearmBottomL[0],
		l.forearmBottomL[1] - l.handHH,
		l.forearmBottomL[2],
	}
	l.handCenterR = Vec3{
		l.forearmBottomR[0],
		l.forearmBottomR[1] - l.handHH,
		l.forearmBottomR[2],
	}

	// ── Legs ──────────────────────────────────────────────────────────────
	upperLegVecL := vec3Sub(l.upperLegBottomL, l.upperLegTopL)
	upperLegVecR := vec3Sub(l.upperLegBottomR, l.upperLegTopR)
	lowerLegVecL := vec3Sub(l.lowerLegBottomL, l.lowerLegTopL)
	lowerLegVecR := vec3Sub(l.lowerLegBottomR, l.lowerLegTopR)

	l.upperLegBottomL = vec3Add(l.upperLegTopL, vec3Scale(upperLegVecL, s))
	l.upperLegBottomR = vec3Add(l.upperLegTopR, vec3Scale(upperLegVecR, s))

	// Knee = new upper leg bottom
	l.lowerLegTopL = l.upperLegBottomL
	l.lowerLegTopR = l.upperLegBottomR
	l.lowerLegBottomL = vec3Add(l.lowerLegTopL, vec3Scale(lowerLegVecL, s))
	l.lowerLegBottomR = vec3Add(l.lowerLegTopR, vec3Scale(lowerLegVecR, s))

	// Ankle / foot
	l.footCenterL = Vec3{
		l.lowerLegBottomL[0],
		l.lowerLegBottomL[1] + l.footHH,
		l.footCenterL[2],
	}
	l.footCenterR = Vec3{
		l.lowerLegBottomR[0],
		l.lowerLegBottomR[1] + l.footHH,
		l.footCenterR[2],
	}
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
