# Implementation Plan: Phase 2 Enhanced Geometry

## Project Context
- **What it does**: Deterministic procedural generation of humanoid character meshes from seed + parameters for use in open-world games
- **Current goal**: Complete Phase 2 "Enhanced Geometry" features from ROADMAP.md
- **Estimated Scope**: Medium (~8 items above threshold across multiple features)

## Goal-Achievement Status
| Stated Goal | Current Status | This Plan Addresses |
|-------------|---------------|---------------------|
| Phase 1: Core Implementation | ✅ Complete | No |
| Phase 2: Topology upgrade (seams) | ❌ Missing | Yes |
| Phase 2: Advanced facial morphing | ❌ Missing | Yes |
| Phase 2: Ear geometry | ✅ Complete | No |
| Phase 2: Finger geometry | ❌ Missing | Yes |
| Phase 2: Toe geometry | ❌ Missing | Yes |
| Phase 2: Musculature detail | ❌ Blocked (needs Phase 3 UV) | No |
| Phase 2: Hair/skull cap placeholder | ❌ Missing | Yes |
| Species × Build interaction tuning | ⚠️ Untuned | Yes |

## Metrics Summary
- **Complexity hotspots on goal-critical paths**: 1 function above threshold (generateCylinder: CC=10)
- **Duplication ratio**: 0% (resolved in previous work)
- **Doc coverage**: 100% (all exports documented)
- **Package coupling**: Minimal (2 packages: `unpeople`, `unpeople/example`)

## Key Metrics (go-stats-generator)
```
Total lines: 1038
Total functions: 41 (6 methods)
Files processed: 8

Complexity leaders:
  generateCylinder (primitive.go:22) - CC=10 (ABOVE threshold 9.0)
  generateEllipsoid (primitive.go:131) - CC=5
  applyFacialFeatures (transforms.go:398) - CC=5
```

## Prerequisites from ROADMAP.md (Required Before Full Phase 2)
The ROADMAP identifies 4 technical prerequisites that must be addressed:

1. **Vertex merging algorithm** — Required for topology upgrade
2. **Face mesh vertex positions** — Required for advanced facial morphing
3. **Ear attachment coordinates** — ✅ Already implemented (`earAttachL`, `earAttachR` in bodyLayout)
4. **Finger bone hierarchy** — Required for finger geometry

---

## Implementation Steps

### Step 1: Reduce generateCylinder Complexity
- **Deliverable**: Refactor `primitive.go:generateCylinder` to extract cap generation into helper functions, reducing cyclomatic complexity from 10 to ≤8
- **Dependencies**: None
- **Goal Impact**: Code maintainability; enables cleaner extension for topology upgrade
- **Acceptance**: `go-stats-generator analyze . --skip-tests --format json | jq '.functions[] | select(.name=="generateCylinder") | .complexity.cyclomatic'` returns ≤8
- **Validation**: `go test ./... && go vet ./...`

### Step 2: Implement Finger Geometry Foundation
- **Deliverable**: 
  - Add finger generation in `primitive.go` as `generateFinger(base Vec3, direction Vec3, segments []float32, radius float32) ([]Vertex, []uint32)`
  - Update `basemesh.go:buildMesh` to generate 5 fingers per hand (proximal/middle/distal segments)
  - Use existing `fingerRadius`, `proximalLength`, `middleLength`, `distalLength` constants from `basemesh.go:75-81`
- **Dependencies**: Step 1 (for cleaner primitive generation patterns)
- **Goal Impact**: Phase 2 "Finger geometry" feature
- **Acceptance**: Visual fingers on generated mesh; `fingerLengthMult` field affects geometry
- **Validation**: `go test ./... -run TestAllFingerLengths && go vet ./...`

### Step 3: Implement Toe Geometry
- **Deliverable**:
  - Add toe segment constants to `basemesh.go` (similar to finger constants)
  - Add `generateToe` primitive or reuse `generateFinger` with toe dimensions
  - Update `basemesh.go:buildMesh` to generate 5 toes per foot
- **Dependencies**: Step 2 (reuses finger generation pattern)
- **Goal Impact**: Phase 2 "Toe geometry" feature
- **Acceptance**: Visual toes on generated mesh
- **Validation**: `go test ./... && go vet ./...`

### Step 4: Add Hair/Skull Cap Placeholder Mesh
- **Deliverable**:
  - Add `generateSkullCap(headCenter Vec3, headRX, headRY, headRZ float32) ([]Vertex, []uint32)` in `primitive.go`
  - Update `basemesh.go:buildMesh` to append skull cap mesh
  - Add `HasHairSlot bool` to `Params` (default true)
- **Dependencies**: None (can parallelize with Step 2-3)
- **Goal Impact**: Phase 2 "Hair/skull cap placeholder" feature
- **Acceptance**: Skull cap mesh present when `HasHairSlot=true`; absent when false
- **Validation**: `go test ./... -run TestMeshIsValid && go vet ./...`

### Step 5: Implement Species × Build Interaction Matrix
- **Deliverable**:
  - Add `speciesBuildInteraction(s Species, b Build) (chestMult, limbMult float32)` in `transforms.go`
  - Modify `applyBuild` to call `speciesBuildInteraction` and apply species-aware multipliers
  - Address the Orc+Fragile, Troll+Lean, etc. awkward combinations documented in GAPS.md
- **Dependencies**: None
- **Goal Impact**: Resolves "Species × Build Interaction Untuned" gap
- **Acceptance**: `SpeciesOrc + BuildFragile` produces visually reasonable proportions (manual inspection)
- **Validation**: `go test ./... -run TestAllSpecies -run TestAllBuilds && go vet ./...`

### Step 6: Document Vertex Merging Algorithm Design
- **Deliverable**:
  - Create `docs/vertex-merging.md` documenting the KD-tree spatial lookup approach vs explicit correspondence tables
  - Include pseudocode for identifying boundary vertices between adjacent body parts
  - Define epsilon threshold strategy based on `bodyLayout` dimensions
- **Dependencies**: None (documentation task)
- **Goal Impact**: ROADMAP prerequisite "Vertex merging algorithm" — enables future topology upgrade
- **Acceptance**: Design document exists with clear implementation path
- **Validation**: File exists at `docs/vertex-merging.md`

### Step 7: Design Face Mesh Template
- **Deliverable**:
  - Create `docs/face-mesh-template.md` documenting vertex group positions for jaw, brow, cheekbones, nose, chin
  - Define mapping from `FaceShape`, `Jaw`, `Brow` parameters to vertex displacement vectors
  - Include reference coordinates relative to `headCenter` and head radii
- **Dependencies**: None (documentation task)
- **Goal Impact**: ROADMAP prerequisite "Face mesh vertex positions" — enables advanced facial morphing
- **Acceptance**: Design document exists with coordinate definitions
- **Validation**: File exists at `docs/face-mesh-template.md`

### Step 8: Implement Basic Face Mesh Structure
- **Deliverable**:
  - Add `generateFaceMesh(layout bodyLayout, fs FaceShape, j Jaw, br Brow) ([]Vertex, []uint32)` in `primitive.go`
  - Generate simplified face geometry overlaid on head ellipsoid with distinct regions for jaw, brow, cheeks
  - Update `basemesh.go:buildMesh` to include face mesh
- **Dependencies**: Step 7 (face template design)
- **Goal Impact**: Phase 2 "Advanced facial morphing" (first iteration)
- **Acceptance**: Facial parameters produce visually distinguishable face shapes
- **Validation**: `go test ./... -run TestAllFaceShapes -run TestAllJaws -run TestAllBrows && go vet ./...`

---

## Deferred Items (Not In This Plan)

| Item | Reason |
|------|--------|
| Topology upgrade (seam elimination) | Requires vertex merging implementation; recommend separate Phase 2.5 plan |
| Musculature detail | Blocked by Phase 3 UV atlas (dependency chain) |
| Phase 3-6 features | Out of scope for this plan |

---

## Validation Commands

### Full Test Suite
```bash
go test ./... -v
```

### Complexity Check
```bash
go-stats-generator analyze . --skip-tests --format json --sections functions | \
  jq '[.functions[] | select(.complexity.cyclomatic > 9)] | length'
# Expected: 0
```

### Documentation Coverage
```bash
go-stats-generator analyze . --skip-tests --format json --sections documentation | \
  jq '.documentation.coverage.overall'
# Expected: 100
```

### Benchmark Performance
```bash
go test -bench=BenchmarkGenerate -benchmem
# Expected: <100ms per generation (per success criteria)
```

---

## Success Metrics

| Metric | Current | Target |
|--------|---------|--------|
| Functions above CC=9 | 1 | 0 |
| Phase 2 features complete | 1/7 | 5/7 |
| GAPS.md resolved items | 6 | 7 |
| Test coverage (enum iterations) | 100% | 100% |

---

*Plan generated: 2026-04-03 using go-stats-generator metrics analysis*
