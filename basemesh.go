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

// bodyLayout stores every dimensional parameter needed to assemble the mesh.
type bodyLayout struct {
	totalHeight float32

	// ── Head ────────────────────────────────────────────────────────────────
	headCenter  Vec3
	headRX      float32 // radius X (lateral)
	headRY      float32 // radius Y (vertical)
	headRZ      float32 // radius Z (front-back)

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
		totalHeight: 1.75,

		headCenter: Vec3{0, 1.665, 0},
		headRX:     0.090,
		headRY:     0.115,
		headRZ:     0.090,

		neckBottom: Vec3{0, 1.500, 0},
		neckTop:    Vec3{0, 1.555, 0},
		neckRadius: 0.045,

		chestBottom: Vec3{0, 1.150, 0},
		chestTop:    Vec3{0, 1.500, 0},
		chestRX:     0.185,
		chestRZ:     0.115,

		abdomenBottom: Vec3{0, 0.950, 0},
		abdomenTop:    Vec3{0, 1.150, 0},
		abdomenRX:     0.155,
		abdomenRZ:     0.095,

		hipsBottom: Vec3{0, 0.820, 0},
		hipsTop:    Vec3{0, 0.950, 0},
		hipsRX:     0.165,
		hipsRZ:     0.110,

		upperArmTopL:    Vec3{-0.235, 1.430, 0},
		upperArmBottomL: Vec3{-0.235, 1.100, 0},
		upperArmTopR:    Vec3{0.235, 1.430, 0},
		upperArmBottomR: Vec3{0.235, 1.100, 0},
		upperArmRadius:  0.048,

		forearmTopL:    Vec3{-0.235, 1.100, 0},
		forearmBottomL: Vec3{-0.235, 0.780, 0},
		forearmTopR:    Vec3{0.235, 1.100, 0},
		forearmBottomR: Vec3{0.235, 0.780, 0},
		forearmRadius:  0.038,

		handCenterL: Vec3{-0.235, 0.710, 0},
		handCenterR: Vec3{0.235, 0.710, 0},
		handHW:      0.045,
		handHH:      0.065,
		handHD:      0.022,

		upperLegTopL:    Vec3{-0.095, 0.820, 0},
		upperLegBottomL: Vec3{-0.095, 0.480, 0},
		upperLegTopR:    Vec3{0.095, 0.820, 0},
		upperLegBottomR: Vec3{0.095, 0.480, 0},
		upperLegRadius:  0.078,

		lowerLegTopL:    Vec3{-0.095, 0.480, 0},
		lowerLegBottomL: Vec3{-0.095, 0.090, 0},
		lowerLegTopR:    Vec3{0.095, 0.480, 0},
		lowerLegBottomR: Vec3{0.095, 0.090, 0},
		lowerLegRadius:  0.055,

		footCenterL: Vec3{-0.095, 0.038, 0.040},
		footCenterR: Vec3{0.095, 0.038, 0.040},
		footHW:      0.058,
		footHH:      0.038,
		footHD:      0.120,
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
