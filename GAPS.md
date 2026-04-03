# Implementation Gaps — 2026-04-03

This document catalogs gaps between the project's stated goals and its current implementation. Each gap includes the impact on users and steps to close it.

---

## A-Pose Skeleton Export

- **Stated Goal**: ROADMAP.md Phase 4 lists "A-pose export" as a completed feature.
- **Current State**: The skeleton is generated exclusively in T-pose (arms horizontal at 90° from body). There is no option to generate an A-pose skeleton where shoulders are rotated 30-45° downward.
- **Impact**: Game animation pipelines (particularly those using Unity or Unreal Engine humanoid rigs) often prefer A-pose because it provides better shoulder deformation during arm movements. T-pose can cause shoulder mesh "candy wrapper" artifacts when arms are lowered. Users targeting these pipelines must manually adjust the skeleton in a DCC tool after export.
- **Closing the Gap**:
  1. Add `SkeletonPose` enum to `Params` with values `TPose` (default) and `APose`
  2. In `skeleton.go:computeJointPositions()`, apply a quaternion rotation to `JointLeftShoulder` and `JointRightShoulder` when A-pose is requested (~45° rotation around Z-axis)
  3. Rotate shoulder-attached vertices (`upperArmTopL`, `upperArmTopR`) to match the new bind pose
  4. Add test `TestAPoseExport` verifying shoulder angles in the output
  5. Update CLI (`-pose` flag) and REST API (add `pose` field to JSON) to expose the option
  6. Validation: Export glTF with A-pose; import into Blender; verify shoulder angles are ~45° from horizontal

---

## CLI Test Coverage Below Target

- **Stated Goal**: ROADMAP.md Priority 3 states "Increase CLI Test Coverage (70.2% → 85%+)"
- **Current State**: `cmd/unpeopled` has 70.2% test coverage. Tests exist for basic OBJ generation but not for `gltf`, `glb`, `binary`, or `lod` output formats. Error paths (invalid JSON, unsupported format) are untested.
- **Impact**: Regressions in CLI output formats may go undetected. Users relying on the CLI for build pipelines have less assurance of stability than users of the Go API.
- **Closing the Gap**:
  1. Add table-driven tests in `cmd/unpeopled/main_test.go` covering each format:
     - `TestGenerateOBJ` — verify OBJ header comment and vertex count
     - `TestGenerateGLTF` — verify JSON structure with "asset.version": "2.0"
     - `TestGenerateGLB` — verify "glTF" magic bytes at offset 0
     - `TestGenerateBinary` — verify "UNPM" magic bytes
     - `TestGenerateLOD` — verify LOD level selection
  2. Add error case tests:
     - `TestInvalidJSON` — malformed JSON returns non-zero exit
     - `TestInvalidFormat` — unknown format flag returns error
     - `TestInvalidLODLevel` — out-of-range LOD level fails gracefully
  3. Validation: `go test -cover ./cmd/unpeopled` should report ≥85%

---

## No Continuous Integration

- **Stated Goal**: ROADMAP.md Priority 2 specifies "Continuous Integration Setup" with a complete workflow YAML.
- **Current State**: No `.github/workflows/` directory exists. Tests, vet, and build are run manually.
- **Impact**: Contributors may accidentally introduce regressions that aren't caught until manual testing. The project cannot display a CI badge indicating build health, reducing confidence for potential users evaluating the library.
- **Closing the Gap**:
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
  2. Add CI badge to README.md: `![CI](https://github.com/opd-ai/unpeople/actions/workflows/ci.yml/badge.svg)`
  3. Validation: Merge to main; verify green CI status

---

## Unreferenced Exported Functions

- **Stated Goal**: Clean, maintainable codebase with minimal dead code.
- **Current State**: `go-stats-generator` reports 29 unreferenced functions. Most are exported API functions (e.g., `UnlitMaterial`, `SSSkinMaterial`, enum value helpers) that exist for downstream consumers but are not called within the repository itself.
- **Impact**: Minor—these functions serve as public API entry points. However, genuinely dead internal helpers increase cognitive load for maintainers and may mislead new contributors about what code is actually used.
- **Closing the Gap**:
  1. Run `go-stats-generator analyze . --format json | jq '.maintenance.dead_code'` to get the full list
  2. For each function, grep the codebase: `grep -r "FunctionName" .`
  3. Remove internal helpers that are truly unused
  4. For exported functions kept for API completeness, add doc comments explaining their purpose (e.g., "// UnlitMaterial is provided for preview rendering use cases")
  5. Consider adding usage examples in `example/` to exercise exported API functions
  6. Validation: Re-run analysis; dead code count should decrease by ≥50%

---

## Missing ToKaijuVertices Helper

- **Stated Goal**: `mesh.go:9-11` comments reference "a ToKaijuVertices helper (to be added in Phase 6) for a safe conversion path once the Kaiju module is directly importable."
- **Current State**: The comment exists but the function is not implemented. The `kaiju/` package provides `ToKaijuMesh()` which performs the full mesh conversion, but there's no standalone vertex slice helper.
- **Impact**: Minimal—`ToKaijuMesh()` covers the primary use case. The missing helper would only benefit users who need to convert vertex slices independently of the mesh structure, which is an edge case.
- **Closing the Gap**:
  1. Decide whether the helper is truly needed (likely not, given `ToKaijuMesh` exists)
  2. Either implement `ToKaijuVertices([]Vertex) []rendering.Vertex` in the kaiju package, or
  3. Update the comment in `mesh.go` to reference `kaiju.ToKaijuMesh()` as the recommended approach
  4. Validation: Comment or code accurately reflects available functionality

---

## README Documentation Sparse

- **Stated Goal**: Documentation sufficient for the target audience (game developers).
- **Current State**: README.md contains only a single sentence: "Procedurally generated humanoids by deterministic transformations of a base makehuman-exported model". There's no usage example, API overview, or installation instructions.
- **Impact**: New users must read source code or ROADMAP.md to understand how to use the library. This creates a higher barrier to adoption than necessary given the library's maturity.
- **Closing the Gap**:
  1. Add installation section: `go get github.com/opd-ai/unpeople`
  2. Add basic usage example:
     ```go
     gen := unpeople.NewGenerator()
     params := unpeople.DefaultParams()
     params.Seed = 42
     params.Species = unpeople.SpeciesElf
     mesh, err := gen.Generate(params)
     ```
  3. Add section on export formats with CLI examples
  4. Add feature list summarizing species, parameters, and output options
  5. Link to `docs/` for advanced topics (Kaiju integration, REST API)
  6. Validation: README should answer "What is this? How do I install it? How do I use it?" within the first screen

---

## Summary

| Gap | Severity | Effort |
|-----|----------|--------|
| A-Pose Skeleton | Medium | ~2 hours |
| CLI Test Coverage | Medium | ~1 hour |
| CI Setup | Low | ~15 minutes |
| Dead Code Cleanup | Low | ~30 minutes |
| ToKaijuVertices Helper | Low | ~15 minutes |
| README Enhancement | Low | ~30 minutes |

All gaps are operational improvements rather than missing core functionality. The library is production-ready for its stated use case despite these gaps.
