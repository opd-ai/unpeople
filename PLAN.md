# Implementation Plan: Complete Remaining Gaps

## Project Context
- **What it does**: Go library for deterministic procedural generation of humanoid character meshes from a seed and descriptive parameters, producing Kaiju engine-compatible output.
- **Current goal**: Complete the one partial feature (A-Pose Skeleton Export) and address operational gaps (CI, CLI test coverage, documentation).
- **Estimated Scope**: Small (41/42 features complete; 4 operational improvements)

## Goal-Achievement Status
| Stated Goal | Current Status | This Plan Addresses |
|-------------|---------------|---------------------|
| Phase 1: Core Implementation (14 items) | ✅ Complete | No |
| Phase 2: Enhanced Geometry (7 items) | ✅ Complete | No |
| Phase 3: Texture & Material (4 items) | ✅ Complete | No |
| Phase 4: Skeletal Rig Support (A-pose) | ⚠️ Partial | **Yes** |
| Phase 5: Performance & Scalability (4 items) | ✅ Complete | No |
| Phase 6: Ecosystem Integration (5 items) | ✅ Complete | No |
| Continuous Integration Setup | ❌ Missing | **Yes** |
| CLI Test Coverage ≥85% | ⚠️ 70.2% | **Yes** |
| README Documentation | ⚠️ Sparse | **Yes** |

## Metrics Summary
- Complexity hotspots on goal-critical paths: **0** functions above threshold (CC>9)
- Medium complexity (CC 6): 7 functions (acceptable for export/CLI code)
- Duplication ratio: **0.72%** (excellent)
- Doc coverage: **98.4%** (excellent)
- Test coverage: 86.9% (main), 82.9% (server), **70.2% (CLI)**, 100% (kaiju)
- Package coupling: Clean (no circular dependencies)
- Total: 4,366 LoC, 284 functions, 3 packages

## Competitive Context
There is no equivalent Go library for procedural human mesh generation. Comparable tools (MakeHuman, Blender MPFB) are Python/C++ and require external asset pipelines. `unpeople` enables runtime generation directly in Go game engines—a differentiator for the Kaiju ecosystem.

---

## Implementation Steps

### Step 1: A-Pose Skeleton Export
- **Deliverable**: Add `SkeletonPose` enum to `params.go`, implement shoulder rotation in `skeleton.go`, update CLI `-pose` flag and REST API `pose` field
- **Dependencies**: None
- **Goal Impact**: Completes Phase 4 (the only partial goal)
- **Acceptance**: 
  - `TestAPoseExport` passes verifying shoulder angles ~45° from horizontal
  - glTF export with A-pose loads correctly in Blender
- **Validation**: 
  ```bash
  go test -v -run TestAPoseExport ./...
  ```

**Implementation Details**:
1. Add `SkeletonPose` enum to `params.go`:
   ```go
   type SkeletonPose int
   const (
       PoseTPose SkeletonPose = iota
       PoseAPose
   )
   ```
2. Add `Pose` field to `Params` struct (default `PoseTPose`)
3. In `skeleton.go:computeJointPositions()`, apply quaternion rotation to `JointLeftShoulder` and `JointRightShoulder` when `PoseAPose` is requested (~45° rotation around Z-axis)
4. Rotate shoulder-attached vertices to match bind pose
5. Update `cmd/unpeopled/main.go` to add `-pose` flag
6. Update `cmd/unpeople-server/main.go` to accept `pose` in JSON

---

### Step 2: Continuous Integration Setup
- **Deliverable**: `.github/workflows/ci.yml` with vet/build/test gates
- **Dependencies**: None
- **Goal Impact**: Prevents regressions, enables CI badge for README
- **Acceptance**: CI workflow runs successfully on push/PR
- **Validation**: 
  ```bash
  # Verify workflow syntax locally
  cat .github/workflows/ci.yml | head -20
  # After push: check GitHub Actions tab for green status
  ```

**Implementation Details**:
1. Create `.github/workflows/ci.yml`:
   ```yaml
   name: CI
   on: [push, pull_request]
   jobs:
     test:
       runs-on: ubuntu-latest
       steps:
         - uses: actions/checkout@v4
         - uses: actions/setup-go@v5
           with:
             go-version: '1.21'
         - run: go vet ./...
         - run: go build ./...
         - run: go test -race -cover ./...
   ```

---

### Step 3: CLI Test Coverage Improvement
- **Deliverable**: Additional tests in `cmd/unpeopled/main_test.go` covering all output formats and error paths
- **Dependencies**: None
- **Goal Impact**: Raises CLI coverage from 70.2% to ≥85%
- **Acceptance**: `go test -cover ./cmd/unpeopled` reports ≥85%
- **Validation**: 
  ```bash
  go test -cover ./cmd/unpeopled
  ```

**Implementation Details**:
1. Add format-specific tests:
   - `TestGenerateOBJ` — verify OBJ header and vertex output
   - `TestGenerateGLTF` — verify JSON with `"asset":{"version":"2.0"}`
   - `TestGenerateGLB` — verify `glTF` magic bytes at offset 0
   - `TestGenerateBinary` — verify `UNPM` magic bytes
   - `TestGenerateLOD` — verify LOD level selection works
2. Add error path tests:
   - `TestInvalidJSON` — malformed JSON returns non-zero exit
   - `TestInvalidFormat` — unknown format flag returns error
   - `TestInvalidLODLevel` — out-of-range LOD fails gracefully
3. Use `bytes.Buffer` as mock stdin for parameterized tests

---

### Step 4: README Enhancement
- **Deliverable**: Comprehensive README.md with installation, usage, API overview, and feature list
- **Dependencies**: Step 2 (CI badge requires workflow to exist)
- **Goal Impact**: Lowers adoption barrier for new users
- **Acceptance**: README answers "What? How to install? How to use?" within first screen
- **Validation**: Manual review — README should have:
  - [ ] CI badge
  - [ ] One-line description
  - [ ] Installation command
  - [ ] Basic usage example
  - [ ] Feature list (species, parameters, exports)
  - [ ] Link to docs/

**Implementation Details**:
1. Add CI badge: `![CI](https://github.com/opd-ai/unpeople/actions/workflows/ci.yml/badge.svg)`
2. Add installation: `go get github.com/opd-ai/unpeople`
3. Add basic usage example:
   ```go
   gen := unpeople.NewGenerator()
   params := unpeople.DefaultParams()
   params.Seed = 42
   params.Species = unpeople.SpeciesElf
   mesh, err := gen.Generate(params)
   ```
4. Add feature list summarizing 10 species, ~20 parameters, 3 export formats
5. Add CLI usage examples
6. Link to `docs/` for Kaiju integration, REST API

---

## Lower Priority Items (Not in This Plan)

These items are documented in GAPS.md but have minimal impact:

| Item | Reason Deferred |
|------|-----------------|
| Dead code cleanup (29 functions) | Mostly exported API for downstream; no functional impact |
| ToKaijuVertices helper | `ToKaijuMesh()` covers primary use case |
| Code organization suggestions | Stylistic; no functional impact |

---

## Execution Order

```
Step 1 (A-Pose) ─────────────────────────┐
                                         ├──→ Step 4 (README)
Step 2 (CI Setup) ───────────────────────┘
       │
       └──→ (enables badge for Step 4)

Step 3 (CLI Tests) ─────────────────────────→ (independent)
```

Steps 1, 2, and 3 can proceed in parallel. Step 4 should wait for Step 2 to add the CI badge.

---

## Success Criteria

- [ ] All 42/42 roadmap goals marked complete (A-pose implemented)
- [ ] CI badge shows green on main branch
- [ ] `go test -cover ./cmd/unpeopled` ≥85%
- [ ] README has installation, usage, and examples
- [ ] `go vet ./...` passes with no new warnings
- [ ] `go test -race ./...` passes
