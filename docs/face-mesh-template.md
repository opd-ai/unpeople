# Face Mesh Template Design

This document defines the vertex group positions and parameter mappings for advanced facial morphing in Phase 2.

## Problem Statement

Currently, facial features are implemented by adjusting head ellipsoid radii only (`transforms.go:404-465`). This results in:
- Jaw prominence indistinguishable from jaw roundness
- No brow ridge representation
- Absent cheekbone and nasal structure
- All facial variation limited to 3 dimensions (head RX/RY/RZ)

## Design Goals

1. **Distinct facial regions** with independent vertex groups
2. **Parameter-driven deformation** mapping FaceShape, Jaw, Brow enums to vertex displacements
3. **Species-appropriate defaults** — Orc jaws differ from Elf jaws
4. **Overlay compatibility** — face mesh sits on head ellipsoid

## Coordinate System

All coordinates are relative to `headCenter` (origin at ellipsoid center):
- **X-axis**: Lateral (positive = character's right)
- **Y-axis**: Vertical (positive = up, toward crown)
- **Z-axis**: Depth (positive = forward, toward viewer)

Reference head dimensions for coordinate normalization:
```
defaultHeadRX = 0.090  // Ear to ear
defaultHeadRY = 0.115  // Chin to crown
defaultHeadRZ = 0.090  // Nose to occiput
```

## Vertex Groups

### 1. Jaw Region

The jaw region forms the lower third of the face, controlling chin shape and jawline.

**Vertex positions (relative to headCenter):**
```
         +Y (crown)
           │
           │       [brow]
           │    ╱─────────╲
           │   │   nose    │
           │   │     ∆     │
    -X ────┼───│  ╱   ╲   │─── +X
  (left)   │   │ (mouth)  │  (right)
           │   ╲___jaw___╱
           │      chin
           │
         -Y (chin)
```

**Jaw vertices (12 vertices):**
```go
// Relative to headCenter, all values in metres
jawVertices := []Vec3{
    // Chin (center, lower)
    {0.000, -0.105, 0.070},         // chin_center
    {-0.030, -0.100, 0.065},        // chin_left
    {0.030, -0.100, 0.065},         // chin_right
    
    // Jaw corners (widest lateral extent)
    {-0.080, -0.060, 0.020},        // jaw_corner_left
    {0.080, -0.060, 0.020},         // jaw_corner_right
    
    // Mandible line (connects corners to chin)
    {-0.055, -0.085, 0.045},        // mandible_left
    {0.055, -0.085, 0.045},         // mandible_right
    
    // Jaw back (near ear attachment)
    {-0.085, -0.030, -0.010},       // jaw_back_left
    {0.085, -0.030, -0.010},        // jaw_back_right
    
    // Under-chin (connects to neck)
    {0.000, -0.095, 0.000},         // under_chin_center
    {-0.040, -0.090, -0.010},       // under_chin_left
    {0.040, -0.090, -0.010},        // under_chin_right
}
```

**Parameter mapping (Jaw enum → displacement vectors):**

| Jaw Value | Chin Z | Chin Y | Corner X | Description |
|-----------|--------|--------|----------|-------------|
| JawProminent | +0.015 | -0.008 | +0.005 | Forward, dropped chin; wider corners |
| JawAverage | 0 | 0 | 0 | Default |
| JawSubtle | -0.008 | +0.005 | -0.008 | Retracted, lifted chin; narrower |
| JawAngular | +0.005 | 0 | +0.012 | Sharper corner definition |
| JawRounded | 0 | +0.003 | -0.005 | Softer corners |

### 2. Brow Region

The brow ridge forms the upper orbit area, controlling forehead slope and eye socket depth.

**Brow vertices (10 vertices):**
```go
browVertices := []Vec3{
    // Brow ridge (above eyes)
    {0.000, 0.050, 0.085},          // brow_center (glabella)
    {-0.045, 0.048, 0.082},         // brow_left
    {0.045, 0.048, 0.082},          // brow_right
    {-0.070, 0.040, 0.070},         // brow_outer_left
    {0.070, 0.040, 0.070},          // brow_outer_right
    
    // Forehead transition
    {0.000, 0.080, 0.075},          // forehead_center
    {-0.055, 0.075, 0.065},         // forehead_left
    {0.055, 0.075, 0.065},          // forehead_right
    
    // Temple area
    {-0.082, 0.035, 0.045},         // temple_left
    {0.082, 0.035, 0.045},          // temple_right
}
```

**Parameter mapping (Brow enum → displacement vectors):**

| Brow Value | Ridge Z | Ridge Y | Temple Z | Description |
|------------|---------|---------|----------|-------------|
| BrowHeavy | +0.012 | -0.005 | +0.005 | Protruding, lowered ridge |
| BrowNormal | 0 | 0 | 0 | Default |
| BrowLight | -0.008 | +0.003 | -0.003 | Retracted, higher ridge |
| BrowArched | +0.005 | +0.008 | 0 | Center elevated (arch) |

### 3. Cheekbone Region

Cheekbones define the mid-face width and orbital structure.

**Cheekbone vertices (8 vertices):**
```go
cheekboneVertices := []Vec3{
    // Upper cheekbone (near eye corner)
    {-0.078, 0.010, 0.055},         // cheek_high_left
    {0.078, 0.010, 0.055},          // cheek_high_right
    
    // Mid cheekbone (widest point)
    {-0.088, -0.015, 0.035},        // cheek_mid_left
    {0.088, -0.015, 0.035},         // cheek_mid_right
    
    // Lower cheek (toward jaw)
    {-0.075, -0.045, 0.045},        // cheek_low_left
    {0.075, -0.045, 0.045},         // cheek_low_right
    
    // Inner cheek (nasolabial area)
    {-0.040, -0.035, 0.070},        // cheek_inner_left
    {0.040, -0.035, 0.070},         // cheek_inner_right
}
```

### 4. Nose Region

The nose bridge and tip, primarily affecting profile silhouette.

**Nose vertices (7 vertices):**
```go
noseVertices := []Vec3{
    // Bridge (between eyes)
    {0.000, 0.030, 0.088},          // nose_bridge_top
    {0.000, 0.000, 0.095},          // nose_bridge_mid
    
    // Tip
    {0.000, -0.025, 0.105},         // nose_tip
    
    // Nostrils
    {-0.018, -0.035, 0.090},        // nostril_left
    {0.018, -0.035, 0.090},         // nostril_right
    
    // Nose sides
    {-0.025, -0.010, 0.085},        // nose_side_left
    {0.025, -0.010, 0.085},         // nose_side_right
}
```

### 5. FaceShape Compound Mapping

FaceShape affects multiple vertex groups simultaneously:

| FaceShape | Jaw X | Cheek X | Brow X | Chin Y | Description |
|-----------|-------|---------|--------|--------|-------------|
| Oval | 0 | 0 | 0 | 0 | Balanced proportions |
| Round | +0.005 | +0.010 | +0.005 | +0.005 | Wider, shorter |
| Square | +0.010 | +0.008 | +0.008 | -0.005 | Angular, broad |
| Heart | -0.005 | -0.003 | +0.008 | +0.010 | Wide brow, narrow chin |
| Diamond | -0.008 | +0.012 | -0.005 | +0.008 | Wide cheeks, narrow forehead/chin |
| Oblong | -0.008 | -0.005 | -0.005 | -0.012 | Narrow, elongated |

## Implementation Structure

```go
// faceRegion identifies a facial vertex group
type faceRegion int

const (
    regionJaw faceRegion = iota
    regionBrow
    regionCheekbone
    regionNose
)

// faceVertex stores a vertex's position and region membership
type faceVertex struct {
    basePosition Vec3       // Default position relative to headCenter
    region       faceRegion // Which region this vertex belongs to
    weights      [4]float32 // Blend weights for compound shapes
}

// generateFaceMesh creates facial geometry overlaid on the head ellipsoid
func generateFaceMesh(
    layout bodyLayout,
    fs FaceShape,
    j Jaw,
    br Brow,
) ([]Vertex, []uint32) {
    // 1. Get base vertex positions (scaled to current head size)
    baseVerts := scaleFaceVertices(baseFaceTemplate, layout.headRX, layout.headRY, layout.headRZ)
    
    // 2. Apply FaceShape compound displacement
    applyFaceShapeDisplacement(baseVerts, fs)
    
    // 3. Apply Jaw displacement
    applyJawDisplacement(baseVerts, j)
    
    // 4. Apply Brow displacement
    applyBrowDisplacement(baseVerts, br)
    
    // 5. Offset all vertices to headCenter
    translateToHead(baseVerts, layout.headCenter)
    
    // 6. Generate triangle indices for face surface
    idxs := generateFaceSurfaceIndices(baseVerts)
    
    return baseVerts, idxs
}
```

## Species-Specific Base Templates

Different species have different default face templates:

| Species | Brow Offset | Jaw Offset | Nose Offset | Notes |
|---------|-------------|------------|-------------|-------|
| Human | 0 | 0 | 0 | Default template |
| Elf | 0 | -0.3 | +0.2 | Refined, delicate |
| Dwarf | +0.5 | +0.3 | +0.1 | Heavy, broad |
| Orc | +0.8 | +0.5 | -0.2 | Prognathic, pronounced |
| Goblin | -0.3 | +0.2 | +0.4 | Pointed, prominent nose |

Values are multipliers on the displacement vectors.

## Integration with Existing Head Ellipsoid

The face mesh sits 2mm above the head ellipsoid surface to avoid z-fighting:
```go
const faceOffset = 0.002 // metres above ellipsoid surface
```

For each face vertex:
1. Calculate position on head ellipsoid (ray from headCenter)
2. Push vertex outward along normal by `faceOffset`
3. Apply parameter-driven displacements

## Testing Strategy

1. **Parameter iteration tests**: Generate all FaceShape × Jaw × Brow combinations
2. **Mesh validity**: All face mesh indices within bounds
3. **Visual inspection**: Known parameter combinations produce expected shapes
4. **Determinism**: Same parameters → identical face mesh

---

*Design document created: 2026-04-03*
*Prerequisite for: Phase 2 Advanced Facial Morphing*
