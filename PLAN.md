# Implementation Plan: Complete Phase 1 & Prepare Phase 2

## Project Context
- **What it does**: Deterministic procedural generation of humanoid character meshes for Go game engines (Kaiju-compatible)
- **Current goal**: Complete Phase 1 gaps (test coverage, proportions fixes) and reduce technical debt before Phase 2
- **Estimated Scope**: Medium (10 items above thresholds)

## Goal-Achievement Status
| Stated Goal | Current Status | This Plan Addresses |
|-------------|---------------|---------------------|
| Deterministic generation | ✅ Achieved | No |
| Kaiju Vertex compatibility | ✅ Achieved | No |
| Zero external dependencies | ✅ Achieved | No |
| 10 Species variations | ✅ Achieved | No |
| Proportion styles (Heroic/Caricature) | ⚠️ Incomplete | Yes |
| Comprehensive test coverage | ⚠️ Incomplete | Yes |
| Maintainable codebase | ⚠️ High duplication | Yes |
| Phase 2: Enhanced Geometry | ❌ Not started | Prepares |

## Metrics Summary
- **Complexity hotspots**: 2 functions above threshold
  - `Validate()`: cyclomatic=19, overall=25.2 (threshold: 10)
  - `generateCylinder()`: cyclomatic=8, overall=11.4 (acceptable)
- **Duplication ratio**: 24.3% (439 duplicated lines, 25 clone pairs)
- **Doc coverage**: 100%
- **Package coupling**: Clean (unpeople has 0 external dependencies; example depends only on unpeople)

## Implementation Steps

### Step 1: Add Missing Enum Test Coverage ✅
- [x] **Resolved**
- **Deliverable**: Add `TestAllProportions`, `TestAllPhenotypes`, `TestAllPostures`, `TestAllFaceShapes`, `TestAllJaws`, `TestAllBrows`, `TestAllEars`, `TestAllShoulderWidths`, `TestAllHipWidths`, `TestAllLimbLengths`, `TestAllNeckLengths`, `TestAllHandSizes`, `TestAllFingerLengths`, `TestAllFootSizes` to `generator_test.go`
- **Dependencies**: None
- **Goal Impact**: Completes ROADMAP Phase 1 claim of "all enums" test coverage
- **Acceptance**: All new tests pass; `go test ./... -v` shows 24+ passing tests (was 10)
- **Validation**: `go test ./... -v 2>&1 | grep -c PASS`

### Step 2: Fix ProportionsHeroic Leg Elongation ✅
- [x] **Resolved**
- **Deliverable**: Modify `applyProportions()` in `transforms.go` to add `scaleLimbs(l, 1.08)` within `case ProportionsHeroic:`
- **Dependencies**: Step 1 (TestAllProportions must exist to verify)
- **Goal Impact**: Closes GAPS.md "Heroic Proportions Missing Leg Elongation"
- **Acceptance**: `TestAllProportions` passes; heroic characters have visually longer legs (8% increase)
- **Validation**: `go test ./... -run TestAllProportions -v`

### Step 3: Fix ProportionsCaricature Extremity Reduction ✅
- [x] **Resolved**
- **Deliverable**: Modify `applyProportions()` in `transforms.go` to add hand/foot reduction within `case ProportionsCaricature:`:
  ```go
  l.handHW *= 0.85
  l.handHH *= 0.85
  l.handHD *= 0.85
  l.footHW *= 0.85
  l.footHD *= 0.85
  ```
- **Dependencies**: Step 1 (TestAllProportions must exist to verify)
- **Goal Impact**: Closes GAPS.md "Caricature Proportions Missing Extremity Reduction"
- **Acceptance**: `TestAllProportions` passes; caricature characters have smaller hands/feet
- **Validation**: `go test ./... -run TestAllProportions -v`

### Step 4: Refactor Validate() to Table-Driven ✅
- [x] **Resolved**
- **Deliverable**: Replace 18 sequential if-statements in `params.go:Validate()` with table-driven validation using a slice of `{name, val, min, max}` structs
- **Dependencies**: None
- **Goal Impact**: Reduces cyclomatic complexity from 19 to ~3; improves maintainability
- **Acceptance**: `Validate()` complexity drops below 10; all existing tests pass
- **Validation**: `go-stats-generator analyze . --skip-tests --format json 2>/dev/null | jq -r '.functions[] | select(.name == "Validate") | .complexity.cyclomatic'` returns value ≤ 10

### Step 5: Reduce Scale Helper Duplication
- [x] **Resolved**
- **Deliverable**: Create a unified field accessor mechanism in `transforms.go` that `scaleAll`, `scaleHeight`, and `scaleLimbs` can share. Options:
  - (A) Extract common field lists to data tables
  - (B) Use a helper that accepts field pointers
- **Dependencies**: None
- **Goal Impact**: Reduces duplication ratio; closes GAPS.md "Code Duplication in Scale Helpers"
- **Acceptance**: Duplication ratio drops below 15%; all tests pass
- **Validation**: `go-stats-generator analyze . --skip-tests --format json 2>/dev/null | jq -r '.duplication.duplication_ratio'` returns value < 0.15

### Step 6: Pre-allocate Slices in Primitives
- [x] **Resolved**
- **Deliverable**: Modify `generateEllipsoid`, `generateCylinder`, `generateBox` in `primitive.go` to pre-allocate vertex/index slices using `make([]Vertex, 0, capacity)`
- **Dependencies**: None
- **Goal Impact**: Addresses performance anti-patterns flagged by go-stats-generator; reduces GC pressure
- **Acceptance**: All tests pass; BenchmarkGenerate shows reduced allocations
- **Validation**: `go test -bench=BenchmarkGenerate -benchmem ./... 2>&1 | grep allocs`

### Step 7: Pre-allocate Slice in meshBuilder.append
- [x] **Resolved**
- **Deliverable**: Modify `meshBuilder.append()` in `mesh.go` to pre-allocate indices slice capacity
- **Dependencies**: None
- **Goal Impact**: Addresses "append() in loop without pre-allocation" anti-pattern
- **Acceptance**: All tests pass
- **Validation**: `go test ./... -v`

### Step 8: Add Age × Species Interaction for Child Proportions
- [ ] **Pending**
- **Deliverable**: Modify `applyAge()` in `transforms.go` to add species-aware head scaling for `AgeChild` and `AgeToddler`
- **Dependencies**: Step 1 (TestAllAges covers this)
- **Goal Impact**: Addresses GAPS.md "Age × Species Interaction Missing"
- **Acceptance**: All tests pass; Gnome toddlers have proportionally larger heads than Human toddlers
- **Validation**: `go test ./... -run TestAllAges -v`

### Step 9: Add Posture × Age Interaction
- [ ] **Pending**
- **Deliverable**: Modify `applyPosture()` in `transforms.go` to auto-adjust posture for `AgeDecrepit`/`AgeElderly` when `PostureUpright` is set
- **Dependencies**: Step 1 (TestAllPostures covers this)
- **Goal Impact**: Addresses GAPS.md "Posture × Age Interaction Missing"
- **Acceptance**: All tests pass; elderly characters with `PostureUpright` receive automatic mild slouch
- **Validation**: `go test ./... -run TestAllPostures -v`

### Step 10: Document Phase 2 Prerequisites in ROADMAP.md
- [ ] **Pending**
- **Deliverable**: Update `ROADMAP.md` Phase 2 section to list technical prerequisites discovered during this work:
  - Need vertex merging algorithm for topology upgrade
  - Need face mesh vertex positions for facial morphing
  - Need ear attachment point coordinates
- **Dependencies**: Steps 1-9
- **Goal Impact**: Prepares Phase 2 work; documents architectural decisions
- **Acceptance**: ROADMAP.md updated with specific technical requirements
- **Validation**: Manual review

## Scope Assessment Details

| Metric | Current | Threshold | Status |
|--------|---------|-----------|--------|
| Functions above complexity 10 | 1 (Validate=19) | <5 | ✅ Small |
| Duplication ratio | 24.3% | <10% | ⚠️ Large |
| Doc coverage gap | 0% | <10% | ✅ Small |
| Missing enum tests | 14 | <5 | ⚠️ Medium |

**Overall Scope**: Medium (duplication is the main driver)

## Verification Commands

```bash
# Full test suite
go test ./... -v

# Benchmark with memory stats
go test -bench=. -benchmem ./...

# Complexity check after Step 4
go-stats-generator analyze . --skip-tests --format json 2>/dev/null | \
  jq -r '.functions[] | select(.complexity.cyclomatic > 10) | "\(.name): \(.complexity.cyclomatic)"'

# Duplication check after Step 5
go-stats-generator analyze . --skip-tests --format json 2>/dev/null | \
  jq -r '.duplication | "ratio=\(.duplication_ratio) clones=\(.clone_pairs)"'

# Count passing tests after Step 1
go test ./... -v 2>&1 | grep -c "^--- PASS"
```

## Notes

- This plan focuses on Phase 1 completion and technical debt reduction
- No source code changes to mesh topology (Phase 2 scope)
- All steps preserve determinism (same seed → same mesh)
- All steps maintain zero external dependencies
- Steps 1-7 can be parallelized; Steps 8-9 depend on Step 1; Step 10 depends on all others

---
*Plan generated 2026-04-03 based on go-stats-generator v1.0.0 analysis and ROADMAP.md/GAPS.md stated goals*
