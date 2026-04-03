package unpeople

import "math"

const tau = math.Pi * 2.0

// ─── Ellipsoid Vertex Helper ──────────────────────────────────────────────────

// ellipsoidVertex computes a single vertex on an ellipsoid surface given spherical
// angles theta (latitude) and phi (longitude), center position, radii, and UV coords.
func ellipsoidVertex(center Vec3, rx, ry, rz, theta, phi, uvU, uvV float32) Vertex {
	sinT := float32(math.Sin(float64(theta)))
	cosT := float32(math.Cos(float64(theta)))
	sinP := float32(math.Sin(float64(phi)))
	cosP := float32(math.Cos(float64(phi)))

	// Unit-sphere direction
	nx := cosP * sinT
	ny := cosT
	nz := sinP * sinT

	pos := Vec3{
		center[0] + rx*nx,
		center[1] + ry*ny,
		center[2] + rz*nz,
	}
	// Normal on an ellipsoid is the gradient of the implicit function,
	// proportional to (nx/rx², ny/ry², nz/rz²).
	n := vec3Normalize(Vec3{nx / (rx * rx), ny / (ry * ry), nz / (rz * rz)})

	return Vertex{
		Position: pos,
		Normal:   n,
		UV0:      Vec2{uvU, uvV},
		Color:    ColorGray,
		Tangent:  Vec4{-sinP, 0, cosP, 1},
	}
}

// ─── perpendicular ───────────────────────────────────────────────────────────

// perpendicular returns a unit vector perpendicular to v.
func perpendicular(v Vec3) Vec3 {
	if math.Abs(float64(v[0])) < 0.9 {
		return vec3Normalize(vec3Cross(v, Vec3{1, 0, 0}))
	}
	return vec3Normalize(vec3Cross(v, Vec3{0, 1, 0}))
}

// ─── Cylinder ────────────────────────────────────────────────────────────────

// cylinderCapacity calculates vertex and index capacities for a cylinder.
func cylinderCapacity(segments int, addBottomCap, addTopCap bool) (vertCap, idxCap int) {
	vertCap = 2 * segments
	idxCap = 6 * segments
	if addBottomCap {
		vertCap++
		idxCap += 3 * segments
	}
	if addTopCap {
		vertCap++
		idxCap += 3 * segments
	}
	return vertCap, idxCap
}

// appendCylinderCap adds a flat disc cap to vertex/index buffers.
// ringStart is the first vertex index of the ring. normal is the cap's facing
// direction. flipWinding reverses triangle winding for bottom caps.
func appendCylinderCap(
	verts []Vertex,
	idxs []uint32,
	center Vec3,
	normal Vec3,
	ringStart, segments int,
	flipWinding bool,
) ([]Vertex, []uint32) {
	capIdx := uint32(len(verts))
	verts = append(verts, Vertex{
		Position: center,
		Normal:   normal,
		UV0:      Vec2{0.5, 0.5},
		Color:    ColorGray,
	})
	for i := 0; i < segments; i++ {
		next := (i + 1) % segments
		if flipWinding {
			idxs = append(idxs, capIdx, uint32(ringStart+next), uint32(ringStart+i))
		} else {
			idxs = append(idxs, capIdx, uint32(ringStart+i), uint32(ringStart+next))
		}
	}
	return verts, idxs
}

// cylinderRingVertex creates a vertex for a cylinder ring at the given angle.
func cylinderRingVertex(center, perp, biperp Vec3, radius, angle, vCoord float32, tangent Vec4, seg, segments int) Vertex {
	c := float32(math.Cos(float64(angle)))
	s := float32(math.Sin(float64(angle)))
	offset := vec3Add(vec3Scale(perp, c), vec3Scale(biperp, s))
	return Vertex{
		Position: vec3Add(center, vec3Scale(offset, radius)),
		Normal:   vec3Normalize(offset),
		UV0:      Vec2{float32(seg) / float32(segments), vCoord},
		Color:    ColorGray,
		Tangent:  tangent,
	}
}

// appendCylinderRing adds a ring of vertices around a center point.
func appendCylinderRing(verts []Vertex, center, perp, biperp Vec3, radius, vCoord float32, tangent Vec4, segments int) []Vertex {
	for i := 0; i < segments; i++ {
		angle := float32(i) * float32(tau) / float32(segments)
		verts = append(verts, cylinderRingVertex(center, perp, biperp, radius, angle, vCoord, tangent, i, segments))
	}
	return verts
}

// appendCylinderSideIndices adds the side quad indices for a cylinder.
func appendCylinderSideIndices(idxs []uint32, bottomStart, topStart, segments int) []uint32 {
	for i := 0; i < segments; i++ {
		next := (i + 1) % segments
		b0 := uint32(bottomStart + i)
		b1 := uint32(bottomStart + next)
		t0 := uint32(topStart + i)
		t1 := uint32(topStart + next)
		idxs = append(idxs, b0, t0, b1, b1, t0, t1)
	}
	return idxs
}

// generateCylinder builds a (possibly tapered) cylinder from bottomCenter to
// topCenter. segments controls the radial resolution. addBottomCap /
// addTopCap add flat disc endcaps.
func generateCylinder(
	bottomCenter, topCenter Vec3,
	bottomRadius, topRadius float32,
	segments int,
	addBottomCap, addTopCap bool,
) ([]Vertex, []uint32) {
	vertCap, idxCap := cylinderCapacity(segments, addBottomCap, addTopCap)
	verts := make([]Vertex, 0, vertCap)
	idxs := make([]uint32, 0, idxCap)

	axis := vec3Normalize(vec3Sub(topCenter, bottomCenter))
	perp := perpendicular(axis)
	biperp := vec3Cross(axis, perp)
	tangent := Vec4{perp[0], perp[1], perp[2], 1}

	bottomStart := 0
	verts = appendCylinderRing(verts, bottomCenter, perp, biperp, bottomRadius, 0, tangent, segments)

	topStart := segments
	verts = appendCylinderRing(verts, topCenter, perp, biperp, topRadius, 1, tangent, segments)

	idxs = appendCylinderSideIndices(idxs, bottomStart, topStart, segments)

	if addBottomCap {
		downNorm := vec3Scale(axis, -1)
		verts, idxs = appendCylinderCap(verts, idxs, bottomCenter, downNorm, bottomStart, segments, true)
	}
	if addTopCap {
		verts, idxs = appendCylinderCap(verts, idxs, topCenter, axis, topStart, segments, false)
	}

	return verts, idxs
}

// ─── Sphere / Ellipsoid ──────────────────────────────────────────────────────

// appendQuadGridIndices adds triangle indices for a lat×lon vertex grid.
// The grid has (latSegs+1)×(lonSegs+1) vertices arranged in rows.
// This creates two triangles per quad, suitable for ellipsoid-like surfaces.
func appendQuadGridIndices(idxs []uint32, latSegs, lonSegs int) []uint32 {
	stride := uint32(lonSegs + 1)
	for lat := 0; lat < latSegs; lat++ {
		for lon := 0; lon < lonSegs; lon++ {
			a := uint32(lat)*stride + uint32(lon)
			b := a + stride
			idxs = append(idxs, a, b, a+1)
			idxs = append(idxs, b, b+1, a+1)
		}
	}
	return idxs
}

// generateEllipsoid builds a UV-sphere with independently controllable radii
// on each axis, centred at center. latSegs × lonSegs controls resolution.
func generateEllipsoid(
	center Vec3,
	rx, ry, rz float32,
	latSegs, lonSegs int,
) ([]Vertex, []uint32) {
	// Pre-calculate capacities: (latSegs+1)*(lonSegs+1) vertices,
	// and 6*latSegs*lonSegs indices (2 triangles per quad)
	vertCap := (latSegs + 1) * (lonSegs + 1)
	idxCap := 6 * latSegs * lonSegs
	verts := make([]Vertex, 0, vertCap)
	idxs := make([]uint32, 0, idxCap)

	for lat := 0; lat <= latSegs; lat++ {
		theta := float32(lat) * float32(math.Pi) / float32(latSegs)
		uvV := float32(lat) / float32(latSegs)

		for lon := 0; lon <= lonSegs; lon++ {
			phi := float32(lon) * float32(tau) / float32(lonSegs)
			uvU := float32(lon) / float32(lonSegs)
			verts = append(verts, ellipsoidVertex(center, rx, ry, rz, theta, phi, uvU, uvV))
		}
	}

	idxs = appendQuadGridIndices(idxs, latSegs, lonSegs)
	return verts, idxs
}

// ─── Box ─────────────────────────────────────────────────────────────────────

// generateBox builds an axis-aligned box centred at center with half-extents
// (hw, hh, hd). Each of the 6 faces is a flat quad with its own normal.
func generateBox(center Vec3, hw, hh, hd float32) ([]Vertex, []uint32) {
	cx, cy, cz := center[0], center[1], center[2]

	corners := [8]Vec3{
		{cx - hw, cy - hh, cz - hd}, // 0
		{cx + hw, cy - hh, cz - hd}, // 1
		{cx + hw, cy + hh, cz - hd}, // 2
		{cx - hw, cy + hh, cz - hd}, // 3
		{cx - hw, cy - hh, cz + hd}, // 4
		{cx + hw, cy - hh, cz + hd}, // 5
		{cx + hw, cy + hh, cz + hd}, // 6
		{cx - hw, cy + hh, cz + hd}, // 7
	}

	// Each face: CCW winding when viewed from the normal direction
	faces := [6][4]int{
		{0, 3, 2, 1}, // -Z (back)
		{4, 5, 6, 7}, // +Z (front)
		{0, 4, 7, 3}, // -X (left)
		{1, 2, 6, 5}, // +X (right)
		{0, 1, 5, 4}, // -Y (bottom)
		{3, 7, 6, 2}, // +Y (top)
	}
	normals := [6]Vec3{
		{0, 0, -1},
		{0, 0, 1},
		{-1, 0, 0},
		{1, 0, 0},
		{0, -1, 0},
		{0, 1, 0},
	}
	uvCorners := [4]Vec2{{0, 0}, {1, 0}, {1, 1}, {0, 1}}

	// Pre-allocate: 6 faces * 4 vertices = 24 vertices, 6 faces * 6 indices = 36 indices
	verts := make([]Vertex, 0, 24)
	idxs := make([]uint32, 0, 36)

	for f := 0; f < 6; f++ {
		base := uint32(len(verts))
		for v := 0; v < 4; v++ {
			verts = append(verts, Vertex{
				Position: corners[faces[f][v]],
				Normal:   normals[f],
				UV0:      uvCorners[v],
				Color:    ColorGray,
			})
		}
		idxs = append(idxs, base, base+1, base+2)
		idxs = append(idxs, base, base+2, base+3)
	}
	return verts, idxs
}

// ─── Ear ─────────────────────────────────────────────────────────────────────

// earDimensions holds computed dimensions for ear generation.
type earDimensions struct {
	height   float32
	width    float32
	depth    float32
	curve    float32
	tipRatio float32
	dir      float32
}

// computeEarDimensions calculates ear dimensions based on scale and side.
func computeEarDimensions(scale float32, isLeftEar bool) earDimensions {
	dir := float32(1.0)
	if isLeftEar {
		dir = -1.0
	}
	return earDimensions{
		height:   scale * 0.28,
		width:    scale * 0.12,
		depth:    scale * 0.08,
		curve:    scale * 0.04,
		tipRatio: 0.3,
		dir:      dir,
	}
}

// earSegmentPosition computes the inner and outer edge positions for an ear segment.
func earSegmentPosition(attachPoint Vec3, dim earDimensions, t float32) (inner, outer Vec3) {
	y := attachPoint[1] - dim.height*0.4 + dim.height*t
	widthMult := 1.0 - t*(1.0-dim.tipRatio)
	curveOffset := dim.curve * float32(math.Sin(float64(t*math.Pi)))

	inner = Vec3{
		attachPoint[0] + dim.dir*0.002,
		y,
		attachPoint[2] + dim.depth*0.5*widthMult - curveOffset*0.5,
	}
	outer = Vec3{
		attachPoint[0] + dim.dir*(dim.width*widthMult+curveOffset),
		y,
		attachPoint[2] - dim.depth*0.5*widthMult,
	}
	return inner, outer
}

// appendEarSegmentVertices adds inner and outer vertices for an ear segment.
func appendEarSegmentVertices(verts []Vertex, inner, outer, normal Vec3, vCoord float32) []Vertex {
	verts = append(verts, Vertex{
		Position: inner,
		Normal:   normal,
		UV0:      Vec2{0.0, vCoord},
		Color:    ColorGray,
	})
	verts = append(verts, Vertex{
		Position: outer,
		Normal:   normal,
		UV0:      Vec2{1.0, vCoord},
		Color:    ColorGray,
	})
	return verts
}

// appendEarQuadIndices adds indices for a quad strip segment with proper winding.
func appendEarQuadIndices(idxs []uint32, base uint32, isLeftEar bool) []uint32 {
	if isLeftEar {
		idxs = append(idxs, base, base+2, base+1)
		idxs = append(idxs, base+1, base+2, base+3)
	} else {
		idxs = append(idxs, base, base+1, base+2)
		idxs = append(idxs, base+1, base+3, base+2)
	}
	return idxs
}

// generateEar creates a simplified ear mesh as a curved, tapered shell.
func generateEar(attachPoint Vec3, scale float32, isLeftEar bool) ([]Vertex, []uint32) {
	dim := computeEarDimensions(scale, isLeftEar)
	const segs = 4

	verts := make([]Vertex, 0, (segs+1)*2)
	idxs := make([]uint32, 0, segs*6)
	outerNorm := Vec3{dim.dir, 0, 0}

	for i := 0; i <= segs; i++ {
		t := float32(i) / float32(segs)
		inner, outer := earSegmentPosition(attachPoint, dim, t)
		verts = appendEarSegmentVertices(verts, inner, outer, outerNorm, t)
	}

	for i := 0; i < segs; i++ {
		idxs = appendEarQuadIndices(idxs, uint32(i*2), isLeftEar)
	}

	return verts, idxs
}

// ─── Finger ──────────────────────────────────────────────────────────────────

// generateFinger creates a multi-segment finger using tapered cylinders.
// base is the attachment point, direction is the finger pointing direction,
// segments is a slice of segment lengths [proximal, middle, distal],
// radius is the finger cylinder radius. Returns vertices and indices.
func generateFinger(base, direction Vec3, segments []float32, radius float32) ([]Vertex, []uint32) {
	if len(segments) == 0 {
		return nil, nil
	}

	var builder meshBuilder
	dir := vec3Normalize(direction)
	segSegs := 4 // radial resolution for finger cylinders

	pos := base
	for i, length := range segments {
		// Taper radius along finger: each segment slightly smaller
		taperMult := 1.0 - float32(i)*0.12
		bottomR := radius * taperMult
		topR := radius * taperMult * 0.90

		endPos := vec3Add(pos, vec3Scale(dir, length))
		v, idx := generateCylinder(pos, endPos, bottomR, topR, segSegs, i == 0, i == len(segments)-1)
		builder.append(v, idx)
		pos = endPos
	}

	m := builder.build("")
	return m.Vertices, m.Indices
}

// handConfig holds configuration for hand generation.
type handConfig struct {
	fingerSegs []float32
	thumbSegs  []float32
	palmEdgeY  float32
	baseZ      float32
	xDir       float32
	dir        Vec3
}

// computeHandConfig sets up hand generation parameters.
func computeHandConfig(handCenter Vec3, hh float32, direction Vec3, isLeftHand bool, fingerLengthMult, proximalLen, middleLen, distalLen, thumbProximalLen, thumbDistalLen float32) handConfig {
	xDir := float32(1.0)
	if isLeftHand {
		xDir = -1.0
	}
	return handConfig{
		fingerSegs: []float32{
			proximalLen * fingerLengthMult,
			middleLen * fingerLengthMult,
			distalLen * fingerLengthMult,
		},
		thumbSegs: []float32{
			thumbProximalLen * fingerLengthMult,
			thumbDistalLen * fingerLengthMult,
		},
		palmEdgeY: handCenter[1] - hh,
		baseZ:     handCenter[2],
		xDir:      xDir,
		dir:       vec3Normalize(direction),
	}
}

// fingerScale returns the length scale factor for a finger by index.
func fingerScale(fingerIdx int) float32 {
	if fingerIdx == 3 {
		return 0.75 // Pinky is shorter
	}
	return 1.0
}

// scaleFingerSegments applies a scale factor to finger segment lengths.
func scaleFingerSegments(segs []float32, scale float32) []float32 {
	scaled := make([]float32, len(segs))
	for i, s := range segs {
		scaled[i] = s * scale
	}
	return scaled
}

// appendMainFingers generates the four main fingers (index, middle, ring, pinky).
func appendMainFingers(builder *meshBuilder, handCenter Vec3, hw float32, cfg handConfig, fingerRadius float32) {
	offsets := []float32{hw * 0.55, hw * 0.18, -hw * 0.18, -hw * 0.55}
	for i, xOff := range offsets {
		segs := scaleFingerSegments(cfg.fingerSegs, fingerScale(i))
		base := Vec3{handCenter[0] + xOff*cfg.xDir, cfg.palmEdgeY, cfg.baseZ}
		v, idx := generateFinger(base, cfg.dir, segs, fingerRadius)
		builder.append(v, idx)
	}
}

// appendThumb generates the thumb.
func appendThumb(builder *meshBuilder, handCenter Vec3, hw, hh, hd float32, cfg handConfig, fingerRadius float32) {
	thumbDir := vec3Normalize(Vec3{cfg.xDir * 0.5, -0.85, 0.2})
	thumbBase := Vec3{
		handCenter[0] + hw*0.95*cfg.xDir,
		handCenter[1] - hh*0.3,
		handCenter[2] + hd*0.3,
	}
	v, idx := generateFinger(thumbBase, thumbDir, cfg.thumbSegs, fingerRadius*1.15)
	builder.append(v, idx)
}

// generateHand creates a complete hand with palm box and five fingers.
func generateHand(
	handCenter Vec3,
	hw, hh, hd float32,
	direction Vec3,
	isLeftHand bool,
	fingerRadius, proximalLen, middleLen, distalLen float32,
	thumbProximalLen, thumbDistalLen float32,
	fingerSpacing float32,
	fingerLengthMult float32,
) ([]Vertex, []uint32) {
	var builder meshBuilder

	v, idx := generateBox(handCenter, hw, hh, hd)
	builder.append(v, idx)

	cfg := computeHandConfig(handCenter, hh, direction, isLeftHand, fingerLengthMult, proximalLen, middleLen, distalLen, thumbProximalLen, thumbDistalLen)

	appendMainFingers(&builder, handCenter, hw, cfg, fingerRadius)
	appendThumb(&builder, handCenter, hw, hh, hd, cfg, fingerRadius)

	m := builder.build("")
	return m.Vertices, m.Indices
}

// ─── Foot with Toes ──────────────────────────────────────────────────────────

// toeConfig holds the configuration for generating a single toe.
type toeConfig struct {
	segments []float32
	radius   float32
}

// buildToeConfig computes the toe segments and radius based on toe index.
func buildToeConfig(toeIdx int, toeRadius, toeProximal, toeMiddle, toeDistal, bigToeProximal, bigToeDistal float32) toeConfig {
	if toeIdx == 0 {
		// Big toe: 2 segments, larger radius
		return toeConfig{
			segments: []float32{bigToeProximal, bigToeDistal},
			radius:   toeRadius * 1.3,
		}
	}

	// Other toes: 3 segments, progressively smaller for outer toes
	scale := float32(1.0)
	if toeIdx == 4 {
		scale = 0.70 // Pinky is smallest
	} else if toeIdx == 3 {
		scale = 0.85 // Ring is slightly smaller
	}
	return toeConfig{
		segments: []float32{toeProximal * scale, toeMiddle * scale, toeDistal * scale},
		radius:   toeRadius,
	}
}

// generateFoot creates a complete foot with box base and five toes.
// footCenter is the foot box center, hw/hh/hd are foot half-extents, direction
// is the toe pointing direction (typically forward +Z), isLeftFoot determines
// toe placement mirroring, and remaining params define toe dimensions.
func generateFoot(
	footCenter Vec3,
	hw, hh, hd float32,
	direction Vec3,
	isLeftFoot bool,
	toeRadius, toeProximal, toeMiddle, toeDistal float32,
	bigToeProximal, bigToeDistal float32,
	toeSpacing float32,
) ([]Vertex, []uint32) {
	var builder meshBuilder

	v, idx := generateBox(footCenter, hw, hh, hd)
	builder.append(v, idx)

	dir := vec3Normalize(direction)
	toeEdgeZ := footCenter[2] + hd
	toeY := footCenter[1] - hh*0.3

	xDir := float32(1.0)
	if isLeftFoot {
		xDir = -1.0
	}

	toeXOffsets := []float32{hw * 0.60, hw * 0.25, -hw * 0.05, -hw * 0.35, -hw * 0.60}

	for i, xOff := range toeXOffsets {
		cfg := buildToeConfig(i, toeRadius, toeProximal, toeMiddle, toeDistal, bigToeProximal, bigToeDistal)
		toeBase := Vec3{footCenter[0] + xOff*xDir, toeY, toeEdgeZ}
		v, idx := generateFinger(toeBase, dir, cfg.segments, cfg.radius)
		builder.append(v, idx)
	}

	m := builder.build("")
	return m.Vertices, m.Indices
}

// ─── Skull Cap ───────────────────────────────────────────────────────────────

// generateSkullCap creates a hemisphere mesh covering the top of the head.
// This serves as a placeholder attachment surface for hair systems.
// headCenter is the ellipsoid center, rx/ry/rz are the head radii.
// The cap covers the upper half of the head with a slight overlap.
func generateSkullCap(headCenter Vec3, rx, ry, rz float32) ([]Vertex, []uint32) {
	// Skull cap: upper portion of head ellipsoid (above equator)
	// We generate only the top hemisphere with a slight overlap below equator
	const (
		latSegs = 4 // latitude rings (equator to pole)
		lonSegs = 8 // longitude segments
	)

	// Pre-calculate capacities
	vertCap := (latSegs + 1) * (lonSegs + 1)
	idxCap := 6 * latSegs * lonSegs
	verts := make([]Vertex, 0, vertCap)
	idxs := make([]uint32, 0, idxCap)

	// Offset cap slightly outward to sit on top of head
	capOffset := float32(0.002)
	capRX := rx + capOffset
	capRY := ry + capOffset
	capRZ := rz + capOffset

	// Generate hemisphere (theta from 0 to pi/2, with slight overlap at -0.1*pi)
	startTheta := float32(-0.1 * 3.14159) // Slight overlap below equator
	endTheta := float32(0.5 * 3.14159)    // Top pole

	for lat := 0; lat <= latSegs; lat++ {
		t := float32(lat) / float32(latSegs)
		theta := startTheta + t*(endTheta-startTheta)

		for lon := 0; lon <= lonSegs; lon++ {
			phi := float32(lon) * float32(tau) / float32(lonSegs)
			uvU := float32(lon) / float32(lonSegs)
			verts = append(verts, ellipsoidVertex(headCenter, capRX, capRY, capRZ, theta, phi, uvU, t))
		}
	}

	idxs = appendQuadGridIndices(idxs, latSegs, lonSegs)
	return verts, idxs
}

// ─── Face Mesh ───────────────────────────────────────────────────────────────

// faceDisplacement stores displacement offsets for facial features
type faceDisplacement struct {
	chinZ, chinY float32 // Chin forward/back, up/down
	jawCornerX   float32 // Jaw width adjustment
	browRidgeZ   float32 // Brow ridge protrusion
	browRidgeY   float32 // Brow ridge height
	cheekX       float32 // Cheekbone width
}

// getFaceShapeDisplacement returns displacements for a face shape
func getFaceShapeDisplacement(fs FaceShape) faceDisplacement {
	switch fs {
	case FaceShapeRound:
		return faceDisplacement{chinZ: 0.002, chinY: 0.003, jawCornerX: 0.005, cheekX: 0.008}
	case FaceShapeSquare:
		return faceDisplacement{chinZ: 0.003, chinY: -0.003, jawCornerX: 0.008, cheekX: 0.006}
	case FaceShapeHeart:
		return faceDisplacement{chinZ: -0.002, chinY: 0.005, jawCornerX: -0.005, browRidgeY: 0.003, cheekX: -0.003}
	case FaceShapeDiamond:
		return faceDisplacement{chinZ: 0, chinY: 0.005, jawCornerX: -0.006, cheekX: 0.010}
	case FaceShapeOblong:
		return faceDisplacement{chinZ: -0.003, chinY: -0.008, jawCornerX: -0.006, cheekX: -0.004}
	default: // FaceShapeOval
		return faceDisplacement{}
	}
}

// getJawDisplacement returns displacements for jaw type
func getJawDisplacement(j Jaw) faceDisplacement {
	switch j {
	case JawProminent:
		return faceDisplacement{chinZ: 0.012, chinY: -0.006, jawCornerX: 0.004}
	case JawAngular:
		return faceDisplacement{chinZ: 0.004, jawCornerX: 0.010}
	case JawRounded:
		return faceDisplacement{chinY: 0.002, jawCornerX: -0.004}
	case JawSubtle:
		return faceDisplacement{chinZ: -0.006, chinY: 0.004, jawCornerX: -0.006}
	default: // JawAverage
		return faceDisplacement{}
	}
}

// getBrowDisplacement returns displacements for brow type
func getBrowDisplacement(br Brow) faceDisplacement {
	switch br {
	case BrowHeavy:
		return faceDisplacement{browRidgeZ: 0.010, browRidgeY: -0.004}
	case BrowLight:
		return faceDisplacement{browRidgeZ: -0.006, browRidgeY: 0.002}
	case BrowArched:
		return faceDisplacement{browRidgeZ: 0.004, browRidgeY: 0.006}
	default: // BrowNormal
		return faceDisplacement{}
	}
}

// faceVertexIndices holds the vertex indices for all face regions.
type faceVertexIndices struct {
	browCenter, browLeft, browRight, browOuterLeft, browOuterRight      int
	noseBridge, noseTip, noseLeft, noseRight                            int
	cheekHighLeft, cheekHighRight, cheekMidLeft, cheekMidRight          int
	cheekLowLeft, cheekLowRight                                         int
	jawCornerLeft, jawCornerRight, mandibleLeft, mandibleRight          int
	chinCenter, chinLeft, chinRight, mouthLeft, mouthRight, mouthCenter int
}

// faceBuilder helps build face mesh vertices.
type faceBuilder struct {
	verts         []Vertex
	headCenter    Vec3
	headRX        float32
	headRY        float32
	sx, sy, sz    float32
	surfaceOffset float32
}

// addVertex adds a face vertex at the specified relative position and normal.
func (fb *faceBuilder) addVertex(relX, relY, relZ, nx, ny, nz float32) int {
	pos := Vec3{
		fb.headCenter[0] + relX*fb.sx + fb.surfaceOffset*nx,
		fb.headCenter[1] + relY*fb.sy + fb.surfaceOffset*ny,
		fb.headCenter[2] + relZ*fb.sz + fb.surfaceOffset*nz,
	}
	fb.verts = append(fb.verts, Vertex{
		Position: pos,
		Normal:   vec3Normalize(Vec3{nx, ny, nz}),
		UV0:      Vec2{0.5 + relX/(2*fb.headRX), 0.5 + relY/(2*fb.headRY)},
		Color:    ColorGray,
	})
	return len(fb.verts) - 1
}

// buildBrowVertices creates the brow region vertices.
func (fb *faceBuilder) buildBrowVertices(disp faceDisplacement, idx *faceVertexIndices) {
	idx.browCenter = fb.addVertex(0, 0.040+disp.browRidgeY, 0.080+disp.browRidgeZ, 0, 0.3, 1)
	idx.browLeft = fb.addVertex(-0.045, 0.038+disp.browRidgeY, 0.075+disp.browRidgeZ, -0.2, 0.3, 1)
	idx.browRight = fb.addVertex(0.045, 0.038+disp.browRidgeY, 0.075+disp.browRidgeZ, 0.2, 0.3, 1)
	idx.browOuterLeft = fb.addVertex(-0.065, 0.030+disp.browRidgeY, 0.060+disp.browRidgeZ, -0.4, 0.2, 0.8)
	idx.browOuterRight = fb.addVertex(0.065, 0.030+disp.browRidgeY, 0.060+disp.browRidgeZ, 0.4, 0.2, 0.8)
}

// buildNoseVertices creates the nose region vertices.
func (fb *faceBuilder) buildNoseVertices(idx *faceVertexIndices) {
	idx.noseBridge = fb.addVertex(0, 0.020, 0.088, 0, 0.1, 1)
	idx.noseTip = fb.addVertex(0, -0.020, 0.095, 0, -0.1, 1)
	idx.noseLeft = fb.addVertex(-0.015, -0.025, 0.085, -0.3, -0.1, 0.9)
	idx.noseRight = fb.addVertex(0.015, -0.025, 0.085, 0.3, -0.1, 0.9)
}

// buildCheekVertices creates the cheekbone region vertices.
func (fb *faceBuilder) buildCheekVertices(disp faceDisplacement, idx *faceVertexIndices) {
	idx.cheekHighLeft = fb.addVertex(-0.070+disp.cheekX, 0.005, 0.050, -0.7, 0.1, 0.6)
	idx.cheekHighRight = fb.addVertex(0.070+disp.cheekX, 0.005, 0.050, 0.7, 0.1, 0.6)
	idx.cheekMidLeft = fb.addVertex(-0.078+disp.cheekX, -0.020, 0.035, -0.8, -0.1, 0.5)
	idx.cheekMidRight = fb.addVertex(0.078+disp.cheekX, -0.020, 0.035, 0.8, -0.1, 0.5)
	idx.cheekLowLeft = fb.addVertex(-0.065+disp.cheekX*0.5, -0.045, 0.045, -0.6, -0.2, 0.6)
	idx.cheekLowRight = fb.addVertex(0.065+disp.cheekX*0.5, -0.045, 0.045, 0.6, -0.2, 0.6)
}

// buildJawAndMouthVertices creates the jaw and mouth region vertices.
func (fb *faceBuilder) buildJawAndMouthVertices(disp faceDisplacement, idx *faceVertexIndices) {
	idx.jawCornerLeft = fb.addVertex(-0.072+disp.jawCornerX, -0.055, 0.020, -0.8, -0.3, 0.3)
	idx.jawCornerRight = fb.addVertex(0.072+disp.jawCornerX, -0.055, 0.020, 0.8, -0.3, 0.3)
	idx.mandibleLeft = fb.addVertex(-0.050+disp.jawCornerX*0.6, -0.075, 0.040+disp.chinZ*0.5, -0.5, -0.4, 0.6)
	idx.mandibleRight = fb.addVertex(0.050+disp.jawCornerX*0.6, -0.075, 0.040+disp.chinZ*0.5, 0.5, -0.4, 0.6)
	idx.chinCenter = fb.addVertex(0, -0.090+disp.chinY, 0.065+disp.chinZ, 0, -0.5, 0.8)
	idx.chinLeft = fb.addVertex(-0.025, -0.085+disp.chinY, 0.060+disp.chinZ, -0.2, -0.5, 0.8)
	idx.chinRight = fb.addVertex(0.025, -0.085+disp.chinY, 0.060+disp.chinZ, 0.2, -0.5, 0.8)
	idx.mouthLeft = fb.addVertex(-0.030, -0.050, 0.075, -0.2, -0.2, 0.9)
	idx.mouthRight = fb.addVertex(0.030, -0.050, 0.075, 0.2, -0.2, 0.9)
	idx.mouthCenter = fb.addVertex(0, -0.055, 0.080, 0, -0.2, 1)
}

// buildFaceTriangles creates the face triangle topology from vertex indices.
func buildFaceTriangles(idx faceVertexIndices) []uint32 {
	triangles := [][3]int{
		// Brow region
		{idx.browCenter, idx.browLeft, idx.noseBridge},
		{idx.browCenter, idx.noseBridge, idx.browRight},
		{idx.browLeft, idx.browOuterLeft, idx.cheekHighLeft},
		{idx.browLeft, idx.cheekHighLeft, idx.noseBridge},
		{idx.browRight, idx.noseBridge, idx.cheekHighRight},
		{idx.browRight, idx.cheekHighRight, idx.browOuterRight},
		// Nose region
		{idx.noseBridge, idx.cheekHighLeft, idx.noseLeft},
		{idx.noseBridge, idx.noseLeft, idx.noseTip},
		{idx.noseBridge, idx.noseTip, idx.noseRight},
		{idx.noseBridge, idx.noseRight, idx.cheekHighRight},
		// Cheek region
		{idx.cheekHighLeft, idx.cheekMidLeft, idx.noseLeft},
		{idx.noseLeft, idx.cheekMidLeft, idx.cheekLowLeft},
		{idx.cheekHighRight, idx.noseRight, idx.cheekMidRight},
		{idx.noseRight, idx.cheekLowRight, idx.cheekMidRight},
		// Mouth area
		{idx.noseLeft, idx.cheekLowLeft, idx.mouthLeft},
		{idx.noseLeft, idx.mouthLeft, idx.noseTip},
		{idx.noseTip, idx.mouthLeft, idx.mouthCenter},
		{idx.noseTip, idx.mouthCenter, idx.mouthRight},
		{idx.noseTip, idx.mouthRight, idx.noseRight},
		{idx.noseRight, idx.mouthRight, idx.cheekLowRight},
		// Jaw region
		{idx.cheekMidLeft, idx.jawCornerLeft, idx.cheekLowLeft},
		{idx.cheekLowLeft, idx.jawCornerLeft, idx.mandibleLeft},
		{idx.cheekLowLeft, idx.mandibleLeft, idx.mouthLeft},
		{idx.mouthLeft, idx.mandibleLeft, idx.chinLeft},
		{idx.mouthLeft, idx.chinLeft, idx.mouthCenter},
		{idx.mouthCenter, idx.chinLeft, idx.chinCenter},
		{idx.mouthCenter, idx.chinCenter, idx.chinRight},
		{idx.mouthCenter, idx.chinRight, idx.mouthRight},
		{idx.mouthRight, idx.chinRight, idx.mandibleRight},
		{idx.mouthRight, idx.mandibleRight, idx.cheekLowRight},
		{idx.cheekLowRight, idx.mandibleRight, idx.jawCornerRight},
		{idx.cheekLowRight, idx.jawCornerRight, idx.cheekMidRight},
	}

	idxs := make([]uint32, 0, len(triangles)*3)
	for _, tri := range triangles {
		idxs = append(idxs, uint32(tri[0]), uint32(tri[1]), uint32(tri[2]))
	}
	return idxs
}

// combineFaceDisplacements merges displacements from face shape, jaw, and brow.
func combineFaceDisplacements(fs FaceShape, j Jaw, br Brow) faceDisplacement {
	fsDisp := getFaceShapeDisplacement(fs)
	jawDisp := getJawDisplacement(j)
	browDisp := getBrowDisplacement(br)

	return faceDisplacement{
		chinZ:      fsDisp.chinZ + jawDisp.chinZ,
		chinY:      fsDisp.chinY + jawDisp.chinY,
		jawCornerX: fsDisp.jawCornerX + jawDisp.jawCornerX,
		browRidgeZ: fsDisp.browRidgeZ + browDisp.browRidgeZ,
		browRidgeY: fsDisp.browRidgeY + browDisp.browRidgeY,
		cheekX:     fsDisp.cheekX,
	}
}

// generateFaceMesh creates simplified face geometry overlaid on the head.
// The face mesh provides distinct regions for jaw, brow, and cheeks that
// respond to FaceShape, Jaw, and Brow parameters.
func generateFaceMesh(
	headCenter Vec3,
	headRX, headRY, headRZ float32,
	fs FaceShape,
	j Jaw,
	br Brow,
) ([]Vertex, []uint32) {
	disp := combineFaceDisplacements(fs, j, br)

	fb := &faceBuilder{
		verts:         make([]Vertex, 0, 32),
		headCenter:    headCenter,
		headRX:        headRX,
		headRY:        headRY,
		sx:            headRX / 0.090,
		sy:            headRY / 0.115,
		sz:            headRZ / 0.090,
		surfaceOffset: 0.003,
	}

	var idx faceVertexIndices
	fb.buildBrowVertices(disp, &idx)
	fb.buildNoseVertices(&idx)
	fb.buildCheekVertices(disp, &idx)
	fb.buildJawAndMouthVertices(disp, &idx)

	return fb.verts, buildFaceTriangles(idx)
}
