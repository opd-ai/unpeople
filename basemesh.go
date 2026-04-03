// Package unpeople – base mesh / body layout
//
// The bodyLayout struct holds the skeletal dimensions of a humanoid figure in
// a neutral T-pose.  All measurements are in metres.  The coordinate system is
// right-handed with Y-up: feet rest at Y≈0, the top of the head is at
// Y≈totalHeight; X is lateral (positive = character's right); Z is forward
// (positive = towards viewer).
//
// This approximates a MakeHuman-exported neutral base model: a bipedal
// humanoid with average adult proportions.  The actual geometry is assembled
// from cylindrical / ellipsoidal / box primitives that match the proportions
// of a standard MakeHuman export (head ~12 % of total height, upper leg ~25 %,
// arm span ≈ height, etc.).
package unpeople

// ─── Default Body Dimensions (metres) ────────────────────────────────────────
// These constants define a neutral adult humanoid body matching MakeHuman
// proportions. All values are in metres. See MakeHuman documentation for
// anatomical reference points.

const (
	// Overall body measurements
	defaultTotalHeight = 1.75 // Adult average standing height

	// Head dimensions (ellipsoid radii)
	defaultHeadCenterY = 1.665 // Vertical center of head ellipsoid
	defaultHeadRX      = 0.090 // Lateral radius (ear to ear)
	defaultHeadRY      = 0.115 // Vertical radius (chin to crown)
	defaultHeadRZ      = 0.090 // Front-back radius (nose to occiput)

	// Neck dimensions
	defaultNeckBottomY = 1.500 // Base of neck (at shoulders)
	defaultNeckTopY    = 1.555 // Top of neck (at skull base)
	defaultNeckRadius  = 0.045 // Neck cylinder radius

	// Chest/thorax dimensions
	defaultChestBottomY = 1.150 // Bottom of chest (at diaphragm)
	defaultChestTopY    = 1.500 // Top of chest (at clavicle)
	defaultChestRX      = 0.185 // Chest lateral half-width
	defaultChestRZ      = 0.115 // Chest front-back half-depth

	// Abdomen dimensions
	defaultAbdomenBottomY = 0.950 // Bottom of abdomen (at iliac crest)
	defaultAbdomenTopY    = 1.150 // Top of abdomen (at diaphragm)
	defaultAbdomenRX      = 0.155 // Abdomen lateral half-width
	defaultAbdomenRZ      = 0.095 // Abdomen front-back half-depth

	// Hips/pelvis dimensions
	defaultHipsBottomY = 0.820 // Bottom of hips (at hip joint)
	defaultHipsTopY    = 0.950 // Top of hips (at iliac crest)
	defaultHipsRX      = 0.165 // Hips lateral half-width
	defaultHipsRZ      = 0.110 // Hips front-back half-depth

	// Arm attachment (shoulder) X offset from centerline
	defaultShoulderX = 0.235

	// Upper arm dimensions
	defaultUpperArmTopY    = 1.430 // Shoulder height
	defaultUpperArmBottomY = 1.100 // Elbow height
	defaultUpperArmRadius  = 0.048 // Upper arm cylinder radius

	// Forearm dimensions
	defaultForearmBottomY = 0.780 // Wrist height
	defaultForearmRadius  = 0.038 // Forearm cylinder radius

	// Hand dimensions (box half-extents)
	defaultHandCenterY = 0.710 // Hand center height (palm)
	defaultHandHW      = 0.045 // Hand half-width (across palm)
	defaultHandHH      = 0.065 // Hand half-height (finger length)
	defaultHandHD      = 0.022 // Hand half-depth (palm thickness)

	// Leg attachment (hip socket) X offset from centerline
	defaultHipSocketX = 0.095

	// Upper leg dimensions
	defaultUpperLegTopY    = 0.820 // Hip joint height
	defaultUpperLegBottomY = 0.480 // Knee height
	defaultUpperLegRadius  = 0.078 // Upper leg cylinder radius

	// Lower leg dimensions
	defaultLowerLegBottomY = 0.090 // Ankle height
	defaultLowerLegRadius  = 0.055 // Lower leg cylinder radius

	// Foot dimensions (box half-extents)
	defaultFootCenterY = 0.038 // Foot center height (mid-foot)
	defaultFootCenterZ = 0.040 // Foot center offset (forward)
	defaultFootHW      = 0.058 // Foot half-width
	defaultFootHH      = 0.038 // Foot half-height
	defaultFootHD      = 0.120 // Foot half-depth (toe to heel)
)

// bodyLayout stores every dimensional parameter needed to assemble the mesh.
type bodyLayout struct {
	totalHeight float32

	// ── Head ────────────────────────────────────────────────────────────────
	headCenter Vec3
	headRX     float32 // radius X (lateral)
	headRY     float32 // radius Y (vertical)
	headRZ     float32 // radius Z (front-back)

	// ── Neck ────────────────────────────────────────────────────────────────
	neckBottom Vec3
	neckTop    Vec3
	neckRadius float32

	// ── Chest ───────────────────────────────────────────────────────────────
	chestBottom Vec3
	chestTop    Vec3
	chestRX     float32 // half-width (lateral)
	chestRZ     float32 // half-depth (front-back)

	// ── Abdomen ─────────────────────────────────────────────────────────────
	abdomenBottom Vec3
	abdomenTop    Vec3
	abdomenRX     float32
	abdomenRZ     float32

	// ── Hips / Pelvis ───────────────────────────────────────────────────────
	hipsBottom Vec3
	hipsTop    Vec3
	hipsRX     float32
	hipsRZ     float32

	// ── Upper arms ──────────────────────────────────────────────────────────
	upperArmTopL    Vec3
	upperArmBottomL Vec3
	upperArmTopR    Vec3
	upperArmBottomR Vec3
	upperArmRadius  float32

	// ── Forearms ────────────────────────────────────────────────────────────
	forearmTopL    Vec3
	forearmBottomL Vec3
	forearmTopR    Vec3
	forearmBottomR Vec3
	forearmRadius  float32

	// ── Hands ───────────────────────────────────────────────────────────────
	handCenterL Vec3
	handCenterR Vec3
	handHW      float32 // half-width
	handHH      float32 // half-height (finger direction)
	handHD      float32 // half-depth

	// ── Upper legs ──────────────────────────────────────────────────────────
	upperLegTopL    Vec3
	upperLegBottomL Vec3
	upperLegTopR    Vec3
	upperLegBottomR Vec3
	upperLegRadius  float32

	// ── Lower legs ──────────────────────────────────────────────────────────
	lowerLegTopL    Vec3
	lowerLegBottomL Vec3
	lowerLegTopR    Vec3
	lowerLegBottomR Vec3
	lowerLegRadius  float32

	// ── Feet ────────────────────────────────────────────────────────────────
	footCenterL Vec3
	footCenterR Vec3
	footHW      float32
	footHH      float32
	footHD      float32 // half-depth (toe-to-heel)
}

// defaultBodyLayout returns a neutral 1.75 m adult humanoid in T-pose.
func defaultBodyLayout() bodyLayout {
	return bodyLayout{
		totalHeight: defaultTotalHeight,

		headCenter: Vec3{0, defaultHeadCenterY, 0},
		headRX:     defaultHeadRX,
		headRY:     defaultHeadRY,
		headRZ:     defaultHeadRZ,

		neckBottom: Vec3{0, defaultNeckBottomY, 0},
		neckTop:    Vec3{0, defaultNeckTopY, 0},
		neckRadius: defaultNeckRadius,

		chestBottom: Vec3{0, defaultChestBottomY, 0},
		chestTop:    Vec3{0, defaultChestTopY, 0},
		chestRX:     defaultChestRX,
		chestRZ:     defaultChestRZ,

		abdomenBottom: Vec3{0, defaultAbdomenBottomY, 0},
		abdomenTop:    Vec3{0, defaultAbdomenTopY, 0},
		abdomenRX:     defaultAbdomenRX,
		abdomenRZ:     defaultAbdomenRZ,

		hipsBottom: Vec3{0, defaultHipsBottomY, 0},
		hipsTop:    Vec3{0, defaultHipsTopY, 0},
		hipsRX:     defaultHipsRX,
		hipsRZ:     defaultHipsRZ,

		upperArmTopL:    Vec3{-defaultShoulderX, defaultUpperArmTopY, 0},
		upperArmBottomL: Vec3{-defaultShoulderX, defaultUpperArmBottomY, 0},
		upperArmTopR:    Vec3{defaultShoulderX, defaultUpperArmTopY, 0},
		upperArmBottomR: Vec3{defaultShoulderX, defaultUpperArmBottomY, 0},
		upperArmRadius:  defaultUpperArmRadius,

		forearmTopL:    Vec3{-defaultShoulderX, defaultUpperArmBottomY, 0},
		forearmBottomL: Vec3{-defaultShoulderX, defaultForearmBottomY, 0},
		forearmTopR:    Vec3{defaultShoulderX, defaultUpperArmBottomY, 0},
		forearmBottomR: Vec3{defaultShoulderX, defaultForearmBottomY, 0},
		forearmRadius:  defaultForearmRadius,

		handCenterL: Vec3{-defaultShoulderX, defaultHandCenterY, 0},
		handCenterR: Vec3{defaultShoulderX, defaultHandCenterY, 0},
		handHW:      defaultHandHW,
		handHH:      defaultHandHH,
		handHD:      defaultHandHD,

		upperLegTopL:    Vec3{-defaultHipSocketX, defaultUpperLegTopY, 0},
		upperLegBottomL: Vec3{-defaultHipSocketX, defaultUpperLegBottomY, 0},
		upperLegTopR:    Vec3{defaultHipSocketX, defaultUpperLegTopY, 0},
		upperLegBottomR: Vec3{defaultHipSocketX, defaultUpperLegBottomY, 0},
		upperLegRadius:  defaultUpperLegRadius,

		lowerLegTopL:    Vec3{-defaultHipSocketX, defaultUpperLegBottomY, 0},
		lowerLegBottomL: Vec3{-defaultHipSocketX, defaultLowerLegBottomY, 0},
		lowerLegTopR:    Vec3{defaultHipSocketX, defaultUpperLegBottomY, 0},
		lowerLegBottomR: Vec3{defaultHipSocketX, defaultLowerLegBottomY, 0},
		lowerLegRadius:  defaultLowerLegRadius,

		footCenterL: Vec3{-defaultHipSocketX, defaultFootCenterY, defaultFootCenterZ},
		footCenterR: Vec3{defaultHipSocketX, defaultFootCenterY, defaultFootCenterZ},
		footHW:      defaultFootHW,
		footHH:      defaultFootHH,
		footHD:      defaultFootHD,
	}
}

// buildMesh assembles the full humanoid mesh from the given body layout.
// The mesh key is used by the Kaiju engine's mesh cache.
func buildMesh(layout bodyLayout, key string) *Mesh {
	var b meshBuilder

	const (
		circSegs = 8 // radial resolution for cylinders
		latSegs  = 6 // latitude rings for ellipsoid head
		lonSegs  = 8 // longitude segments for ellipsoid head
	)

	// Head
	v, i := generateEllipsoid(layout.headCenter,
		layout.headRX, layout.headRY, layout.headRZ, latSegs, lonSegs)
	b.append(v, i)

	// Neck
	v, i = generateCylinder(layout.neckBottom, layout.neckTop,
		layout.neckRadius, layout.neckRadius, circSegs, false, false)
	b.append(v, i)

	// Chest (tapered: slightly narrower at bottom)
	v, i = generateCylinder(layout.chestBottom, layout.chestTop,
		layout.chestRX*0.82, layout.chestRX, circSegs, false, false)
	b.append(v, i)

	// Abdomen
	v, i = generateCylinder(layout.abdomenBottom, layout.abdomenTop,
		layout.abdomenRX, layout.abdomenRX*0.88, circSegs, false, false)
	b.append(v, i)

	// Hips / pelvis (closed at bottom)
	v, i = generateCylinder(layout.hipsBottom, layout.hipsTop,
		layout.hipsRX, layout.hipsRX*0.95, circSegs, true, false)
	b.append(v, i)

	// Upper arms
	v, i = generateCylinder(layout.upperArmTopL, layout.upperArmBottomL,
		layout.upperArmRadius, layout.upperArmRadius*0.85, circSegs, false, false)
	b.append(v, i)
	v, i = generateCylinder(layout.upperArmTopR, layout.upperArmBottomR,
		layout.upperArmRadius, layout.upperArmRadius*0.85, circSegs, false, false)
	b.append(v, i)

	// Forearms
	v, i = generateCylinder(layout.forearmTopL, layout.forearmBottomL,
		layout.forearmRadius, layout.forearmRadius*0.80, circSegs, false, false)
	b.append(v, i)
	v, i = generateCylinder(layout.forearmTopR, layout.forearmBottomR,
		layout.forearmRadius, layout.forearmRadius*0.80, circSegs, false, false)
	b.append(v, i)

	// Hands
	v, i = generateBox(layout.handCenterL, layout.handHW, layout.handHH, layout.handHD)
	b.append(v, i)
	v, i = generateBox(layout.handCenterR, layout.handHW, layout.handHH, layout.handHD)
	b.append(v, i)

	// Upper legs
	v, i = generateCylinder(layout.upperLegTopL, layout.upperLegBottomL,
		layout.upperLegRadius, layout.upperLegRadius*0.85, circSegs, false, false)
	b.append(v, i)
	v, i = generateCylinder(layout.upperLegTopR, layout.upperLegBottomR,
		layout.upperLegRadius, layout.upperLegRadius*0.85, circSegs, false, false)
	b.append(v, i)

	// Lower legs (closed at ankle)
	v, i = generateCylinder(layout.lowerLegTopL, layout.lowerLegBottomL,
		layout.lowerLegRadius, layout.lowerLegRadius*0.75, circSegs, false, true)
	b.append(v, i)
	v, i = generateCylinder(layout.lowerLegTopR, layout.lowerLegBottomR,
		layout.lowerLegRadius, layout.lowerLegRadius*0.75, circSegs, false, true)
	b.append(v, i)

	// Feet
	v, i = generateBox(layout.footCenterL, layout.footHW, layout.footHH, layout.footHD)
	b.append(v, i)
	v, i = generateBox(layout.footCenterR, layout.footHW, layout.footHH, layout.footHD)
	b.append(v, i)

	return b.build(key)
}
