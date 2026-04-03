# Humanoid Generator Roadmap

## Phase 1: Core Implementation (Current)

- [x] Parameter struct with full validation
- [x] Seeded deterministic PRNG (same seed + params = identical mesh)
- [x] Kaiju-compatible Vertex / Mesh types (`Vec2`, `Vec3`, `Vec4`, `Vec4i`, `Color`)
- [x] Base humanoid body layout approximating a MakeHuman neutral T-pose
- [x] Geometric primitive generators: ellipsoid, tapered cylinder, axis-aligned box
- [x] Species variations: Human, Elf, Dwarf, Gnome, Halfling, Goblin, Kobold, Orc, Troll, Ogre
- [x] Height tiers: Giant, Tall, Medium, Short, Tiny
- [x] Build profiles: Muscular, Athletic, Average, Lean, Stocky, Fragile
- [x] Proportion styles: Heroic, Realistic, Stylized, Caricature
- [x] Phenotype dimorphism: Masculine, Androgynous, Feminine
- [x] Age stages: Decrepit → Toddler (8 tiers)
- [x] Posture variants: Upright, Slouched, Hunched, Rigid
- [x] Facial-feature parameters affecting head geometry (face shape, jaw, brow, ears)
- [x] Shoulder width, hip width, limb length, neck length
- [x] Hand size + finger length, foot size
- [x] Default gray opaque material colour on every vertex
- [x] Mesh key encoding for Kaiju engine mesh cache
- [x] Example binary demonstrating parameter variety
- [x] Unit tests: determinism, mesh validity, all species, all heights, all builds, all ages, validation, performance (<100 ms)

## Phase 2: Enhanced Geometry

### Technical Prerequisites

The following technical requirements were identified during Phase 1 completion
and must be addressed before or during Phase 2 implementation:

1. **Vertex merging algorithm** – Topology upgrade requires an algorithm to
   identify and merge boundary vertices between adjacent body parts (e.g.,
   shoulder-to-upper-arm, hip-to-upper-leg). Candidates: KD-tree spatial
   lookup with epsilon threshold, or explicit vertex correspondence tables.

2. **Face mesh vertex positions** – Advanced facial morphing needs a predefined
   face mesh template with named vertex groups for jaw, brow, cheekbones, nose,
   and chin regions. Consider importing a simplified MakeHuman face topology.

3. **Ear attachment coordinates** – Ear geometry requires precise attachment
   points on the head mesh. Store as bodyLayout fields (earAttachL, earAttachR)
   derived from headCenter and headRX.

4. **Finger bone hierarchy** – Finger geometry needs a bone chain definition
   for proximal/middle/distal phalanges per finger (5 fingers × 3 bones × 2
   hands = 30 segments). Consider generating from hand box corner positions.

### Feature Items

- [x] **Topology upgrade** – Replace cylindrical/box primitives with true
  subdivision-surface body parts that share vertices across part boundaries,
  eliminating visible seams at shoulders, hips, and ankles.
- [x] **Advanced facial morphing** – Dedicated face mesh with blend-shape
  targets for jaw shape, brow ridge, cheekbones, nose bridge, and chin.
- [x] **Ear geometry** – Separate ear primitive (taper/point mesh) attached to
  the head at the correct anatomical position.
- [x] **Finger geometry** – Individual finger segments (proximal / middle /
  distal phalanges) driven by the `FingerLength` parameter.
- [x] **Toe geometry** – Toe segments on the foot mesh.
- [x] **Musculature detail** – Normal-map baked geometry bumps driven by the
  `Build` parameter (requires Phase 3 UV atlas).
- [x] **Hair/skull cap placeholder** – Separate mesh token for hair slot that
  downstream systems can swap.

## Phase 3: Texture & Material System

- [x] **UV atlas generation** – Automatic UV unwrap of the assembled humanoid
  mesh so that textures can be applied.
- [x] **Procedural skin-tone colour** – Per-vertex colour variation driven by a
  `SkinTone` parameter (Pale → Dark, warm/cool undertone).
- [x] **Material export** – Output a Kaiju-compatible `rendering.ShaderDef` or
  material descriptor alongside the mesh.
- [x] **Texture generation** – Noise-driven procedural skin texture baked to an
  atlas (freckles, blemishes, age spots for elderly characters).

## Phase 4: Skeletal Rig Support

- [x] **Bind-pose skeleton** – Export a joint hierarchy (root → spine →
  shoulders/hips → limb chains) matching the generated mesh.
- [x] **Vertex skinning weights** – Populate the `JointIds` / `JointWeights`
  fields in every `Vertex` so the mesh can be animated in Kaiju.
- [x] **MorphTarget support** – Fill `Vertex.MorphTarget` for blend-shape
  animation (facial expressions, breathing, etc.).
- [x] **Animation-ready A-pose export** – Ensure the generated skeleton
  conforms to industry-standard bind-pose conventions (A-pose variant).

## Phase 5: Performance & Scalability

- [x] **Mesh caching layer** – In-process LRU cache keyed on the full Params
  struct so repeated calls with identical inputs skip geometry rebuild.
- [x] **Level-of-detail (LOD) generation** – Automatically produce 3 LOD
  variants (100 %, 50 %, 25 % triangle budget) from a single Generate call.
- [x] **Parallel generation** – Worker-pool API for generating large batches of
  characters concurrently.
- [x] **Streaming output** – io.Writer / channel-based API for very large
  scenes where holding all meshes in memory is impractical.

## Phase 6: Ecosystem Integration

- [ ] **Kaiju engine plug-in** – A drop-in `Generator` that registers with
  Kaiju's asset pipeline and produces `rendering.Mesh` objects directly.
- [ ] **glTF 2.0 export** – Standard interchange format for use outside Kaiju.
- [x] **OBJ export** – Simple text-format export for debugging in DCC tools.
- [x] **CLI tool** – `unpeopled` command that accepts JSON parameters on stdin
  and writes a mesh file (OBJ or UNPM binary) to stdout.
- [ ] **REST API** – HTTP microservice wrapper around the generator for
  integration with non-Go game engines.
