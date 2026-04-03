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

	// Bottom ring
	bottomStart := 0
	for i := 0; i < segments; i++ {
		angle := float32(i) * float32(tau) / float32(segments)
		c := float32(math.Cos(float64(angle)))
		s := float32(math.Sin(float64(angle)))
		offset := vec3Add(vec3Scale(perp, c), vec3Scale(biperp, s))
		verts = append(verts, Vertex{
			Position: vec3Add(bottomCenter, vec3Scale(offset, bottomRadius)),
			Normal:   vec3Normalize(offset),
			UV0:      Vec2{float32(i) / float32(segments), 0},
			Color:    ColorGray,
			Tangent:  tangent,
		})
	}

	// Top ring
	topStart := segments
	for i := 0; i < segments; i++ {
		angle := float32(i) * float32(tau) / float32(segments)
		c := float32(math.Cos(float64(angle)))
		s := float32(math.Sin(float64(angle)))
		offset := vec3Add(vec3Scale(perp, c), vec3Scale(biperp, s))
		verts = append(verts, Vertex{
			Position: vec3Add(topCenter, vec3Scale(offset, topRadius)),
			Normal:   vec3Normalize(offset),
			UV0:      Vec2{float32(i) / float32(segments), 1},
			Color:    ColorGray,
			Tangent:  tangent,
		})
	}

	// Side quads
	for i := 0; i < segments; i++ {
		next := (i + 1) % segments
		b0 := uint32(bottomStart + i)
		b1 := uint32(bottomStart + next)
		t0 := uint32(topStart + i)
		t1 := uint32(topStart + next)
		idxs = append(idxs, b0, t0, b1)
		idxs = append(idxs, b1, t0, t1)
	}

	// Caps
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

	stride := uint32(lonSegs + 1)
	for lat := 0; lat < latSegs; lat++ {
		for lon := 0; lon < lonSegs; lon++ {
			a := uint32(lat)*stride + uint32(lon)
			b := a + stride
			idxs = append(idxs, a, b, a+1)
			idxs = append(idxs, b, b+1, a+1)
		}
	}

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
