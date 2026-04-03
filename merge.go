package unpeople

import "math"

// ─── Vertex Merging ──────────────────────────────────────────────────────────
//
// This file implements vertex merging to eliminate visible seams between body
// parts. The algorithm uses a spatial grid to efficiently find nearby vertices
// within an epsilon threshold, then merges them with averaged normals.
//
// The merge function preserves determinism by processing vertices in index
// order and using a consistent spatial hash.

// gridKey identifies a cell in the spatial hash grid.
type gridKey struct {
	x, y, z int32
}

// spatialGrid is a grid-based spatial hash for efficient nearest-neighbor lookup.
// Cell size should be 2× epsilon for optimal coverage.
type spatialGrid struct {
	cellSize float32
	cells    map[gridKey][]int
}

// newSpatialGrid creates a spatial grid with the given cell size.
func newSpatialGrid(cellSize float32) *spatialGrid {
	return &spatialGrid{
		cellSize: cellSize,
		cells:    make(map[gridKey][]int),
	}
}

// insert adds a vertex index to the grid at the given position.
func (g *spatialGrid) insert(pos Vec3, idx int) {
	key := g.keyFor(pos)
	g.cells[key] = append(g.cells[key], idx)
}

// keyFor computes the grid cell key for a position.
func (g *spatialGrid) keyFor(pos Vec3) gridKey {
	return gridKey{
		x: int32(math.Floor(float64(pos[0] / g.cellSize))),
		y: int32(math.Floor(float64(pos[1] / g.cellSize))),
		z: int32(math.Floor(float64(pos[2] / g.cellSize))),
	}
}

// findWithinEpsilon returns all vertex indices within epsilon distance of pos.
// It checks the cell containing pos and all 26 neighboring cells.
func (g *spatialGrid) findWithinEpsilon(pos Vec3, epsilon float32, verts []Vertex) []int {
	centerKey := g.keyFor(pos)
	var result []int

	// Check 3×3×3 neighborhood (27 cells including center)
	for dx := int32(-1); dx <= 1; dx++ {
		for dy := int32(-1); dy <= 1; dy++ {
			for dz := int32(-1); dz <= 1; dz++ {
				key := gridKey{
					x: centerKey.x + dx,
					y: centerKey.y + dy,
					z: centerKey.z + dz,
				}
				result = g.appendMatchingVertices(key, pos, epsilon, verts, result)
			}
		}
	}

	return result
}

// appendMatchingVertices appends vertex indices from a grid cell that are within epsilon.
func (g *spatialGrid) appendMatchingVertices(key gridKey, pos Vec3, epsilon float32, verts []Vertex, result []int) []int {
	for _, idx := range g.cells[key] {
		if vec3Dist(verts[idx].Position, pos) < epsilon {
			result = append(result, idx)
		}
	}
	return result
}

// vec3Dist computes the Euclidean distance between two vectors.
func vec3Dist(a, b Vec3) float32 {
	dx := a[0] - b[0]
	dy := a[1] - b[1]
	dz := a[2] - b[2]
	return float32(math.Sqrt(float64(dx*dx + dy*dy + dz*dz)))
}

// initMergeState initializes the merge mapping and normal accumulation arrays.
func initMergeState(verts []Vertex) (mergeMap []int, normalSum []Vec3, normalCount []int) {
	mergeMap = make([]int, len(verts))
	for i := range mergeMap {
		mergeMap[i] = i
	}

	normalSum = make([]Vec3, len(verts))
	normalCount = make([]int, len(verts))
	for i, v := range verts {
		normalSum[i] = v.Normal
		normalCount[i] = 1
	}
	return mergeMap, normalSum, normalCount
}

// findAndMergeVertices processes vertices to find merge candidates.
func findAndMergeVertices(grid *spatialGrid, verts []Vertex, epsilon float32, mergeMap []int, normalSum []Vec3, normalCount []int) {
	for i := range verts {
		if mergeMap[i] != i {
			continue // Already merged into another vertex
		}

		nearby := grid.findWithinEpsilon(verts[i].Position, epsilon, verts)
		for _, j := range nearby {
			if j <= i || mergeMap[j] != j {
				continue // Skip self, earlier vertices, and already-merged vertices
			}

			// Merge vertex j into vertex i
			mergeMap[j] = i
			normalSum[i] = vec3Add(normalSum[i], verts[j].Normal)
			normalCount[i]++
		}
	}
}

// applyAveragedNormals normalizes accumulated normals for merged vertices.
func applyAveragedNormals(verts []Vertex, mergeMap []int, normalSum []Vec3, normalCount []int) {
	for i := range verts {
		if mergeMap[i] == i && normalCount[i] > 1 {
			verts[i].Normal = vec3Normalize(normalSum[i])
		}
	}
}

// buildCompactedVertices creates the compacted vertex list and new index mapping.
func buildCompactedVertices(verts []Vertex, mergeMap []int) ([]Vertex, []int) {
	newIdx := make([]int, len(verts))
	compacted := make([]Vertex, 0, len(verts))
	for i, v := range verts {
		if mergeMap[i] == i {
			newIdx[i] = len(compacted)
			compacted = append(compacted, v)
		} else {
			newIdx[i] = newIdx[mergeMap[i]]
		}
	}
	return compacted, newIdx
}

// remapIndices creates a new index buffer using the new vertex indices.
func remapIndices(idxs []uint32, newIdx []int) []uint32 {
	newIndices := make([]uint32, len(idxs))
	for i, oldIdx := range idxs {
		newIndices[i] = uint32(newIdx[int(oldIdx)])
	}
	return newIndices
}

// mergeVertices eliminates duplicate vertices within epsilon distance by
// merging them and averaging their normals. The function preserves determinism
// by processing vertices in index order.
//
// Returns a compacted vertex slice and remapped index buffer.
func mergeVertices(verts []Vertex, idxs []uint32, epsilon float32) ([]Vertex, []uint32) {
	if len(verts) == 0 {
		return verts, idxs
	}

	// Build spatial grid with cell size = 2×epsilon for optimal coverage
	grid := newSpatialGrid(epsilon * 2)
	for i, v := range verts {
		grid.insert(v.Position, i)
	}

	mergeMap, normalSum, normalCount := initMergeState(verts)
	findAndMergeVertices(grid, verts, epsilon, mergeMap, normalSum, normalCount)
	applyAveragedNormals(verts, mergeMap, normalSum, normalCount)

	compacted, newIdx := buildCompactedVertices(verts, mergeMap)
	newIndices := remapIndices(idxs, newIdx)

	return compacted, newIndices
}

// defaultMergeEpsilon is the base epsilon for vertex merging (2mm).
// This is scaled proportionally for species with different body sizes.
const defaultMergeEpsilon = 0.002

// scaledEpsilon computes the merge epsilon scaled to the character's total height.
func scaledEpsilon(totalHeight float32) float32 {
	return defaultMergeEpsilon * totalHeight / defaultTotalHeight
}
