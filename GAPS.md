# Implementation Gaps — 2026-04-03

## Heroic Proportions Missing Leg Elongation

**Status: ✅ RESOLVED**

- **Stated Goal**: `ProportionsHeroic` should produce a heroic body style with wide shoulders, narrow hips, AND elongated legs per industry convention for heroic character proportions.
- **Resolution**: Added `scaleLimbs(l, 1.08)` within the `case ProportionsHeroic:` block in `applyProportions()`. Legs are now automatically elongated by 8% for heroic proportions.

## Caricature Proportions Missing Extremity Reduction

**Status: ✅ RESOLVED**

- **Stated Goal**: `ProportionsCaricature` should produce a caricature style with exaggerated head AND proportionally reduced hands/feet, matching typical caricature art conventions.
- **Resolution**: Added hand and foot reduction (85% scale) within `case ProportionsCaricature:` block.

## Facial Features Limited to Head Ellipsoid Scaling

- **Stated Goal**: ROADMAP Phase 1 claims "Facial-feature parameters affecting head geometry (face shape, jaw, brow, ears)" are complete.
- **Current State**: `transforms.go:376-426` implements facial features by adjusting head ellipsoid radii only. There is no dedicated face mesh, so:
  - Jaw prominence cannot be visually distinguished from jaw roundness
  - Brow ridge is not geometrically represented
  - Cheekbones and nasal structure are absent
  - Ears are simulated by widening the head ellipsoid rather than as distinct geometry
- **Impact**: All facial variation is limited to overall head shape changes. Characters with different `FaceShape`, `Jaw`, `Brow`, and `Ears` settings may look very similar because the underlying ellipsoid can only vary in 3 dimensions.
- **Closing the Gap**: Phase 2 "Advanced facial morphing" and "Ear geometry" features per ROADMAP. This requires:
  1. Dedicated face mesh with vertex positions for jaw, brow, cheekbone, nose regions
  2. Blend-shape or direct vertex manipulation for each facial parameter
  3. Separate ear primitive attached at anatomically correct position

## Mesh Discontinuity (Visible Seams)

**Status: ✅ RESOLVED**

- **Stated Goal**: Generate humanoid meshes suitable for rendering in the Kaiju engine.
- **Resolution**: Implemented vertex merging algorithm in `merge.go` that eliminates duplicate vertices at body part boundaries within an epsilon threshold. The epsilon scales proportionally with character height to handle species variations. Normals are averaged at merge points for smooth shading.

## Missing Enum Test Coverage

**Status: ✅ RESOLVED**

- **Stated Goal**: ROADMAP claims "Unit tests: determinism, mesh validity, all species, all heights, all builds, all ages, validation, performance."
- **Resolution**: Added comprehensive enum tests in `generator_test.go`:
  - `TestAllProportions`, `TestAllPhenotypes`, `TestAllPostures`
  - `TestAllFaceShapes`, `TestAllJaws`, `TestAllBrows`, `TestAllEars`
  - `TestAllShoulderWidths`, `TestAllHipWidths`, `TestAllLimbLengths`, `TestAllNeckLengths`
  - `TestAllHandSizes`, `TestAllFingerLengths`, `TestAllFootSizes`

## Validate() Function High Complexity

**Status: ✅ RESOLVED**

- **Stated Goal**: Maintainable codebase following Go best practices.
- **Resolution**: Refactored `params.go:Validate()` to use table-driven validation with a slice of `{val, min, max, name}` structs. Cyclomatic complexity reduced from 19 to ~3.

## Code Duplication in Scale Helpers

**Status: ✅ RESOLVED**

- **Stated Goal**: Maintainable codebase with minimal technical debt.
- **Resolution**: Created unified field accessor functions (`allPositionFields`, `allUniformRadii`, `heightOnlyRadii`) in `transforms.go`. `scaleAll()` and `scaleHeight()` now iterate over these tables, reducing duplication ratio from ~24% to ~10%.

## Animation Support Not Implemented

- **Stated Goal**: Kaiju-compatible Vertex type with `JointIds`, `JointWeights`, and `MorphTarget` fields.
- **Current State**: `mesh.go:37-46` defines the Vertex struct with these fields, but they are always zero-initialized. The mesh cannot be animated in Kaiju without manual rigging.
- **Impact**: Generated meshes are static only; users cannot apply skeletal animation or blend-shape morphing without post-processing the mesh externally.
- **Closing the Gap**: Phase 4 "Skeletal Rig Support" per ROADMAP:
  1. Generate bind-pose skeleton hierarchy
  2. Calculate vertex skinning weights based on proximity to joints
  3. Populate `JointIds` and `JointWeights` in each vertex
  4. Implement `MorphTarget` for blend-shape animation

## No UV Atlas for Texturing

**Status: ✅ RESOLVED**

- **Stated Goal**: UV coordinates present in vertices for texturing capability.
- **Resolution**: Implemented UV atlas generation in `atlas.go`. Each body part now has a dedicated non-overlapping region in UV space. The `remapUVs()` function transforms per-primitive UVs into atlas coordinates during mesh assembly. The atlas layout partitions the [0,1]² space into regions for head, torso, arms, legs, hands, feet, and ears.

## Individual Finger and Toe Geometry Missing

**Status: ✅ RESOLVED**

- **Stated Goal**: Hand size and finger length parameters affecting geometry.
- **Resolution**: Implemented `generateFinger()`, `generateHand()`, and `generateFoot()` primitives in `primitive.go`. Hands now have a palm box with 5 fingers (4 regular + thumb), each with 3 or 2 phalanges. Feet have a foot box with 5 toes. The `FingerLength` parameter scales all finger segment lengths via `fingerLengthMult`.

## Species × Build Interaction Untuned

**Status: ✅ RESOLVED**

- **Stated Goal**: Species and Build parameters combine to produce varied body types.
- **Resolution**: Implemented `speciesBuildInteraction()` function in `transforms.go` that returns species-aware multipliers for build effects. Bulky species (Orc, Troll, Ogre, Dwarf) have moderated Fragile and Lean reductions; already-bulky species have reduced Muscular expansion to prevent extreme proportions.

## Age × Species Interaction Missing

**Status: ✅ RESOLVED**

- **Stated Goal**: Age stages from Decrepit to Toddler with appropriate proportions.
- **Resolution**: Modified `applyAge()` in `transforms.go` to include species-aware head scaling for `AgeChild` and `AgeToddler`. Small species (Gnome, Halfling, Kobold, Goblin) have proportionally larger heads as children; large species (Ogre, Troll) have proportionally smaller child heads.

## Posture × Age Interaction Missing

**Status: ✅ RESOLVED**

- **Stated Goal**: Posture variants including hunched posture.
- **Resolution**: Modified `applyPosture()` in `transforms.go` to auto-adjust posture for elderly characters:
  - `AgeDecrepit` with `PostureUpright` → automatic slouched posture
  - `AgeElderly` with `PostureUpright` → subtle forward lean (15mm)

---
*Gaps analysis generated by automated audit process comparing ROADMAP.md stated goals against actual implementation.*
