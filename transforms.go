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
	applyAge(&l, p.Age)
	applyShoulderWidth(&l, p.ShoulderWidth)
	applyHipWidth(&l, p.HipWidth)
	applyLimbLength(&l, p.LimbLength)
	applyNeckLength(&l, p.NeckLength)
	applyHandSize(&l, p.HandSize, p.FingerLength)
	applyFootSize(&l, p.FootSize)
	applyFacialFeatures(&l, p.FaceShape, p.Jaw, p.Brow, p.Ears)
	applyPosture(&l, p.Posture, rng)

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

func applyAge(l *bodyLayout, a Age) {
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
		l.headRX *= 1.15
		l.headRY *= 1.18
		l.headRZ *= 1.15
	case AgeToddler:
		scaleAll(l, 0.45)
		l.headRX *= 1.35
		l.headRY *= 1.40
		l.headRZ *= 1.35
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

	switch e {
	case EarsLarge:
		l.headRX *= 1.05
	case EarsPointed:
		l.headRX *= 1.04
	case EarsSmall:
		l.headRX *= 0.97
	case EarsMedium, EarsRounded:
		// default
	}
}

// ─── Posture ─────────────────────────────────────────────────────────────────

func applyPosture(l *bodyLayout, p Posture, rng *splitmix64) {
	jitter := (rng.Float32() - 0.5) * 0.003
	switch p {
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

// scaleAll uniformly scales all positions and radii around the world origin
// (feet remain near Y≈0).
func scaleAll(l *bodyLayout, s float32) {
	l.totalHeight *= s

	scaleV3(&l.headCenter, s)
	l.headRX *= s
	l.headRY *= s
	l.headRZ *= s

	scaleV3(&l.neckBottom, s)
	scaleV3(&l.neckTop, s)
	l.neckRadius *= s

	scaleV3(&l.chestBottom, s)
	scaleV3(&l.chestTop, s)
	l.chestRX *= s
	l.chestRZ *= s

	scaleV3(&l.abdomenBottom, s)
	scaleV3(&l.abdomenTop, s)
	l.abdomenRX *= s
	l.abdomenRZ *= s

	scaleV3(&l.hipsBottom, s)
	scaleV3(&l.hipsTop, s)
	l.hipsRX *= s
	l.hipsRZ *= s

	scaleV3(&l.upperArmTopL, s)
	scaleV3(&l.upperArmBottomL, s)
	scaleV3(&l.upperArmTopR, s)
	scaleV3(&l.upperArmBottomR, s)
	l.upperArmRadius *= s

	scaleV3(&l.forearmTopL, s)
	scaleV3(&l.forearmBottomL, s)
	scaleV3(&l.forearmTopR, s)
	scaleV3(&l.forearmBottomR, s)
	l.forearmRadius *= s

	scaleV3(&l.handCenterL, s)
	scaleV3(&l.handCenterR, s)
	l.handHW *= s
	l.handHH *= s
	l.handHD *= s

	scaleV3(&l.upperLegTopL, s)
	scaleV3(&l.upperLegBottomL, s)
	scaleV3(&l.upperLegTopR, s)
	scaleV3(&l.upperLegBottomR, s)
	l.upperLegRadius *= s

	scaleV3(&l.lowerLegTopL, s)
	scaleV3(&l.lowerLegBottomL, s)
	scaleV3(&l.lowerLegTopR, s)
	scaleV3(&l.lowerLegBottomR, s)
	l.lowerLegRadius *= s

	scaleV3(&l.footCenterL, s)
	scaleV3(&l.footCenterR, s)
	l.footHW *= s
	l.footHH *= s
	l.footHD *= s
}

// scaleHeight scales only the Y component of all position vectors and vertical
// radii (used by species that are shorter but not narrower).
func scaleHeight(l *bodyLayout, s float32) {
	l.totalHeight *= s

	scaleV3Y(&l.headCenter, s)
	l.headRY *= s
	scaleV3Y(&l.neckBottom, s)
	scaleV3Y(&l.neckTop, s)
	scaleV3Y(&l.chestBottom, s)
	scaleV3Y(&l.chestTop, s)
	scaleV3Y(&l.abdomenBottom, s)
	scaleV3Y(&l.abdomenTop, s)
	scaleV3Y(&l.hipsBottom, s)
	scaleV3Y(&l.hipsTop, s)

	scaleV3Y(&l.upperArmTopL, s)
	scaleV3Y(&l.upperArmBottomL, s)
	scaleV3Y(&l.upperArmTopR, s)
	scaleV3Y(&l.upperArmBottomR, s)

	scaleV3Y(&l.forearmTopL, s)
	scaleV3Y(&l.forearmBottomL, s)
	scaleV3Y(&l.forearmTopR, s)
	scaleV3Y(&l.forearmBottomR, s)

	scaleV3Y(&l.handCenterL, s)
	scaleV3Y(&l.handCenterR, s)
	l.handHH *= s

	scaleV3Y(&l.upperLegTopL, s)
	scaleV3Y(&l.upperLegBottomL, s)
	scaleV3Y(&l.upperLegTopR, s)
	scaleV3Y(&l.upperLegBottomR, s)

	scaleV3Y(&l.lowerLegTopL, s)
	scaleV3Y(&l.lowerLegBottomL, s)
	scaleV3Y(&l.lowerLegTopR, s)
	scaleV3Y(&l.lowerLegBottomR, s)

	scaleV3Y(&l.footCenterL, s)
	scaleV3Y(&l.footCenterR, s)
	l.footHH *= s
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
