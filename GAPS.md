# Known Gaps and Limitations

## Current Implementation Gaps

### Parameter interactions
- **Species × Build blending** – Species and Build scale factors are applied
  sequentially rather than blended.  A `SpeciesOrc` + `BuildFragile`
  combination produces a geometrically correct result but the species' natural
  bulk and the frail build partially cancel out in an untuned way.
- **Age × Species** – Child/Toddler ages shrink the entire body uniformly.
  Species-specific juvenile proportions (e.g. a Gnome child should have a
  proportionally even larger head than a Human child) are not modelled.
- **LimbLength × ShoulderWidth** – When both are set to extremes (e.g. Long
  + Broad), the arm X-offset and length scaling are applied independently,
  which can cause arms to clip through the torso on some parameter
  combinations.
- **Posture × Age** – Elderly/Decrepit characters do not automatically adopt a
  more hunched posture; the Posture parameter must be set explicitly.

### Facial geometry
- Facial-feature parameters (`FaceShape`, `Jaw`, `Brow`, `Ears`) adjust head
  radii on the shared ellipsoid but do not produce distinct facial topology.
  There is no dedicated face mesh, so jaw prominence, brow ridge, cheekbones,
  and nasal structure are not geometrically represented — only the overall head
  proportions change.
- Ears are not a separate mesh.  `EarsPointed` / `EarsLarge` slightly widen
  the head ellipsoid as a proxy; pointed or large ears are not visible as
  distinct geometry.

### Mesh continuity
- The body is assembled from disconnected primitives (cylinders, boxes,
  ellipsoid).  There are visible gaps / seams at every joint (shoulder, hip,
  elbow, knee, ankle, neck).  Vertices at part boundaries are not shared.
- Winding order at the seam between the bottom cap of the hips cylinder and
  the top of the upper-leg cylinders may produce incorrect normals under some
  lighting models.

### Proportions mode
- `ProportionsHeroic` widens shoulders and narrows hips but does not elongate
  legs, which is a key characteristic of heroic proportions.  The
  `LimbLength` parameter must be set to `LimbLengthLong` separately.
- `ProportionsCaricature` enlarges the head but does not shrink the hands and
  feet to match, which is typical of caricature style.

## Missing Features

- **Texture / UV support** – All vertices carry a UV coordinate but the
  assembled mesh has no shared UV atlas.  Textures cannot be applied without a
  subsequent UV-unwrap step (planned for Phase 3).
- **Vertex skinning / animation rig** – `JointIds` and `JointWeights` are
  zeroed in every vertex.  The mesh cannot be animated inside Kaiju without
  a manual rigging step (planned for Phase 4).
- **MorphTarget support** – The `MorphTarget` field in `Vertex` is always zero.
  Blend-shape animation (expressions, breathing) is not possible without Phase
  4 work.
- **Individual finger and toe geometry** – Hands are represented as flat boxes;
  fingers and toes are not modelled (planned for Phase 2).
- **Hair / head accessory slot** – There is no attachment point or separate
  mesh token for hair, hats, or horns.
- **Level-of-detail (LOD)** – A single triangle density is produced regardless
  of viewing distance (planned for Phase 5).
- **Mesh caching** – Every `Generate` call rebuilds geometry from scratch.
  For scenes with many identical characters this is wasteful (planned for
  Phase 5).
- **glTF / OBJ export** – There is no standard-format output; the mesh can
  only be consumed directly via the Go API (planned for Phase 6).

## Technical Debt

- **Magic float constants** – Body segment dimensions in `basemesh.go` are
  hard-coded float literals derived empirically from MakeHuman proportions.
  They should be driven by a named-constant table or a small data file that
  can be tuned per species without touching logic code.
- **Primitive winding consistency** – The ellipsoid generator uses lat→lon
  iteration order `(a, b, a+1) / (b, b+1, a+1)` while the box generator uses
  `(base, base+1, base+2) / (base, base+2, base+3)`.  A unified winding
  convention and tests for front-face orientation should be added.
- **`scaleAll` / `scaleHeight` duplication** – Both functions repeat the same
  list of layout fields.  Adding a new body part requires updating both
  functions; a table-driven or reflection-based approach would be safer.
- **`transforms.go` length** – All transformation functions live in one 500-line
  file.  Splitting into per-category files (species.go, age.go, etc.) would
  improve readability.
- **No `rand.Rand` interface** – The generator takes a concrete `*rand.Rand`.
  Introducing an interface would allow deterministic testing of the posture
  jitter without relying on the PRNG sequence being stable across Go versions.

## Compatibility Issues

### Kaiju version constraints
- The `Vertex` struct defined in this package mirrors the layout of
  `kaijuengine.com/rendering.Vertex` as of Kaiju commit
  `7ad3393` (go 1.25).  If the upstream Kaiju struct adds or reorders fields,
  the layout will silently break.  A compile-time size/offset assertion (using
  `unsafe.Sizeof` / `unsafe.Offsetof`) should be added once the Kaiju module
  is importable.
- Kaiju's `rendering.NewMesh` expects `[]rendering.Vertex`, not
  `[]unpeople.Vertex`.  Currently the caller must copy or cast the slice.
  Direct import of Kaiju is blocked because its module path is
  `kaijuengine.com` (not `github.com/KaijuEngine/kaiju`) and it requires
  platform-specific CGo dependencies.

### MakeHuman model format limitations
- The base body layout is a Go-code approximation of MakeHuman proportions, not
  loaded from an actual `.mhx2` or `.obj` export.  Fine-grained surface detail
  (skin folds, muscle definition, fingernail geometry) present in MakeHuman
  exports is absent.
- MakeHuman topology (≈14 000 vertices, quads) is intentionally simplified to
  ~354 vertices (triangles only) to keep generation time below 100 ms and to
  avoid embedding large binary data in the package.
