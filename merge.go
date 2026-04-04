package unpeople

import (
	"fmt"
	"math"
)

// ─── Vertex Merging ──────────────────────────────────────────────────────────
//
// This file implements vertex merging to eliminate visible seams between body
// parts. The algorithm uses a spatial grid to efficiently find nearby vertices
// within an epsilon threshold, then merges them with averaged normals.
//
// The merge function preserves determinism by processing vertices in index
// order and using a consistent spatial hash.

// VertexPair represents two vertices at body part boundaries that are candidates
// for merging based on proximity.
type VertexPair struct {
	IndexA int     // Index of first vertex
	IndexB int     // Index of second vertex
	Dist   float32 // Distance between vertices
}

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
// The estimatedVertices parameter helps pre-size the internal map.
func newSpatialGrid(cellSize float32, estimatedVertices int) *spatialGrid {
	// Estimate map size: typically vertices are spread across fewer cells than vertices
	mapSize := estimatedVertices / 4
	if mapSize < 64 {
		mapSize = 64
	}
	return &spatialGrid{
		cellSize: cellSize,
		cells:    make(map[gridKey][]int, mapSize),
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

// neighborhoodOffsets contains all 27 offsets for a 3×3×3 grid search.
// Pre-computed to avoid allocations during spatial queries.
var neighborhoodOffsets = [27]gridKey{
	{-1, -1, -1},
	{-1, -1, 0},
	{-1, -1, 1},
	{-1, 0, -1},
	{-1, 0, 0},
	{-1, 0, 1},
	{-1, 1, -1},
	{-1, 1, 0},
	{-1, 1, 1},
	{0, -1, -1},
	{0, -1, 0},
	{0, -1, 1},
	{0, 0, -1},
	{0, 0, 0},
	{0, 0, 1},
	{0, 1, -1},
	{0, 1, 0},
	{0, 1, 1},
	{1, -1, -1},
	{1, -1, 0},
	{1, -1, 1},
	{1, 0, -1},
	{1, 0, 0},
	{1, 0, 1},
	{1, 1, -1},
	{1, 1, 0},
	{1, 1, 1},
}

// findWithinEpsilon returns all vertex indices within epsilon distance of pos.
// It checks the cell containing pos and all 26 neighboring cells.
func (g *spatialGrid) findWithinEpsilon(pos Vec3, epsilon float32, verts []Vertex) []int {
	centerKey := g.keyFor(pos)
	var result []int

	for _, offset := range neighborhoodOffsets {
		key := gridKey{
			x: centerKey.x + offset.x,
			y: centerKey.y + offset.y,
			z: centerKey.z + offset.z,
		}
		result = g.appendMatchingVertices(key, pos, epsilon, verts, result)
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

// FindBoundaryVertices identifies vertex pairs at body part boundaries that are
// within the given distance threshold. These are candidates for merging to
// eliminate visible seams.
//
// The function returns pairs where IndexA < IndexB to avoid duplicates, and
// pairs are sorted by IndexA for deterministic processing order.
func FindBoundaryVertices(mesh *Mesh, threshold float32) []VertexPair {
	if mesh == nil || len(mesh.Vertices) == 0 {
		return nil
	}

	// Build spatial grid for efficient neighbor lookup
	grid := newSpatialGrid(threshold*2, len(mesh.Vertices))
	for i, v := range mesh.Vertices {
		grid.insert(v.Position, i)
	}

	var pairs []VertexPair
	seen := make(map[uint64]bool)

	// Find vertex pairs within threshold distance
	for i, v := range mesh.Vertices {
		nearby := grid.findWithinEpsilon(v.Position, threshold, mesh.Vertices)
		for _, j := range nearby {
			if j <= i {
				continue // Only keep pairs where j > i
			}
			// Create unique key for the pair
			key := uint64(i)<<32 | uint64(j)
			if seen[key] {
				continue
			}
			seen[key] = true

			dist := vec3Dist(v.Position, mesh.Vertices[j].Position)
			if dist < threshold {
				pairs = append(pairs, VertexPair{
					IndexA: i,
					IndexB: j,
					Dist:   dist,
				})
			}
		}
	}

	return pairs
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

// shouldMerge checks if vertex j should be merged into vertex i.
func shouldMerge(i, j int, mergeMap []int) bool {
	return j > i && mergeMap[j] == j
}

// mergeVertexInto merges vertex j into vertex i, updating merge maps and normal accumulators.
func mergeVertexInto(i, j int, verts []Vertex, mergeMap []int, normalSum []Vec3, normalCount []int) {
	mergeMap[j] = i
	normalSum[i] = vec3Add(normalSum[i], verts[j].Normal)
	normalCount[i]++
}

// processMergeCandidates handles all merge candidates for a single vertex.
func processMergeCandidates(i int, nearby []int, verts []Vertex, mergeMap []int, normalSum []Vec3, normalCount []int) {
	for _, j := range nearby {
		if shouldMerge(i, j, mergeMap) {
			mergeVertexInto(i, j, verts, mergeMap, normalSum, normalCount)
		}
	}
}

// findAndMergeVertices processes vertices to find merge candidates.
func findAndMergeVertices(grid *spatialGrid, verts []Vertex, epsilon float32, mergeMap []int, normalSum []Vec3, normalCount []int) {
	for i := range verts {
		if mergeMap[i] != i {
			continue // Already merged into another vertex
		}

		nearby := grid.findWithinEpsilon(verts[i].Position, epsilon, verts)
		processMergeCandidates(i, nearby, verts, mergeMap, normalSum, normalCount)
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

// averageNormalsAtMergedVertices computes averaged normals for all vertices
// that were merged together. The mergedIndices map stores, for each kept vertex
// index, a list of original vertex indices that were merged into it.
//
// This ensures smooth shading across former body part boundaries.
func averageNormalsAtMergedVertices(mesh *Mesh, mergedIndices map[uint32][]uint32) {
	if mesh == nil || len(mergedIndices) == 0 {
		return
	}

	for targetIdx, sourceIndices := range mergedIndices {
		if int(targetIdx) >= len(mesh.Vertices) {
			continue
		}

		// Accumulate normals from all merged vertices
		var sumNormal Vec3
		for _, srcIdx := range sourceIndices {
			if int(srcIdx) < len(mesh.Vertices) {
				n := mesh.Vertices[srcIdx].Normal
				sumNormal[0] += n[0]
				sumNormal[1] += n[1]
				sumNormal[2] += n[2]
			}
		}

		// Add the target vertex's own normal
		n := mesh.Vertices[targetIdx].Normal
		sumNormal[0] += n[0]
		sumNormal[1] += n[1]
		sumNormal[2] += n[2]

		// Normalize and apply
		mesh.Vertices[targetIdx].Normal = vec3Normalize(sumNormal)
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
	grid := newSpatialGrid(epsilon*2, len(verts))
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

// MergeNearbyVertices consolidates boundary vertex pairs into single vertices,
// updating index buffer references and averaging normals at merged points.
// This eliminates visible seams between body parts.
//
// The threshold parameter specifies the maximum distance between vertices to
// merge. A typical value is 0.002 (2mm) for human-scale meshes.
//
// Returns a new Mesh with consolidated vertices; the original is not modified.
// Determinism is preserved: the same input always produces the same output.
func MergeNearbyVertices(mesh *Mesh, threshold float32) *Mesh {
	if mesh == nil || len(mesh.Vertices) == 0 {
		return mesh
	}

	// Copy vertices to avoid modifying the original
	verts := make([]Vertex, len(mesh.Vertices))
	copy(verts, mesh.Vertices)

	// Copy indices
	idxs := make([]uint32, len(mesh.Indices))
	copy(idxs, mesh.Indices)

	// Merge vertices
	mergedVerts, mergedIdxs := mergeVertices(verts, idxs, threshold)

	return &Mesh{
		Key:      mesh.Key + "_merged",
		Vertices: mergedVerts,
		Indices:  mergedIdxs,
	}
}

// stitchEdgeLoops generates triangles connecting two edge loops to close gaps
// at major body part connections (e.g., shoulder socket to upper arm end-cap).
//
// Each loop is a list of vertex indices that form a closed ring. The loops must
// have the same number of vertices for proper stitching.
//
// Returns an error if the loops have different lengths or contain invalid indices.
func StitchEdgeLoops(mesh *Mesh, loop1, loop2 []uint32) error {
	if mesh == nil {
		return fmt.Errorf("mesh is nil")
	}
	if len(loop1) != len(loop2) {
		return fmt.Errorf("edge loops have different lengths: %d vs %d", len(loop1), len(loop2))
	}
	if len(loop1) < 3 {
		return fmt.Errorf("edge loops must have at least 3 vertices, got %d", len(loop1))
	}

	// Validate all indices are within bounds
	vertexCount := uint32(len(mesh.Vertices))
	for i, idx := range loop1 {
		if idx >= vertexCount {
			return fmt.Errorf("loop1[%d] = %d is out of bounds (vertex count: %d)", i, idx, vertexCount)
		}
	}
	for i, idx := range loop2 {
		if idx >= vertexCount {
			return fmt.Errorf("loop2[%d] = %d is out of bounds (vertex count: %d)", i, idx, vertexCount)
		}
	}

	n := len(loop1)
	newIndices := make([]uint32, 0, n*6)

	// Generate two triangles for each pair of adjacent vertices in the loops
	for i := 0; i < n; i++ {
		next := (i + 1) % n

		// First triangle: loop1[i], loop2[i], loop1[next]
		newIndices = append(newIndices, loop1[i], loop2[i], loop1[next])

		// Second triangle: loop1[next], loop2[i], loop2[next]
		newIndices = append(newIndices, loop1[next], loop2[i], loop2[next])
	}

	// Append stitching triangles to mesh indices
	mesh.Indices = append(mesh.Indices, newIndices...)

	return nil
}
