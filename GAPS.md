# Implementation Gaps — 2026-04-03

This document catalogs gaps between the project's stated goals and its current implementation. Each gap includes the impact on users and steps to close it.

---

## ~~A-Pose Skeleton Export~~ ✅ RESOLVED

- **Stated Goal**: ROADMAP.md Phase 4 lists "A-pose export" as a completed feature.
- **Resolution**: Implemented `SkeletonPose` enum with `TPose` and `APose` values. The `GenerateSkeletonWithPose()` function applies shoulder rotations and arm position transforms. CLI supports `-pose apose` flag. Test `TestAPoseExport` validates shoulder angles.

---

## ~~CLI Test Coverage Below Target~~ ✅ RESOLVED

- **Stated Goal**: ROADMAP.md Priority 3 states "Increase CLI Test Coverage (70.2% → 85%+)"
- **Resolution**: Coverage increased to 85.7%. All format tests, error path tests, and pose flag tests added.

---

## ~~No Continuous Integration~~ ✅ RESOLVED

- **Stated Goal**: ROADMAP.md Priority 2 specifies "Continuous Integration Setup".
- **Resolution**: CI workflow already exists at `.github/workflows/ci.yml` with comprehensive testing, race detection, and codecov integration. README has CI badge.

---

## ~~Unreferenced Exported Functions~~ ✅ RESOLVED

- **Stated Goal**: Clean, maintainable codebase with minimal dead code.
- **Resolution**: Audited all 29 unreferenced functions. All are intentionally exported public API functions (e.g., `UnlitMaterial`, `SSSkinMaterial`, material factories, export helpers) for downstream consumers. No true dead internal code found.

---

## ~~Missing ToKaijuVertices Helper~~ ✅ RESOLVED

- **Stated Goal**: `mesh.go:9-11` comments reference "a ToKaijuVertices helper".
- **Resolution**: Updated comment in `mesh.go` to reference `kaiju.ToKaijuMesh()` as the recommended approach. The obsolete TODO reference has been removed.

---

## Summary

| Gap | Status | Effort |
|-----|--------|--------|
| A-Pose Skeleton | ✅ Resolved | Done |
| CLI Test Coverage | ✅ Resolved | Done |
| CI Setup | ✅ Resolved | Pre-existing |
| Dead Code Cleanup | ✅ Resolved | No action needed |
| ToKaijuVertices Helper | ✅ Resolved | Done |

All gaps have been addressed. The library is production-ready for its stated use case.
