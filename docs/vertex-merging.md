# Vertex Merging Algorithm Design

This document describes the algorithm design for eliminating visible seams between body parts by merging boundary vertices.

## Problem Statement

Currently, the humanoid mesh is assembled from disconnected geometric primitives (ellipsoid head, cylindrical limbs, box hands/feet). Vertices at part boundaries are not shared, resulting in:

- Visible gaps at shoulders, hips, elbows, knees, ankles, and neck
- Potential lighting artifacts at seams under certain shading models
- Winding order inconsistencies between primitive types

## Design Goals

1. **Zero visual seams** at body part boundaries
2. **Preserved determinism** — same seed produces identical merged mesh
3. **Minimal performance overhead** — merging should not significantly increase generation time
4. **Backward-compatible** — merged meshes work with existing Kaiju rendering pipeline

## Algorithm Overview

The vertex merging process runs after primitive assembly but before final mesh construction:

```
[Generate Primitives] → [Identify Boundary Vertices] → [Merge Within Epsilon] → [Rebuild Index Buffer]
```

## Approach: Spatial Lookup with Epsilon Threshold

### Rationale

We chose spatial lookup over explicit correspondence tables because:
- **Flexibility**: Works with any primitive combination
- **No hard-coded mappings**: Adding new body parts doesn't require updating correspondence tables
- **Automatic adaptation**: Handles species/build variations where boundary positions shift

### Data Structure: Grid-Based Spatial Hash

A KD-tree is overkill for our vertex counts (~2000 vertices). Instead, use a simple grid-based spatial hash:

```go
type spatialGrid struct {
    cellSize float32
    cells    map[gridKey][]int  // gridKey → vertex indices
}

type gridKey struct {
    x, y, z int32
}

func (g *spatialGrid) insert(v Vec3, idx int) {
    key := gridKey{
        x: int32(v[0] / g.cellSize),
        y: int32(v[1] / g.cellSize),
        z: int32(v[2] / g.cellSize),
    }
    g.cells[key] = append(g.cells[key], idx)
}

func (g *spatialGrid) findWithinEpsilon(v Vec3, epsilon float32) []int {
    // Check this cell and all 26 neighbors
    // Filter by actual distance
}
```

### Epsilon Threshold Strategy

The merge epsilon must be:
- Large enough to catch boundary vertices that should be shared
- Small enough to not accidentally merge distinct vertices

**Recommended**: `epsilon = 0.002` metres (2mm)

This is derived from:
- Minimum cylinder segment spacing: ~5mm between adjacent vertices at joints
- Maximum primitive placement error: ~1mm from floating-point accumulation
- Safety factor: 2× the placement error

For species with extreme scaling (Ogre at 1.5×), the epsilon scales proportionally:
```go
epsilon := 0.002 * layout.totalHeight / defaultTotalHeight
```

### Boundary Identification

A vertex is considered a **boundary vertex** if:
1. It lies on the edge of a primitive (first/last ring of cylinder, etc.)
2. Its position is near a known joint location

Known joint locations from `bodyLayout`:
- Neck: `neckTop`, `neckBottom`
- Shoulders: `upperArmTopL`, `upperArmTopR`
- Elbows: `upperArmBottomL`, `upperArmBottomR`, `forearmTopL`, `forearmTopR`
- Wrists: `forearmBottomL`, `forearmBottomR`
- Hip joints: `upperLegTopL`, `upperLegTopR`
- Knees: `upperLegBottomL`, `upperLegBottomR`, `lowerLegTopL`, `lowerLegTopR`
- Ankles: `lowerLegBottomL`, `lowerLegBottomR`

```go
func isBoundaryVertex(v Vec3, joints []Vec3, epsilon float32) bool {
    for _, j := range joints {
        if vec3Dist(v, j) < epsilon*3 {
            return true
        }
    }
    return false
}
```

## Pseudocode: Vertex Merging

```go
func mergeVertices(verts []Vertex, idxs []uint32, epsilon float32) ([]Vertex, []uint32) {
    // Build spatial grid
    grid := newSpatialGrid(epsilon * 2)
    for i, v := range verts {
        grid.insert(v.Position, i)
    }
    
    // Build merge mapping: old index → canonical index
    mergeMap := make([]int, len(verts))
    for i := range mergeMap {
        mergeMap[i] = i  // Initially maps to self
    }
    
    // Find vertices to merge
    for i, v := range verts {
        if mergeMap[i] != i {
            continue  // Already merged into another
        }
        
        // Find nearby vertices
        nearby := grid.findWithinEpsilon(v.Position, epsilon)
        for _, j := range nearby {
            if j <= i || mergeMap[j] != j {
                continue
            }
            if vec3Dist(verts[i].Position, verts[j].Position) < epsilon {
                // Merge j into i
                mergeMap[j] = i
            }
        }
    }
    
    // Build compacted vertex list
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
    
    // Remap indices
    newIndices := make([]uint32, len(idxs))
    for i, oldIdx := range idxs {
        newIndices[i] = uint32(newIdx[int(oldIdx)])
    }
    
    return compacted, newIndices
}
```

## Normal Averaging at Merge Points

When vertices are merged, their normals should be averaged to produce smooth shading:

```go
// During merge:
if vec3Dist(verts[i].Position, verts[j].Position) < epsilon {
    // Average normals
    avgNormal := vec3Normalize(vec3Add(verts[i].Normal, verts[j].Normal))
    verts[i].Normal = avgNormal
    mergeMap[j] = i
}
```

## Integration Point

The merge function should be called in `buildMesh` after all primitives are assembled but before returning:

```go
func buildMesh(layout bodyLayout, key string, hasHairSlot bool) *Mesh {
    var builder meshBuilder
    
    // ... assemble all primitives ...
    
    mesh := builder.build(key)
    
    // Merge boundary vertices
    epsilon := 0.002 * layout.totalHeight / defaultTotalHeight
    mesh.Vertices, mesh.Indices = mergeVertices(mesh.Vertices, mesh.Indices, epsilon)
    
    return mesh
}
```

## Performance Considerations

- **Expected vertex count**: ~2000-3000 vertices
- **Expected merge candidates**: ~200-400 vertices (boundary vertices)
- **Grid cell count**: ~50-100 non-empty cells
- **Time complexity**: O(n × k) where k is average vertices per cell (~20-40)
- **Estimated overhead**: <1ms additional per mesh generation

## Testing Strategy

1. **Determinism test**: Same seed → identical merged mesh
2. **Mesh validity test**: All indices within bounds after merge
3. **Seam reduction test**: Count vertices at known joint positions before/after
4. **Normal continuity test**: Check normal variance across former boundaries

## Future Enhancements

1. **Weighted normal averaging** based on incident triangle area
2. **Tangent recalculation** after merge for correct normal mapping
3. **UV coordinate stitching** across boundaries (requires Phase 3 UV atlas)

---

*Design document created: 2026-04-03*
*Prerequisite for: Phase 2 Topology Upgrade*
