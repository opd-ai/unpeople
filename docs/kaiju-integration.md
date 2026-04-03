# Kaiju Engine Integration

This document describes how to integrate the unpeople humanoid mesh generator
with the Kaiju game engine.

## Prerequisites

- Go 1.21 or later
- Kaiju engine installed (`kaijuengine.com/rendering` package available)

## Building with Kaiju Support

Build your project with the `kaiju` build tag to enable direct Kaiju integration:

```bash
go build -tags kaiju ./...
```

Without the tag, the `kaiju` package provides a stub implementation that returns
`*unpeople.Mesh` instead of `*rendering.Mesh`.

## Basic Usage

```go
package main

import (
    "log"
    
    "github.com/opd-ai/unpeople"
    "github.com/opd-ai/unpeople/kaiju"
)

func main() {
    // Create a Kaiju-integrated generator
    gen := kaiju.NewKaijuGenerator()
    
    // Configure character parameters
    params := unpeople.DefaultParams()
    params.Seed = 42
    params.Species = unpeople.SpeciesElf
    params.Height = unpeople.HeightTall
    params.Build = unpeople.BuildAthletic
    
    // Generate mesh (returns *rendering.Mesh when built with -tags kaiju)
    mesh, err := gen.Generate(params)
    if err != nil {
        log.Fatal(err)
    }
    
    // Use mesh directly with Kaiju's rendering pipeline
    // mesh is ready to be added to a scene
}
```

## Converting Existing Meshes

If you have an `*unpeople.Mesh` from the standard generator, you can convert it:

```go
import "github.com/opd-ai/unpeople/kaiju"

// Convert unpeople mesh to Kaiju mesh
kaijuMesh := kaiju.ToKaijuMesh(unpeopleMesh)

// Convert back if needed
unpeopleMesh := kaiju.ToUnpeopleMesh(kaijuMesh)
```

## Vertex Layout Compatibility

The unpeople `Vertex` struct is designed to be layout-compatible with Kaiju's
`rendering.Vertex`:

| Field | Type | Description |
|-------|------|-------------|
| Position | Vec3 (float32×3) | World-space position |
| Normal | Vec3 (float32×3) | Surface normal |
| Tangent | Vec4 (float32×4) | Tangent vector with handedness |
| UV0 | Vec2 (float32×2) | Primary texture coordinates |
| UV1 | Vec2 (float32×2) | Secondary texture coordinates |
| Color | Color (float32×4) | Vertex color (RGBA) |
| JointIds | Vec4i (int32×4) | Skinning joint indices |
| JointWeights | Vec4 (float32×4) | Skinning weights |
| MorphTarget | Vec3 (float32×3) | Morph target delta |

## Mesh Caching

For performance, consider using the cached generator:

```go
import "github.com/opd-ai/unpeople"

// Create a cached generator with LRU eviction
cachedGen := unpeople.NewCachedGenerator(100) // Cache up to 100 meshes

// Then convert results to Kaiju format
mesh, _ := cachedGen.Generate(params)
kaijuMesh := kaiju.ToKaijuMesh(mesh)
```

## LOD Support

Generate multiple levels of detail:

```go
gen := unpeople.NewGenerator()
lodResult, _ := gen.GenerateWithLOD(params)

// lodResult.LODSet.Meshes[0] - Full detail
// lodResult.LODSet.Meshes[1] - 50% triangles
// lodResult.LODSet.Meshes[2] - 25% triangles

// Convert each LOD level as needed
for level, lodMesh := range lodResult.LODSet.Meshes {
    kaijuMesh := kaiju.ToKaijuMesh(lodMesh.Mesh)
    // Register with Kaiju's LOD system
}
```

## Performance Considerations

1. **Mesh Generation**: Typically completes in <10ms for a single character
2. **Conversion Overhead**: The `ToKaijuMesh` conversion involves copying vertex
   data, which is O(n) where n is the vertex count
3. **Caching**: Use `CachedGenerator` when generating many characters with
   potentially repeated parameter combinations

## Troubleshooting

### Build fails with "package kaijuengine.com/rendering not found"

Ensure the Kaiju engine is properly installed and your Go workspace is
configured correctly. The `kaijuengine.com/rendering` package must be
importable.

### Stub implementation returns *unpeople.Mesh

You're building without the `-tags kaiju` flag. Add it to your build command:

```bash
go build -tags kaiju .
```

### Mesh appears incorrect in Kaiju

Verify the mesh using the OBJ export for debugging:

```go
f, _ := os.Create("debug.obj")
unpeople.ExportOBJ(f, mesh, "debug")
f.Close()
```

Then inspect in a 3D tool like Blender to verify geometry.
