// Package unpeople provides deterministic procedural generation of humanoid meshes.
//
// This file implements Level of Detail (LOD) mesh generation using edge
// collapse decimation to reduce triangle count while preserving shape.

package unpeople

import (
	"math"
	"sort"
)

// LODLevel represents a discrete level of detail for mesh rendering.
type LODLevel int

const (
	// LOD0 is the highest detail level (100% triangles)
	LOD0 LODLevel = iota
	// LOD1 is medium detail (50% triangles)
	LOD1
	// LOD2 is lowest detail (25% triangles)
	LOD2
	// LODCount is the total number of LOD levels
	LODCount
)

// LODConfig specifies how to generate each LOD level.
type LODConfig struct {
	// TargetRatio is the target triangle ratio (0.0-1.0) relative to LOD0
	TargetRatio float32
}

// DefaultLODConfigs returns the default configuration for each LOD level.
func DefaultLODConfigs() [LODCount]LODConfig {
	return [LODCount]LODConfig{
		LOD0: {TargetRatio: 1.0},  // Full detail
		LOD1: {TargetRatio: 0.5},  // 50%
		LOD2: {TargetRatio: 0.25}, // 25%
	}
}

// LODMesh represents a mesh at a specific level of detail.
type LODMesh struct {
	Level         LODLevel
	Mesh          *Mesh
	TriangleCount int
	TriangleRatio float32 // Actual ratio compared to LOD0
}

// LODSet contains all LOD variants for a character mesh.
type LODSet struct {
	Params  Params
	Meshes  [LODCount]*LODMesh
	LOD0Key string // Cache key for LOD0
}

// LODResult contains the complete LOD generation result.
type LODResult struct {
	LODSet *LODSet
}

// GenerateWithLOD generates a character mesh at multiple levels of detail.
// Returns LOD0 (full detail), LOD1 (50%), and LOD2 (25%) variants.
func (g *Generator) GenerateWithLOD(p Params) (*LODResult, error) {
	return g.GenerateWithLODConfig(p, DefaultLODConfigs())
}

// GenerateWithLODConfig generates LOD meshes with custom configuration.
func (g *Generator) GenerateWithLODConfig(p Params, configs [LODCount]LODConfig) (*LODResult, error) {
	if err := p.Validate(); err != nil {
		return nil, err
	}

	result := &LODResult{
		LODSet: &LODSet{
			Params: p,
		},
	}

	// Generate LOD0 first (full detail) - this uses standard generation
	mesh0, err := g.Generate(p)
	if err != nil {
		return nil, err
	}

	lod0Triangles := len(mesh0.Indices) / 3
	result.LODSet.Meshes[LOD0] = &LODMesh{
		Level:         LOD0,
		Mesh:          mesh0,
		TriangleCount: lod0Triangles,
		TriangleRatio: 1.0,
	}
	result.LODSet.LOD0Key = mesh0.Key

	// Generate lower LOD levels using decimation
	for level := LOD1; level < LODCount; level++ {
		cfg := configs[level]
		targetTris := int(float32(lod0Triangles) * cfg.TargetRatio)

		// Decimate from LOD0 mesh
		lodMesh := decimateMesh(mesh0, targetTris)
		lodMesh.Key = mesh0.Key + lodSuffix(level)

		triCount := len(lodMesh.Indices) / 3
		result.LODSet.Meshes[level] = &LODMesh{
			Level:         level,
			Mesh:          lodMesh,
			TriangleCount: triCount,
			TriangleRatio: float32(triCount) / float32(lod0Triangles),
		}
	}

	return result, nil
}

// lodSuffix returns a cache key suffix for the LOD level.
func lodSuffix(level LODLevel) string {
	switch level {
	case LOD1:
		return "_lod1"
	case LOD2:
		return "_lod2"
	default:
		return ""
	}
}

// decimateMesh reduces triangle count using edge collapse decimation.
func decimateMesh(src *Mesh, targetTris int) *Mesh {
	if len(src.Indices)/3 <= targetTris {
		return copyMesh(src)
	}

	// Build initial mesh structures
	vertices := make([]Vertex, len(src.Vertices))
	copy(vertices, src.Vertices)

	indices := make([]uint32, len(src.Indices))
	copy(indices, src.Indices)

	// Track which vertices are still valid
	vertexValid := make([]bool, len(vertices))
	for i := range vertexValid {
		vertexValid[i] = true
	}

	// Vertex remapping (original index -> new index after collapses)
	remap := make([]uint32, len(vertices))
	for i := range remap {
		remap[i] = uint32(i)
	}

	// Build edge list with collapse costs
	edges := buildEdgeList(vertices, indices)
	sortEdgesByCost(edges)

	// Collapse edges until we reach target triangle count
	currentTris := len(indices) / 3
	edgeIdx := 0

	for currentTris > targetTris && edgeIdx < len(edges) {
		e := edges[edgeIdx]
		edgeIdx++

		// Skip invalid edges (endpoints already collapsed)
		v0 := followRemap(remap, e.v0)
		v1 := followRemap(remap, e.v1)
		if v0 == v1 {
			continue
		}

		// Collapse v1 into v0
		remap[v1] = v0

		// Blend vertex attributes at midpoint
		blendVertex(&vertices[v0], &vertices[v1], 0.5)

		// Count triangles removed (degenerate after collapse)
		removed := 0
		for i := 0; i < len(indices); i += 3 {
			a := followRemap(remap, indices[i])
			b := followRemap(remap, indices[i+1])
			c := followRemap(remap, indices[i+2])
			if a == b || b == c || c == a {
				removed++
			}
		}
		currentTris = len(indices)/3 - removed
	}

	// Rebuild mesh with collapsed vertices
	return rebuildMesh(vertices, indices, remap)
}

// edge represents a mesh edge with its collapse cost.
type edge struct {
	v0, v1 uint32
	cost   float32
}

// buildEdgeList extracts unique edges from the mesh.
func buildEdgeList(vertices []Vertex, indices []uint32) []edge {
	seen := make(map[uint64]bool)
	var edges []edge

	for i := 0; i < len(indices); i += 3 {
		pairs := [][2]uint32{
			{indices[i], indices[i+1]},
			{indices[i+1], indices[i+2]},
			{indices[i+2], indices[i]},
		}
		for _, p := range pairs {
			v0, v1 := p[0], p[1]
			if v0 > v1 {
				v0, v1 = v1, v0
			}
			key := uint64(v0)<<32 | uint64(v1)
			if !seen[key] {
				seen[key] = true
				cost := edgeCollapseCost(vertices, v0, v1)
				edges = append(edges, edge{v0: v0, v1: v1, cost: cost})
			}
		}
	}
	return edges
}

// edgeCollapseCost computes the cost of collapsing an edge.
// Lower cost edges are collapsed first.
func edgeCollapseCost(vertices []Vertex, v0, v1 uint32) float32 {
	p0 := vertices[v0].Position
	p1 := vertices[v1].Position

	// Edge length component
	dx := p1[0] - p0[0]
	dy := p1[1] - p0[1]
	dz := p1[2] - p0[2]
	length := float32(math.Sqrt(float64(dx*dx + dy*dy + dz*dz)))

	// Normal preservation component (penalize collapsing edges between different normals)
	n0 := vertices[v0].Normal
	n1 := vertices[v1].Normal
	normalDot := n0[0]*n1[0] + n0[1]*n1[1] + n0[2]*n1[2]
	normalCost := 1.0 - normalDot // 0 for parallel normals, 2 for opposite

	return length * (1.0 + float32(normalCost)*0.5)
}

// sortEdgesByCost sorts edges by collapse cost (lowest first).
func sortEdgesByCost(edges []edge) {
	sort.Slice(edges, func(i, j int) bool {
		return edges[i].cost < edges[j].cost
	})
}

// followRemap follows the remap chain to find the final vertex index.
func followRemap(remap []uint32, idx uint32) uint32 {
	for remap[idx] != idx {
		idx = remap[idx]
	}
	return idx
}

// blendVertex blends vertex attributes at the given ratio.
func blendVertex(dst, src *Vertex, t float32) {
	s := 1.0 - t

	dst.Position[0] = dst.Position[0]*s + src.Position[0]*t
	dst.Position[1] = dst.Position[1]*s + src.Position[1]*t
	dst.Position[2] = dst.Position[2]*s + src.Position[2]*t

	// Renormalize normal
	nx := dst.Normal[0]*s + src.Normal[0]*t
	ny := dst.Normal[1]*s + src.Normal[1]*t
	nz := dst.Normal[2]*s + src.Normal[2]*t
	nl := float32(math.Sqrt(float64(nx*nx + ny*ny + nz*nz)))
	if nl > 0.0001 {
		dst.Normal[0] = nx / nl
		dst.Normal[1] = ny / nl
		dst.Normal[2] = nz / nl
	}

	dst.UV0[0] = dst.UV0[0]*s + src.UV0[0]*t
	dst.UV0[1] = dst.UV0[1]*s + src.UV0[1]*t
}

// rebuildMesh creates a new mesh with collapsed vertices and degenerate triangles removed.
func rebuildMesh(vertices []Vertex, indices, remap []uint32) *Mesh {
	// Find which vertices are still used
	used := make([]bool, len(vertices))
	for i := range indices {
		idx := followRemap(remap, indices[i])
		used[idx] = true
	}

	// Build compacted vertex list
	newIdx := make([]uint32, len(vertices))
	var newVerts []Vertex
	for i, v := range vertices {
		if used[i] && remap[i] == uint32(i) {
			newIdx[i] = uint32(len(newVerts))
			newVerts = append(newVerts, v)
		}
	}

	// Rebuild indices, skipping degenerate triangles
	var newIndices []uint32
	for i := 0; i < len(indices); i += 3 {
		a := newIdx[followRemap(remap, indices[i])]
		b := newIdx[followRemap(remap, indices[i+1])]
		c := newIdx[followRemap(remap, indices[i+2])]

		if a != b && b != c && c != a {
			newIndices = append(newIndices, a, b, c)
		}
	}

	return &Mesh{
		Vertices: newVerts,
		Indices:  newIndices,
	}
}

// copyMesh creates a deep copy of a mesh.
func copyMesh(src *Mesh) *Mesh {
	dst := &Mesh{
		Key:      src.Key,
		Vertices: make([]Vertex, len(src.Vertices)),
		Indices:  make([]uint32, len(src.Indices)),
	}
	copy(dst.Vertices, src.Vertices)
	copy(dst.Indices, src.Indices)
	return dst
}

// GetLOD returns the mesh for the requested detail level.
func (ls *LODSet) GetLOD(level LODLevel) *Mesh {
	if level < 0 || level >= LODCount {
		return ls.Meshes[LOD0].Mesh
	}
	return ls.Meshes[level].Mesh
}

// TriangleCounts returns the triangle count for each LOD level.
func (ls *LODSet) TriangleCounts() [LODCount]int {
	var counts [LODCount]int
	for i := LODLevel(0); i < LODCount; i++ {
		counts[i] = ls.Meshes[i].TriangleCount
	}
	return counts
}

// TriangleRatios returns the actual triangle ratios for each LOD level.
func (ls *LODSet) TriangleRatios() [LODCount]float32 {
	var ratios [LODCount]float32
	for i := LODLevel(0); i < LODCount; i++ {
		ratios[i] = ls.Meshes[i].TriangleRatio
	}
	return ratios
}

// SelectLOD chooses the appropriate LOD level based on distance.
func SelectLOD(distance, lod0Distance, lod1Distance float32) LODLevel {
	if distance <= lod0Distance {
		return LOD0
	}
	if distance <= lod1Distance {
		return LOD1
	}
	return LOD2
}

// DefaultLODDistances returns typical LOD transition distances in metres.
func DefaultLODDistances() (lod0, lod1 float32) {
	return 10.0, 25.0 // LOD0 within 10m, LOD1 within 25m, LOD2 beyond
}
