# Architecture Overview

This document describes the internal architecture of the `unpeople` library for contributors and integrators.

## Core Design Principles

1. **Determinism**: Same seed + params = identical mesh (always)
2. **Zero Dependencies**: Standard library only
3. **Kaiju Compatibility**: Vertex layout matches Kaiju engine
4. **Stateless Generation**: Generator has no mutable state

## Generation Pipeline

```
┌────────────┐     ┌─────────────┐     ┌─────────────┐     ┌────────────┐     ┌───────┐
│   Params   │ ──▶ │    PRNG     │ ──▶ │ BodyLayout  │ ──▶ │ Primitives │ ──▶ │  Mesh │
│  (+ Seed)  │     │ (splitmix64)│     │  (geometry) │     │  (shapes)  │     │       │
└────────────┘     └─────────────┘     └─────────────┘     └────────────┘     └───────┘
```

### Stage 1: Parameter Validation

**File**: `params.go`

Input parameters are validated to ensure all enum values are within range. The `Params` struct contains:

- Character attributes: Species, Height, Build, Proportions, Phenotype
- Age & posture: Age (8 stages), Posture (4 types)
- Facial features: FaceShape, Jaw, Brow, Ears
- Body details: ShoulderWidth, HipWidth, LimbLength, NeckLength
- Extremities: HandSize, FingerLength, FootSize
- Appearance: SkinTone, SkinUndertone
- Generation options: Seed, HasHairSlot, MergeVertices

### Stage 2: PRNG Initialization

**File**: `rng.go`

A custom `splitmix64` PRNG is created from the seed. This algorithm was chosen for:
- Deterministic output across Go versions
- No reliance on `math/rand` internal state
- Fast, high-quality random numbers

```go
rng := newSplitmix64(p.Seed)
```

### Stage 3: Body Layout Computation

**Files**: `basemesh.go`, `transforms.go`

The `bodyLayout` struct defines the geometric positions and sizes of all body parts:

```go
type bodyLayout struct {
    totalHeight float32
    headCenter  Vec3
    headRX, headRY, headRZ float32
    neckBottom, neckTop Vec3
    // ... 50+ fields for all body parts
}
```

Transform functions in `transforms.go` modify the base layout according to parameters:

| Function | Effect |
|----------|--------|
| `applySpeciesTransform` | Species-specific proportions |
| `applyHeightTransform` | Scale to target height |
| `applyBuildTransform` | Muscular/thin body shape |
| `applyAgeTransform` | Age-based proportions |
| `applyPostureTransform` | Posture adjustments |
| `applyFaceTransform` | Facial geometry |
| `applyLimbTransform` | Limb length/thickness |

### Stage 4: Mesh Generation

**Files**: `basemesh.go`, `primitive.go`

The `buildMesh` function assembles geometric primitives into a unified mesh:

```go
func buildMesh(layout bodyLayout, key string, opts buildOptions) *Mesh
```

#### Primitive Types

| Primitive | Function | Usage |
|-----------|----------|-------|
| Ellipsoid | `generateEllipsoid` | Head, torso segments |
| Cylinder | `generateCylinder` | Arms, legs, neck, fingers |
| Box | `generateBox` | Hands, feet, nails |
| Face | `generateFaceMesh` | Facial features |

#### Mesh Builder

The internal `meshBuilder` type accumulates vertices and indices:

```go
type meshBuilder struct {
    verts   []Vertex
    indices []uint32
}

func (m *meshBuilder) append(v []Vertex, idx []uint32)
func (m *meshBuilder) build(key string) *Mesh
```

### Stage 5: Post-Processing (Optional)

**File**: `merge.go`

If `Params.MergeVertices` is true, nearby vertices are merged to create seamless topology:

```go
mesh = MergeNearbyVertices(mesh, epsilon)
```

## Key Data Structures

### Vertex

**File**: `mesh.go`

```go
type Vertex struct {
    Position     Vec3  // X, Y, Z position
    Normal       Vec3  // Surface normal
    UV           Vec2  // Texture coordinates
    Color        Color // RGBA vertex color
    Tangent      Vec4  // Tangent vector (for normal mapping)
    JointIds     Vec4  // Skeleton joint IDs (for skinning)
    JointWeights Vec4  // Skinning weights
}
```

### Mesh

```go
type Mesh struct {
    Vertices []Vertex
    Indices  []uint32
    Key      string  // Cache key for Kaiju engine
}
```

### Coordinate System

- **Y-up**: Ground at Y=0, head at Y≈1.8
- **Right-handed**: +X is right, +Z is forward
- **Units**: Meters

## Advanced Features

### Skeleton Generation

**File**: `skeleton.go`

52-joint humanoid skeleton matching body layout:

```
Root → Hips → Spine(4) → Neck → Head → HeadTop
            ↳ LeftLeg(4), RightLeg(4)
       Shoulders → UpperArm → Forearm → Hand → Fingers(5×3)
```

### Skinning

**File**: `skinning.go`

Vertex weights for skeletal animation (max 4 influences per vertex).

### Morph Targets

**File**: `morph.go`

19 blend shapes for facial expressions and body modifications.

### LOD Generation

**File**: `lod.go`

Three detail levels via edge-collapse decimation:
- LOD0: 100% detail
- LOD1: 50% detail
- LOD2: 25% detail

### Export Formats

| Format | File | Function |
|--------|------|----------|
| OBJ | `export_obj.go` | `ExportOBJ`, `ExportOBJWithMTL` |
| glTF | `export_gltf.go` | `ExportGLTF`, `ExportGLB` |
| Binary | `stream.go` | `BinaryMeshWriter.WriteMesh` |
| Animated glTF | `export_animated.go` | `ExportAnimatedGLTF` |

## Extension Points

### Adding a New Species

1. Add enum value to `params.go`:
   ```go
   const (
       SpeciesHuman Species = iota
       // ... existing
       SpeciesNewType  // Add at end!
   )
   ```

2. Update `Validate()` in `params.go` to include new species

3. Add transform case in `transforms.go`:
   ```go
   func applySpeciesTransform(l *bodyLayout, s Species) {
       switch s {
       case SpeciesNewType:
           // Apply proportions
       }
   }
   ```

### Adding a New Parameter

1. Add field to `Params` struct in `params.go`
2. Add validation in `Validate()` method
3. Update mesh key format string in `generator.go`
4. Implement transform function in `transforms.go`
5. Call transform in `computeBodyLayout()`

### Adding a New Body Part

1. Add layout fields to `bodyLayout` struct in `basemesh.go`
2. Initialize in `defaultBodyLayout()`
3. Scale in `scaleAll()` and `scaleHeight()` helpers
4. Generate geometry in `buildMesh()`

## Batch Processing

**File**: `batch.go`

Worker pool for parallel generation:

```go
bg := NewBatchGenerator(4)  // 4 workers
results := bg.GenerateBatch(ctx, paramsList)
```

## Caching

**File**: `cache.go`

LRU cache for repeated generation with same parameters:

```go
cg := NewCachedGenerator(100)  // 100 entries
mesh, _ := cg.Generate(params)  // Cached if same params
```

## Performance Characteristics

- Generation time: ~5-10ms per character
- Memory per mesh: ~0.94MB (after optimization)
- Thread-safe: Generator is stateless

## File Map

```
unpeople/
├── generator.go      # Public API, mesh key generation
├── params.go         # Parameter types and validation
├── basemesh.go       # Body layout, mesh assembly
├── transforms.go     # Parameter-to-geometry transforms
├── primitive.go      # Geometric primitives
├── mesh.go           # Vertex/Mesh types, utilities
├── rng.go            # Deterministic PRNG
├── skeleton.go       # Skeleton generation
├── skinning.go       # Vertex skinning weights
├── morph.go          # Morph targets
├── lod.go            # LOD generation
├── merge.go          # Vertex merging
├── slots.go          # Attachment slots
├── texture.go        # Procedural textures
├── normalmap.go      # Normal map generation
├── material.go       # Material definitions
├── atlas.go          # UV atlas generation
├── cache.go          # LRU caching
├── batch.go          # Parallel generation
├── stream.go         # Binary format
├── export_obj.go     # OBJ export
├── export_gltf.go    # glTF export
├── export_animated.go # Animated glTF
└── bvh.go            # BVH animation import
```
