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

	// Finger dimensions (cylinder segments)
	// Each finger has proximal, middle, and distal phalanges.
	// Thumb has only proximal and distal (2 segments).
	defaultFingerRadius        = 0.008  // Finger cylinder radius
	defaultProximalLength      = 0.025  // Proximal phalanx length
	defaultMiddleLength        = 0.018  // Middle phalanx length
	defaultDistalLength        = 0.015  // Distal phalanx length
	defaultThumbProximalLength = 0.020  // Thumb proximal length
	defaultThumbDistalLength   = 0.018  // Thumb distal length
	defaultFingerSpacing       = 0.0085 // Spacing between fingers

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

	// Toe dimensions (cylinder segments)
	// Each toe has proximal and distal phalanges (big toe has 2, others have 3).
	// Dimensions are smaller than fingers.
	defaultToeRadius         = 0.006  // Toe cylinder radius
	defaultToeProximalLength = 0.015  // Proximal phalanx length
	defaultToeMiddleLength   = 0.010  // Middle phalanx length
	defaultToeDistalLength   = 0.008  // Distal phalanx length
	defaultBigToeProximal    = 0.020  // Big toe proximal length (larger)
	defaultBigToeDistal      = 0.015  // Big toe distal length
	defaultToeSpacing        = 0.0065 // Spacing between toes

	// Ear dimensions
	// Ears attach to the lateral sides of the head at roughly eye level.
	// Ear scale factor controls overall ear size relative to head radius.
	defaultEarScale  = 0.35 // Ear height as fraction of head lateral radius
	defaultEarYRatio = 0.15 // Vertical offset from head center as fraction of headRY
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

	// ── Fingers ─────────────────────────────────────────────────────────────
	// Finger segment lengths (phalanges) and radius
	fingerRadius        float32
	proximalLength      float32 // Proximal phalanx (closest to palm)
	middleLength        float32 // Middle phalanx
	distalLength        float32 // Distal phalanx (fingertip)
	thumbProximalLength float32 // Thumb has different proportions
	thumbDistalLength   float32
	fingerSpacing       float32 // Lateral spacing between fingers
	fingerLengthMult    float32 // Multiplier from FingerLength param

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

	// ── Toes ────────────────────────────────────────────────────────────────
	toeRadius         float32
	toeProximalLength float32
	toeMiddleLength   float32
	toeDistalLength   float32
	bigToeProximal    float32
	bigToeDistal      float32
	toeSpacing        float32

	// ── Ears ────────────────────────────────────────────────────────────────
	earAttachL Vec3    // Left ear attachment point (derived from head geometry)
	earAttachR Vec3    // Right ear attachment point
	earScale   float32 // Overall ear size multiplier
}

// initHeadLayout initializes head and neck dimensions for the body layout.
func initHeadLayout(l *bodyLayout) {
	l.headCenter = Vec3{0, defaultHeadCenterY, 0}
	l.headRX = defaultHeadRX
	l.headRY = defaultHeadRY
	l.headRZ = defaultHeadRZ
	l.neckBottom = Vec3{0, defaultNeckBottomY, 0}
	l.neckTop = Vec3{0, defaultNeckTopY, 0}
	l.neckRadius = defaultNeckRadius
}

// initTorsoLayout initializes chest, abdomen, and hips dimensions.
func initTorsoLayout(l *bodyLayout) {
	l.chestBottom = Vec3{0, defaultChestBottomY, 0}
	l.chestTop = Vec3{0, defaultChestTopY, 0}
	l.chestRX = defaultChestRX
	l.chestRZ = defaultChestRZ
	l.abdomenBottom = Vec3{0, defaultAbdomenBottomY, 0}
	l.abdomenTop = Vec3{0, defaultAbdomenTopY, 0}
	l.abdomenRX = defaultAbdomenRX
	l.abdomenRZ = defaultAbdomenRZ
	l.hipsBottom = Vec3{0, defaultHipsBottomY, 0}
	l.hipsTop = Vec3{0, defaultHipsTopY, 0}
	l.hipsRX = defaultHipsRX
	l.hipsRZ = defaultHipsRZ
}

// initArmLayout initializes upper arms, forearms, hands, and fingers.
func initArmLayout(l *bodyLayout) {
	l.upperArmTopL = Vec3{-defaultShoulderX, defaultUpperArmTopY, 0}
	l.upperArmBottomL = Vec3{-defaultShoulderX, defaultUpperArmBottomY, 0}
	l.upperArmTopR = Vec3{defaultShoulderX, defaultUpperArmTopY, 0}
	l.upperArmBottomR = Vec3{defaultShoulderX, defaultUpperArmBottomY, 0}
	l.upperArmRadius = defaultUpperArmRadius
	l.forearmTopL = Vec3{-defaultShoulderX, defaultUpperArmBottomY, 0}
	l.forearmBottomL = Vec3{-defaultShoulderX, defaultForearmBottomY, 0}
	l.forearmTopR = Vec3{defaultShoulderX, defaultUpperArmBottomY, 0}
	l.forearmBottomR = Vec3{defaultShoulderX, defaultForearmBottomY, 0}
	l.forearmRadius = defaultForearmRadius
	l.handCenterL = Vec3{-defaultShoulderX, defaultHandCenterY, 0}
	l.handCenterR = Vec3{defaultShoulderX, defaultHandCenterY, 0}
	l.handHW = defaultHandHW
	l.handHH = defaultHandHH
	l.handHD = defaultHandHD
	l.fingerRadius = defaultFingerRadius
	l.proximalLength = defaultProximalLength
	l.middleLength = defaultMiddleLength
	l.distalLength = defaultDistalLength
	l.thumbProximalLength = defaultThumbProximalLength
	l.thumbDistalLength = defaultThumbDistalLength
	l.fingerSpacing = defaultFingerSpacing
	l.fingerLengthMult = 1.0
}

// initLegLayout initializes upper legs, lower legs, feet, and toes.
func initLegLayout(l *bodyLayout) {
	l.upperLegTopL = Vec3{-defaultHipSocketX, defaultUpperLegTopY, 0}
	l.upperLegBottomL = Vec3{-defaultHipSocketX, defaultUpperLegBottomY, 0}
	l.upperLegTopR = Vec3{defaultHipSocketX, defaultUpperLegTopY, 0}
	l.upperLegBottomR = Vec3{defaultHipSocketX, defaultUpperLegBottomY, 0}
	l.upperLegRadius = defaultUpperLegRadius
	l.lowerLegTopL = Vec3{-defaultHipSocketX, defaultUpperLegBottomY, 0}
	l.lowerLegBottomL = Vec3{-defaultHipSocketX, defaultLowerLegBottomY, 0}
	l.lowerLegTopR = Vec3{defaultHipSocketX, defaultUpperLegBottomY, 0}
	l.lowerLegBottomR = Vec3{defaultHipSocketX, defaultLowerLegBottomY, 0}
	l.lowerLegRadius = defaultLowerLegRadius
	l.footCenterL = Vec3{-defaultHipSocketX, defaultFootCenterY, defaultFootCenterZ}
	l.footCenterR = Vec3{defaultHipSocketX, defaultFootCenterY, defaultFootCenterZ}
	l.footHW = defaultFootHW
	l.footHH = defaultFootHH
	l.footHD = defaultFootHD
	l.toeRadius = defaultToeRadius
	l.toeProximalLength = defaultToeProximalLength
	l.toeMiddleLength = defaultToeMiddleLength
	l.toeDistalLength = defaultToeDistalLength
	l.bigToeProximal = defaultBigToeProximal
	l.bigToeDistal = defaultBigToeDistal
	l.toeSpacing = defaultToeSpacing
}

// initEarLayout initializes ear attachment points and scale.
func initEarLayout(l *bodyLayout) {
	earY := float32(defaultHeadCenterY + defaultHeadRY*defaultEarYRatio)
	l.earAttachL = Vec3{-defaultHeadRX, earY, 0}
	l.earAttachR = Vec3{defaultHeadRX, earY, 0}
	l.earScale = defaultEarScale
}

// defaultBodyLayout returns a neutral 1.75 m adult humanoid in T-pose.
func defaultBodyLayout() bodyLayout {
	l := bodyLayout{totalHeight: defaultTotalHeight}
	initHeadLayout(&l)
	initTorsoLayout(&l)
	initArmLayout(&l)
	initLegLayout(&l)
	initEarLayout(&l)
	return l
}

// meshBuildContext holds state for mesh assembly operations.
type meshBuildContext struct {
	builder  *meshBuilder
	atlas    UVAtlas
	circSegs int // radial resolution for cylinders
}

// appendPart generates geometry, remaps UVs, and appends to the builder.
func (ctx *meshBuildContext) appendPart(verts []Vertex, idxs []uint32, uvRegion UVRegion) {
	remapUVs(verts, uvRegion)
	ctx.builder.append(verts, idxs)
}

// appendCyl generates a tapered cylinder and appends it to the mesh.
func (ctx *meshBuildContext) appendCyl(bottom, top Vec3, bottomR, topR float32, bottomCap, topCap bool, uv UVRegion) {
	v, i := generateCylinder(bottom, top, bottomR, topR, ctx.circSegs, bottomCap, topCap)
	ctx.appendPart(v, i, uv)
}

// buildHeadParts assembles the head, face, and optional skull cap.
func (ctx *meshBuildContext) buildHeadParts(layout bodyLayout, opts buildOptions) {
	const (
		latSegs = 6 // latitude rings for ellipsoid head
		lonSegs = 8 // longitude segments for ellipsoid head
	)

	v, i := generateEllipsoid(layout.headCenter, layout.headRX, layout.headRY, layout.headRZ, latSegs, lonSegs)
	ctx.appendPart(v, i, ctx.atlas.Head)

	v, i = generateFaceMesh(layout.headCenter, layout.headRX, layout.headRY, layout.headRZ, opts.faceShape, opts.jaw, opts.brow)
	ctx.appendPart(v, i, ctx.atlas.Face)

	if opts.hasHairSlot {
		v, i = generateSkullCap(layout.headCenter, layout.headRX, layout.headRY, layout.headRZ)
		ctx.appendPart(v, i, ctx.atlas.SkullCap)
	}
}

// buildTorso assembles the neck, chest, abdomen, and hips.
func (ctx *meshBuildContext) buildTorso(layout bodyLayout) {
	ctx.appendCyl(layout.neckBottom, layout.neckTop, layout.neckRadius, layout.neckRadius, false, false, ctx.atlas.Neck)
	ctx.appendCyl(layout.chestBottom, layout.chestTop, layout.chestRX*0.82, layout.chestRX, false, false, ctx.atlas.Chest)
	ctx.appendCyl(layout.abdomenBottom, layout.abdomenTop, layout.abdomenRX, layout.abdomenRX*0.88, false, false, ctx.atlas.Abdomen)
	ctx.appendCyl(layout.hipsBottom, layout.hipsTop, layout.hipsRX, layout.hipsRX*0.95, true, false, ctx.atlas.Hips)
}

// buildArms assembles the upper arms and forearms.
func (ctx *meshBuildContext) buildArms(layout bodyLayout) {
	ctx.appendCyl(layout.upperArmTopL, layout.upperArmBottomL, layout.upperArmRadius, layout.upperArmRadius*0.85, false, false, ctx.atlas.UpperArmL)
	ctx.appendCyl(layout.upperArmTopR, layout.upperArmBottomR, layout.upperArmRadius, layout.upperArmRadius*0.85, false, false, ctx.atlas.UpperArmR)
	ctx.appendCyl(layout.forearmTopL, layout.forearmBottomL, layout.forearmRadius, layout.forearmRadius*0.80, false, false, ctx.atlas.ForearmL)
	ctx.appendCyl(layout.forearmTopR, layout.forearmBottomR, layout.forearmRadius, layout.forearmRadius*0.80, false, false, ctx.atlas.ForearmR)
}

// buildHands assembles both hands with fingers.
func (ctx *meshBuildContext) buildHands(layout bodyLayout) {
	fingerDir := Vec3{0, -1, 0} // fingers point down in T-pose
	v, i := generateHand(layout.handCenterL, layout.handHW, layout.handHH, layout.handHD, fingerDir, true,
		layout.fingerRadius, layout.proximalLength, layout.middleLength, layout.distalLength,
		layout.thumbProximalLength, layout.thumbDistalLength, layout.fingerSpacing, layout.fingerLengthMult)
	ctx.appendPart(v, i, ctx.atlas.HandL)

	v, i = generateHand(layout.handCenterR, layout.handHW, layout.handHH, layout.handHD, fingerDir, false,
		layout.fingerRadius, layout.proximalLength, layout.middleLength, layout.distalLength,
		layout.thumbProximalLength, layout.thumbDistalLength, layout.fingerSpacing, layout.fingerLengthMult)
	ctx.appendPart(v, i, ctx.atlas.HandR)
}

// buildLegs assembles the upper and lower legs.
func (ctx *meshBuildContext) buildLegs(layout bodyLayout) {
	ctx.appendCyl(layout.upperLegTopL, layout.upperLegBottomL, layout.upperLegRadius, layout.upperLegRadius*0.85, false, false, ctx.atlas.UpperLegL)
	ctx.appendCyl(layout.upperLegTopR, layout.upperLegBottomR, layout.upperLegRadius, layout.upperLegRadius*0.85, false, false, ctx.atlas.UpperLegR)
	ctx.appendCyl(layout.lowerLegTopL, layout.lowerLegBottomL, layout.lowerLegRadius, layout.lowerLegRadius*0.75, false, true, ctx.atlas.LowerLegL)
	ctx.appendCyl(layout.lowerLegTopR, layout.lowerLegBottomR, layout.lowerLegRadius, layout.lowerLegRadius*0.75, false, true, ctx.atlas.LowerLegR)
}

// buildFeet assembles both feet with toes.
func (ctx *meshBuildContext) buildFeet(layout bodyLayout) {
	toeDir := Vec3{0, 0, 1} // toes point forward in T-pose
	v, i := generateFoot(layout.footCenterL, layout.footHW, layout.footHH, layout.footHD, toeDir, true,
		layout.toeRadius, layout.toeProximalLength, layout.toeMiddleLength, layout.toeDistalLength,
		layout.bigToeProximal, layout.bigToeDistal, layout.toeSpacing)
	ctx.appendPart(v, i, ctx.atlas.FootL)

	v, i = generateFoot(layout.footCenterR, layout.footHW, layout.footHH, layout.footHD, toeDir, false,
		layout.toeRadius, layout.toeProximalLength, layout.toeMiddleLength, layout.toeDistalLength,
		layout.bigToeProximal, layout.bigToeDistal, layout.toeSpacing)
	ctx.appendPart(v, i, ctx.atlas.FootR)
}

// buildEars assembles both ears.
func (ctx *meshBuildContext) buildEars(layout bodyLayout) {
	v, i := generateEar(layout.earAttachL, layout.earScale, true)
	ctx.appendPart(v, i, ctx.atlas.EarL)
	v, i = generateEar(layout.earAttachR, layout.earScale, false)
	ctx.appendPart(v, i, ctx.atlas.EarR)
}

// buildMesh assembles the full humanoid mesh from the given body layout.
// The mesh key is used by the Kaiju engine's mesh cache.
// opts controls optional features like skull cap and face mesh parameters.
func buildMesh(layout bodyLayout, key string, opts buildOptions) *Mesh {
	builder := newMeshBuilder()
	ctx := &meshBuildContext{
		builder:  &builder,
		atlas:    defaultUVAtlas(),
		circSegs: 8, // radial resolution for cylinders
	}

	ctx.buildHeadParts(layout, opts)
	ctx.buildTorso(layout)
	ctx.buildArms(layout)
	ctx.buildHands(layout)
	ctx.buildLegs(layout)
	ctx.buildFeet(layout)
	ctx.buildEars(layout)

	mesh := builder.build(key)

	// Merge boundary vertices to eliminate visible seams at body part joints.
	epsilon := scaledEpsilon(layout.totalHeight)
	mesh.Vertices, mesh.Indices = mergeVertices(mesh.Vertices, mesh.Indices, epsilon)

	applySkinColor(mesh.Vertices, opts.skinColor)
	return mesh
}

// applySkinColor sets the Color field of all vertices to the given skin color.
func applySkinColor(vertices []Vertex, color Color) {
	for i := range vertices {
		vertices[i].Color = color
	}
}
