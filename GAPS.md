# Implementation Gaps — 2026-04-04

This document catalogs gaps between the project's stated goals (README.md) and its current implementation. Each gap includes impact assessment and specific steps to close it.

---

## Animation Pipeline: No BVH Import

- **Stated Goal**: Skeleton and skinning support for animation (implied by "Skinning — Vertex weights for skeletal deformation")
- **Current State**: Skeleton generation and skinning weights are implemented, but no animation data import capability exists
- **Impact**: Users cannot apply motion capture data (BVH format) to generated characters without implementing their own import
- **Closing the Gap**:
  1. Implement stdlib-only BVH parser (no external deps)
  2. Map BVH joint names to unpeople skeleton joints
  3. Add `Generator.GenerateAnimated(params, bvhPath)` method
  4. Export animated glTF with animation channels
  5. **Validation**: Export animated glTF that plays correctly in Blender/Three.js

**Note**: This is documented in ROADMAP.md Priority 2.

---

## Facial Detail: Parametric Head Only

- **Stated Goal**: Face parameters (Shape, Jaw, Brow, Ears) suggest detailed facial geometry
- **Current State**: Facial features only adjust head ellipsoid radii. No dedicated mesh geometry for eyes, nose, mouth, or ears.
- **Impact**: Facial customization has limited visual effect. Close-up renders show featureless head shape.
- **Closing the Gap**:
  1. Implement facial mesh subdivision for eye sockets, nose bridge, mouth
  2. Add ear geometry as separate mesh pieces positioned on head
  3. Scale/morph facial features based on FaceShape, Jaw, Brow, Ears params
  4. **Validation**: Generated face passes visual inspection at 10-unit camera distance

**Note**: This is documented in ROADMAP.md Priority 3.

---

## Hand Geometry: Simplified Boxes

- **Stated Goal**: Hand parameters (HandSize, FingerLength) suggest articulated hands
- **Current State**: Hands are flat boxes without finger geometry. FingerLength parameter adjusts box dimensions but doesn't create visible fingers.
- **Impact**: Hands appear as paddle-like shapes, unsuitable for close-up renders or first-person views
- **Closing the Gap**:
  1. Implement finger cylinders with 3 phalanges per finger (2 for thumb)
  2. Add knuckle joints at each phalanx boundary
  3. Scale finger proportions based on FingerLength param
  4. **Validation**: Hand mesh has 15 distinct finger segments (4 fingers × 3 + thumb × 3)

**Note**: This is documented in ROADMAP.md Priority 4.

---

## Gap Summary

| Gap | Severity | Type | Effort |
|-----|----------|------|--------|
| No BVH animation import | Medium | Feature | Medium-High |
| Parametric-only facial geometry | Medium | Feature | High |
| Box-shaped hands | Medium | Feature | Medium |

---

## Resolved Gaps (from prior audits)

The following gaps from the previous `GAPS.md` have been verified as resolved:

| Gap | Status | Resolution |
|-----|--------|------------|
| Skeleton joint count (56 vs 52) | ✅ Resolved | Documentation updated to show 52-joint hierarchy (matches implementation) |
| Primitive mesh seams | ✅ Resolved | Vertex merging infrastructure added (`MergeNearbyVertices`, `FindBoundaryVertices`, `StitchEdgeLoops`); `Params.MergeVertices` enables seamless topology |
| A-Pose Skeleton Export | ✅ Resolved | `SkeletonPose` enum with T-pose/A-pose; CLI `-pose apose` flag |
| CLI Test Coverage | ✅ Resolved | Coverage at 85.9% (above 85% target) |
| CI Setup | ✅ Resolved | `.github/workflows/ci.yml` with race detection and codecov |
| Unreferenced Functions | ✅ Resolved | All 29 are intentionally exported public API |
| ToKaijuVertices Helper | ✅ Resolved | `kaiju.ToKaijuMesh()` documented in `mesh.go:10` |

---

*Gaps analysis updated 2026-04-04*
