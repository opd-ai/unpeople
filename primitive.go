package unpeople

import "math"

const tau = math.Pi * 2.0

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
		sinT := float32(math.Sin(float64(theta)))
		cosT := float32(math.Cos(float64(theta)))

		for lon := 0; lon <= lonSegs; lon++ {
			phi := float32(lon) * float32(tau) / float32(lonSegs)
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
			uv := Vec2{float32(lon) / float32(lonSegs), float32(lat) / float32(latSegs)}

			verts = append(verts, Vertex{
				Position: pos,
				Normal:   n,
				UV0:      uv,
				Color:    ColorGray,
				Tangent:  Vec4{-sinP, 0, cosP, 1},
			})
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

// generateEar creates a simplified ear mesh as a curved, tapered shell.
// The ear is generated at the attachment point and oriented outward from the
// head. isLeftEar determines which side the ear faces (left = -X, right = +X).
// scale controls the overall size of the ear.
func generateEar(attachPoint Vec3, scale float32, isLeftEar bool) ([]Vertex, []uint32) {
	// Ear dimensions relative to scale
	earHeight := scale * 0.28   // Vertical extent
	earWidth := scale * 0.12    // Lateral protrusion from head
	earDepth := scale * 0.08    // Front-to-back thickness
	earCurve := scale * 0.04    // Curvature of the outer rim
	earTipRatio := float32(0.3) // Top tapers to 30% of base width

	// Direction multiplier: -1 for left ear, +1 for right ear
	dir := float32(1.0)
	if isLeftEar {
		dir = -1.0
	}

	// Generate a simplified ear shape as a curved quad strip
	// 4 vertical segments × 2 columns (inner/outer) = 8 vertices
	const segs = 4
	verts := make([]Vertex, 0, (segs+1)*2)
	idxs := make([]uint32, 0, segs*6)

	// Outer normal direction (pointing away from head)
	outerNorm := Vec3{dir, 0, 0}

	for i := 0; i <= segs; i++ {
		t := float32(i) / float32(segs)

		// Vertical position from bottom to top
		y := attachPoint[1] - earHeight*0.4 + earHeight*t

		// Taper width toward the top
		widthMult := 1.0 - t*(1.0-earTipRatio)

		// Curvature: ear bulges out in the middle vertically
		curveOffset := earCurve * float32(math.Sin(float64(t*math.Pi)))

		// Inner edge (closer to head)
		innerX := attachPoint[0] + dir*0.002
		innerZ := attachPoint[2] + earDepth*0.5*widthMult - curveOffset*0.5

		// Outer edge (away from head)
		outerX := attachPoint[0] + dir*(earWidth*widthMult+curveOffset)
		outerZ := attachPoint[2] - earDepth*0.5*widthMult

		// UV coordinates
		uInner := float32(0.0)
		uOuter := float32(1.0)
		v := t

		verts = append(verts, Vertex{
			Position: Vec3{innerX, y, innerZ},
			Normal:   outerNorm,
			UV0:      Vec2{uInner, v},
			Color:    ColorGray,
		})
		verts = append(verts, Vertex{
			Position: Vec3{outerX, y, outerZ},
			Normal:   outerNorm,
			UV0:      Vec2{uOuter, v},
			Color:    ColorGray,
		})
	}

	// Generate quad strip indices
	for i := 0; i < segs; i++ {
		base := uint32(i * 2)
		// Two triangles per quad, maintaining CCW winding from outside
		if isLeftEar {
			// Left ear faces -X, so triangles wind differently
			idxs = append(idxs, base, base+2, base+1)
			idxs = append(idxs, base+1, base+2, base+3)
		} else {
			// Right ear faces +X
			idxs = append(idxs, base, base+1, base+2)
			idxs = append(idxs, base+1, base+3, base+2)
		}
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

// generateHand creates a complete hand with palm box and five fingers.
// handCenter is the palm center, hw/hh/hd are palm half-extents, direction
// is the finger pointing direction (typically down for T-pose), isLeftHand
// determines finger placement mirroring, and remaining params define fingers.
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

	// Palm box
	v, idx := generateBox(handCenter, hw, hh, hd)
	builder.append(v, idx)

	// Finger segments scaled by length multiplier
	fingerSegs := []float32{
		proximalLen * fingerLengthMult,
		middleLen * fingerLengthMult,
		distalLen * fingerLengthMult,
	}
	thumbSegs := []float32{
		thumbProximalLen * fingerLengthMult,
		thumbDistalLen * fingerLengthMult,
	}

	dir := vec3Normalize(direction)

	// Compute palm edge where fingers attach (bottom of hand box)
	palmEdgeY := handCenter[1] - hh

	// Finger attachment positions (4 fingers + thumb)
	// Fingers spread laterally across the palm width
	fingerBaseZ := handCenter[2]

	// Direction multiplier for left/right hand (X-axis)
	xDir := float32(1.0)
	if isLeftHand {
		xDir = -1.0
	}

	// Four main fingers: index, middle, ring, pinky
	fingerXOffsets := []float32{
		hw * 0.55,  // Index (closer to center)
		hw * 0.18,  // Middle (center)
		-hw * 0.18, // Ring
		-hw * 0.55, // Pinky (outer edge)
	}

	for i, xOff := range fingerXOffsets {
		// Pinky is shorter
		scale := float32(1.0)
		if i == 3 {
			scale = 0.75
		}
		scaledSegs := make([]float32, len(fingerSegs))
		for j, s := range fingerSegs {
			scaledSegs[j] = s * scale
		}

		fingerBase := Vec3{
			handCenter[0] + xOff*xDir,
			palmEdgeY,
			fingerBaseZ,
		}
		v, idx := generateFinger(fingerBase, dir, scaledSegs, fingerRadius)
		builder.append(v, idx)
	}

	// Thumb: attaches to side of palm, points more laterally
	thumbBaseX := handCenter[0] + hw*0.95*xDir
	thumbBaseY := handCenter[1] - hh*0.3
	thumbDir := vec3Normalize(Vec3{xDir * 0.5, -0.85, 0.2})
	thumbBase := Vec3{thumbBaseX, thumbBaseY, handCenter[2] + hd*0.3}
	v, idx = generateFinger(thumbBase, thumbDir, thumbSegs, fingerRadius*1.15)
	builder.append(v, idx)

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
		sinT := float32(math.Sin(float64(theta)))
		cosT := float32(math.Cos(float64(theta)))

		for lon := 0; lon <= lonSegs; lon++ {
			phi := float32(lon) * float32(tau) / float32(lonSegs)
			sinP := float32(math.Sin(float64(phi)))
			cosP := float32(math.Cos(float64(phi)))

			// Unit-sphere direction
			nx := cosP * sinT
			ny := cosT
			nz := sinP * sinT

			pos := Vec3{
				headCenter[0] + capRX*nx,
				headCenter[1] + capRY*ny,
				headCenter[2] + capRZ*nz,
			}
			n := vec3Normalize(Vec3{nx / (capRX * capRX), ny / (capRY * capRY), nz / (capRZ * capRZ)})
			uv := Vec2{float32(lon) / float32(lonSegs), t}

			verts = append(verts, Vertex{
				Position: pos,
				Normal:   n,
				UV0:      uv,
				Color:    ColorGray,
				Tangent:  Vec4{-sinP, 0, cosP, 1},
			})
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
	// Combine displacements from all parameters
	fsDisp := getFaceShapeDisplacement(fs)
	jawDisp := getJawDisplacement(j)
	browDisp := getBrowDisplacement(br)

	disp := faceDisplacement{
		chinZ:      fsDisp.chinZ + jawDisp.chinZ,
		chinY:      fsDisp.chinY + jawDisp.chinY,
		jawCornerX: fsDisp.jawCornerX + jawDisp.jawCornerX,
		browRidgeZ: fsDisp.browRidgeZ + browDisp.browRidgeZ,
		browRidgeY: fsDisp.browRidgeY + browDisp.browRidgeY,
		cheekX:     fsDisp.cheekX,
	}

	// Scale factors for head dimensions
	sx := headRX / 0.090 // Scale relative to default head
	sy := headRY / 0.115
	sz := headRZ / 0.090

	// Offset to sit slightly above head surface
	const surfaceOffset = 0.003

	// Build face vertices organized by region
	verts := make([]Vertex, 0, 32)

	// Helper to add a face vertex
	addVertex := func(relX, relY, relZ, nx, ny, nz float32) int {
		pos := Vec3{
			headCenter[0] + relX*sx + surfaceOffset*nx,
			headCenter[1] + relY*sy + surfaceOffset*ny,
			headCenter[2] + relZ*sz + surfaceOffset*nz,
		}
		verts = append(verts, Vertex{
			Position: pos,
			Normal:   vec3Normalize(Vec3{nx, ny, nz}),
			UV0:      Vec2{0.5 + relX/(2*headRX), 0.5 + relY/(2*headRY)},
			Color:    ColorGray,
		})
		return len(verts) - 1
	}

	// Brow region (upper face)
	browCenterIdx := addVertex(0, 0.040+disp.browRidgeY, 0.080+disp.browRidgeZ, 0, 0.3, 1)
	browLeftIdx := addVertex(-0.045, 0.038+disp.browRidgeY, 0.075+disp.browRidgeZ, -0.2, 0.3, 1)
	browRightIdx := addVertex(0.045, 0.038+disp.browRidgeY, 0.075+disp.browRidgeZ, 0.2, 0.3, 1)
	browOuterLeftIdx := addVertex(-0.065, 0.030+disp.browRidgeY, 0.060+disp.browRidgeZ, -0.4, 0.2, 0.8)
	browOuterRightIdx := addVertex(0.065, 0.030+disp.browRidgeY, 0.060+disp.browRidgeZ, 0.4, 0.2, 0.8)

	// Nose region
	noseBridgeIdx := addVertex(0, 0.020, 0.088, 0, 0.1, 1)
	noseTipIdx := addVertex(0, -0.020, 0.095, 0, -0.1, 1)
	noseLeftIdx := addVertex(-0.015, -0.025, 0.085, -0.3, -0.1, 0.9)
	noseRightIdx := addVertex(0.015, -0.025, 0.085, 0.3, -0.1, 0.9)

	// Cheekbone region
	cheekHighLeftIdx := addVertex(-0.070+disp.cheekX, 0.005, 0.050, -0.7, 0.1, 0.6)
	cheekHighRightIdx := addVertex(0.070+disp.cheekX, 0.005, 0.050, 0.7, 0.1, 0.6)
	cheekMidLeftIdx := addVertex(-0.078+disp.cheekX, -0.020, 0.035, -0.8, -0.1, 0.5)
	cheekMidRightIdx := addVertex(0.078+disp.cheekX, -0.020, 0.035, 0.8, -0.1, 0.5)
	cheekLowLeftIdx := addVertex(-0.065+disp.cheekX*0.5, -0.045, 0.045, -0.6, -0.2, 0.6)
	cheekLowRightIdx := addVertex(0.065+disp.cheekX*0.5, -0.045, 0.045, 0.6, -0.2, 0.6)

	// Jaw region
	jawCornerLeftIdx := addVertex(-0.072+disp.jawCornerX, -0.055, 0.020, -0.8, -0.3, 0.3)
	jawCornerRightIdx := addVertex(0.072+disp.jawCornerX, -0.055, 0.020, 0.8, -0.3, 0.3)
	mandibleLeftIdx := addVertex(-0.050+disp.jawCornerX*0.6, -0.075, 0.040+disp.chinZ*0.5, -0.5, -0.4, 0.6)
	mandibleRightIdx := addVertex(0.050+disp.jawCornerX*0.6, -0.075, 0.040+disp.chinZ*0.5, 0.5, -0.4, 0.6)
	chinCenterIdx := addVertex(0, -0.090+disp.chinY, 0.065+disp.chinZ, 0, -0.5, 0.8)
	chinLeftIdx := addVertex(-0.025, -0.085+disp.chinY, 0.060+disp.chinZ, -0.2, -0.5, 0.8)
	chinRightIdx := addVertex(0.025, -0.085+disp.chinY, 0.060+disp.chinZ, 0.2, -0.5, 0.8)

	// Mouth area (connects nose to jaw)
	mouthLeftIdx := addVertex(-0.030, -0.050, 0.075, -0.2, -0.2, 0.9)
	mouthRightIdx := addVertex(0.030, -0.050, 0.075, 0.2, -0.2, 0.9)
	mouthCenterIdx := addVertex(0, -0.055, 0.080, 0, -0.2, 1)

	// Build triangles connecting the face regions (CCW winding)
	idxs := make([]uint32, 0, 96)

	addTri := func(a, b, c int) {
		idxs = append(idxs, uint32(a), uint32(b), uint32(c))
	}

	// Brow region triangles
	addTri(browCenterIdx, browLeftIdx, noseBridgeIdx)
	addTri(browCenterIdx, noseBridgeIdx, browRightIdx)
	addTri(browLeftIdx, browOuterLeftIdx, cheekHighLeftIdx)
	addTri(browLeftIdx, cheekHighLeftIdx, noseBridgeIdx)
	addTri(browRightIdx, noseBridgeIdx, cheekHighRightIdx)
	addTri(browRightIdx, cheekHighRightIdx, browOuterRightIdx)

	// Nose region triangles
	addTri(noseBridgeIdx, cheekHighLeftIdx, noseLeftIdx)
	addTri(noseBridgeIdx, noseLeftIdx, noseTipIdx)
	addTri(noseBridgeIdx, noseTipIdx, noseRightIdx)
	addTri(noseBridgeIdx, noseRightIdx, cheekHighRightIdx)

	// Cheek region triangles
	addTri(cheekHighLeftIdx, cheekMidLeftIdx, noseLeftIdx)
	addTri(noseLeftIdx, cheekMidLeftIdx, cheekLowLeftIdx)
	addTri(cheekHighRightIdx, noseRightIdx, cheekMidRightIdx)
	addTri(noseRightIdx, cheekLowRightIdx, cheekMidRightIdx)

	// Mouth area triangles
	addTri(noseLeftIdx, cheekLowLeftIdx, mouthLeftIdx)
	addTri(noseLeftIdx, mouthLeftIdx, noseTipIdx)
	addTri(noseTipIdx, mouthLeftIdx, mouthCenterIdx)
	addTri(noseTipIdx, mouthCenterIdx, mouthRightIdx)
	addTri(noseTipIdx, mouthRightIdx, noseRightIdx)
	addTri(noseRightIdx, mouthRightIdx, cheekLowRightIdx)

	// Jaw region triangles
	addTri(cheekMidLeftIdx, jawCornerLeftIdx, cheekLowLeftIdx)
	addTri(cheekLowLeftIdx, jawCornerLeftIdx, mandibleLeftIdx)
	addTri(cheekLowLeftIdx, mandibleLeftIdx, mouthLeftIdx)
	addTri(mouthLeftIdx, mandibleLeftIdx, chinLeftIdx)
	addTri(mouthLeftIdx, chinLeftIdx, mouthCenterIdx)
	addTri(mouthCenterIdx, chinLeftIdx, chinCenterIdx)
	addTri(mouthCenterIdx, chinCenterIdx, chinRightIdx)
	addTri(mouthCenterIdx, chinRightIdx, mouthRightIdx)
	addTri(mouthRightIdx, chinRightIdx, mandibleRightIdx)
	addTri(mouthRightIdx, mandibleRightIdx, cheekLowRightIdx)
	addTri(cheekLowRightIdx, mandibleRightIdx, jawCornerRightIdx)
	addTri(cheekLowRightIdx, jawCornerRightIdx, cheekMidRightIdx)

	return verts, idxs
}
