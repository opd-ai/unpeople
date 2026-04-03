# Implementation Gaps — 2026-04-03

## Heroic Proportions Missing Leg Elongation

- **Stated Goal**: `ProportionsHeroic` should produce a heroic body style with wide shoulders, narrow hips, AND elongated legs per industry convention for heroic character proportions.
- **Current State**: `transforms.go:156-159` widens shoulders (`chestRX *= 1.15`) and narrows hips (`hipsRX *= 0.92`) but does NOT adjust leg length. Users must manually set `LimbLength = LimbLengthLong` to achieve the full heroic look.
- **Impact**: Characters with `ProportionsHeroic` appear stockier than expected; users unfamiliar with the library may not realize they need to combine parameters to achieve the documented "heroic" aesthetic.
- **Closing the Gap**: Add `scaleLimbs(l, 1.08)` within the `case ProportionsHeroic:` block in `applyProportions()`. This provides automatic leg elongation while still allowing users to override with explicit `LimbLength` settings applied afterward.

## Caricature Proportions Missing Extremity Reduction

- **Stated Goal**: `ProportionsCaricature` should produce a caricature style with exaggerated head AND proportionally reduced hands/feet, matching typical caricature art conventions.
- **Current State**: `transforms.go:167-172` enlarges the head (`headRX *= 1.25`, etc.) and slightly shrinks the chest/arms, but hands and feet remain at default size, creating an unbalanced caricature appearance.
- **Impact**: Generated caricature characters have normal-sized hands and feet relative to an oversized head, which looks incorrect compared to traditional caricature artwork.
- **Closing the Gap**: Add hand and foot reduction within `case ProportionsCaricature:`:
  ```go
  l.handHW *= 0.85
  l.handHH *= 0.85
  l.handHD *= 0.85
  l.footHW *= 0.85
  l.footHD *= 0.85
  ```

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

- **Stated Goal**: Generate humanoid meshes suitable for rendering in the Kaiju engine.
- **Current State**: `basemesh.go:159-240` assembles the body from disconnected geometric primitives (ellipsoid head, cylindrical limbs, box hands/feet). Vertices at part boundaries are not shared, resulting in:
  - Visible gaps at shoulders, hips, elbows, knees, ankles, and neck
  - Potential lighting artifacts at seams under certain shading models
  - Winding order inconsistencies between primitive types
- **Impact**: Generated meshes have a "mannequin" appearance with visible part boundaries. This is acceptable for prototyping but not for production-quality character rendering.
- **Closing the Gap**: Phase 2 "Topology upgrade" per ROADMAP:
  1. Replace cylindrical/box primitives with subdivision-surface body parts
  2. Share vertices across part boundaries
  3. Ensure consistent winding order across all primitives

## Missing Enum Test Coverage

- **Stated Goal**: ROADMAP claims "Unit tests: determinism, mesh validity, all species, all heights, all builds, all ages, validation, performance."
- **Current State**: `generator_test.go` has comprehensive tests for Species, Height, Build, and Age enum iteration, but is missing:
  - `TestAllProportions` (4 values: Heroic, Realistic, Stylized, Caricature)
  - `TestAllPhenotypes` (3 values: Masculine, Androgynous, Feminine)
  - `TestAllPostures` (4 values: Upright, Slouched, Hunched, Rigid)
  - `TestAllFaceShapes` (6 values)
  - `TestAllJaws` (5 values)
  - `TestAllBrows` (4 values)
  - `TestAllEars` (5 values)
  - `TestAllShoulderWidths`, `TestAllHipWidths`, `TestAllLimbLengths`, `TestAllNeckLengths`
  - `TestAllHandSizes`, `TestAllFingerLengths`, `TestAllFootSizes`
- **Impact**: New enum values or changes to enum handling could introduce bugs that go undetected until runtime. The test suite does not fully validate the claim of "all" parameter coverage.
- **Closing the Gap**: Add table-driven tests for each missing enum type following the pattern established by `TestAllSpecies`:
  ```go
  func TestAllProportions(t *testing.T) {
      g := unpeople.NewGenerator()
      for p := unpeople.ProportionsHeroic; p <= unpeople.ProportionsCaricature; p++ {
          params := unpeople.DefaultParams()
          params.Proportions = p
          if _, err := g.Generate(params); err != nil {
              t.Errorf("proportions=%d: %v", p, err)
          }
      }
  }
  ```

## Validate() Function High Complexity

- **Stated Goal**: Maintainable codebase following Go best practices.
- **Current State**: `params.go:303-359` has cyclomatic complexity of 19 (well above the recommended threshold of 10). The function contains 18 sequential if-statements, one for each parameter field.
- **Impact**: 
  - Adding new parameters requires manually adding another if-statement, risking copy-paste errors
  - The function is difficult to review at a glance
  - Higher likelihood of validation logic drift between parameters
- **Closing the Gap**: Refactor to table-driven validation:
  ```go
  type paramRange struct {
      name string
      val  int
      min  int
      max  int
  }
  
  func (p *Params) Validate() error {
      checks := []paramRange{
          {"Species", int(p.Species), int(SpeciesHuman), int(SpeciesOgre)},
          {"Height", int(p.Height), int(HeightGiant), int(HeightTiny)},
          // ... remaining parameters
      }
      for _, c := range checks {
          if c.val < c.min || c.val > c.max {
              return fmt.Errorf("invalid %s value", c.name)
          }
      }
      return nil
  }
  ```

## Code Duplication in Scale Helpers

- **Stated Goal**: Maintainable codebase with minimal technical debt.
- **Current State**: `transforms.go:460-598` contains `scaleAll()`, `scaleHeight()`, and `scaleLimbs()` functions that repeat similar patterns of field access. The previous GAPS.md acknowledged: "Adding a new body part requires updating both functions; a table-driven or reflection-based approach would be safer."
- **Impact**: 
  - 19.4% code duplication ratio (343 duplicated lines per go-stats-generator)
  - Adding new body parts (e.g., ears, fingers) requires updating 3 separate functions
  - Risk of inconsistency if one function is updated but others are forgotten
- **Closing the Gap**: Create a unified scaling mechanism:
  1. Define a layout field accessor table mapping field names to getter/setter functions
  2. Implement `scaleLayout(l *bodyLayout, fields []string, factor float32)` that iterates the table
  3. Refactor `scaleAll`, `scaleHeight`, `scaleLimbs` to use the unified mechanism

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

- **Stated Goal**: UV coordinates present in vertices for texturing capability.
- **Current State**: Each primitive generates its own UV coordinates (`primitive.go:47`, `primitive.go:151`, `primitive.go:211-212`), but these are per-primitive and not unified into a shared atlas. Textures cannot be applied without a subsequent UV-unwrap step.
- **Impact**: Users cannot texture the generated meshes directly; they must export and process in external tools like Blender to create a unified UV layout.
- **Closing the Gap**: Phase 3 "UV atlas generation" per ROADMAP:
  1. Calculate a shared UV space for all body parts
  2. Pack primitive UVs into non-overlapping atlas regions
  3. Generate texture coordinate mapping for skin tone and detail textures

## Individual Finger and Toe Geometry Missing

- **Stated Goal**: Hand size and finger length parameters affecting geometry.
- **Current State**: `basemesh.go:211-215` generates hands as flat boxes; `FingerLength` parameter only adjusts the box's half-height (`handHH`), not actual finger geometry. Toes are completely absent.
- **Impact**: Characters have mitten-like hands without visible fingers. The `FingerLength` parameter has minimal visual effect.
- **Closing the Gap**: Phase 2 "Finger geometry" and "Toe geometry" per ROADMAP:
  1. Replace hand box with palm + 5 finger segments (proximal/middle/distal phalanges)
  2. Drive finger segment lengths from `FingerLength` parameter
  3. Add toe segments to foot mesh

## Species × Build Interaction Untuned

- **Stated Goal**: Species and Build parameters combine to produce varied body types.
- **Current State**: `transforms.go:30-150` applies Species scaling first, then Build scaling. These multiply together without blending, so `SpeciesOrc + BuildFragile` produces a result where the species' bulk and the frail build partially cancel out in ways that may not look natural.
- **Impact**: Certain Species × Build combinations produce awkward proportions because the scaling factors were tuned independently rather than as an interaction matrix.
- **Closing the Gap**: Implement species-specific build scaling multipliers:
  ```go
  // Example: Orcs should have a less extreme fragile reduction
  if s == SpeciesOrc && build == BuildFragile {
      // Use 0.90 instead of 0.80 for chest reduction
  }
  ```

## Age × Species Interaction Missing

- **Stated Goal**: Age stages from Decrepit to Toddler with appropriate proportions.
- **Current State**: `transforms.go:200-234` applies uniform scaling for child ages regardless of species. A Gnome child should have a proportionally even larger head than a Human child (Gnomes already have large heads), but this species-specific adjustment is not implemented.
- **Impact**: Child/Toddler characters of different species look more similar than they should; species-distinctive features are less pronounced at young ages.
- **Closing the Gap**: Add species-aware child proportion adjustments:
  ```go
  case AgeChild:
      scaleAll(l, 0.70)
      headScale := 1.15
      if species == SpeciesGnome {
          headScale = 1.25 // Gnomes have proportionally larger heads
      }
      l.headRX *= headScale
      // ...
  ```

## Posture × Age Interaction Missing

- **Stated Goal**: Posture variants including hunched posture.
- **Current State**: `transforms.go:430-440` applies posture independently of age. Elderly/Decrepit characters with `PostureUpright` remain perfectly straight despite the age-related physical changes that typically cause older individuals to adopt a more hunched stance.
- **Impact**: Elderly characters look unnaturally erect unless users remember to set `Posture = PostureHunched` explicitly.
- **Closing the Gap**: Add automatic posture blending for elderly ages:
  ```go
  func applyPosture(l *bodyLayout, p Posture, age Age, rng *splitmix64) {
      // Auto-adjust for age
      if age <= AgeElderly && p == PostureUpright {
          p = PostureSlouched // Automatic mild slouch for elderly
      }
      // ... existing posture logic
  }
  ```

---
*Gaps analysis generated by automated audit process comparing ROADMAP.md stated goals against actual implementation.*
