# unpeople

[![CI](https://github.com/opd-ai/unpeople/actions/workflows/ci.yml/badge.svg)](https://github.com/opd-ai/unpeople/actions/workflows/ci.yml)

Deterministic procedural generation of humanoid character meshes for Go game engines.

## Overview

`unpeople` generates 3D humanoid meshes from a seed and parameter set. Given identical inputs, it always produces bit-identical output, making it ideal for:

- Populating open worlds with varied NPC humanoids
- Generating deterministic characters from compact seed data
- Prototyping character silhouettes across fantasy species

The output is layout-compatible with the [Kaiju game engine](https://kaijuengine.com) but works with any Go-based renderer.

## Installation

```bash
go get github.com/opd-ai/unpeople
```

Requires Go 1.21+. Zero external dependencies (stdlib only).

## Quick Start

```go
package main

import (
    "log"
    "github.com/opd-ai/unpeople"
)

func main() {
    gen := unpeople.NewGenerator()
    
    params := unpeople.DefaultParams()
    params.Seed = 42
    params.Species = unpeople.SpeciesElf
    params.Height = unpeople.HeightTall
    params.Build = unpeople.BuildAthletic
    
    mesh, err := gen.Generate(params)
    if err != nil {
        log.Fatal(err)
    }
    
    log.Printf("Generated: %d vertices, %d triangles",
        len(mesh.Vertices), len(mesh.Indices)/3)
}
```

## Features

### Species (10 types)
Human, Elf, Dwarf, Gnome, Halfling, Goblin, Kobold, Orc, Troll, Ogre

### Parameters (20+ customization options)
- **Body**: Height, Build, Proportions, Phenotype
- **Age & Posture**: 8 age stages, 4 posture types
- **Face**: Shape, Jaw, Brow, Ears
- **Body Details**: Shoulder/Hip width, Limb/Neck length
- **Hands & Feet**: Size variants, finger length
- **Appearance**: 8 skin tones × 3 undertones

### Export Formats
- **OBJ** — Wavefront OBJ with materials
- **glTF 2.0** — JSON with embedded buffers
- **GLB** — Binary glTF (single file)
- **Binary** — Compact UNPM format for fast loading

### Advanced Features
- **Skeleton** — 56-joint hierarchy for animation
- **Skinning** — Vertex weights for skeletal deformation
- **Morph Targets** — 19 blend shapes (facial expressions, body morphs)
- **LOD Generation** — 3 detail levels (100%/50%/25%)
- **Batch Processing** — Parallel generation with worker pools
- **Caching** — LRU cache for repeated generation
- **Textures** — Procedural skin textures (freckles, blemishes, age spots)
- **Normal Maps** — Musculature detail based on build

## CLI Tool

```bash
# Generate OBJ from seed
unpeopled -seed 12345 > character.obj

# Generate glTF
unpeopled -seed 12345 -format gltf > character.gltf

# Generate from JSON params
echo '{"seed": 42, "species": 1}' | unpeopled > elf.obj
```

## REST API Server

```bash
# Start server
unpeople-server -addr :8080

# Generate via HTTP
curl -X POST http://localhost:8080/generate \
  -H "Content-Type: application/json" \
  -d '{"seed": 42, "species": 2}' > dwarf.obj
```

## Documentation

- [REST API Reference](docs/rest-api.md)
- [Kaiju Integration](docs/kaiju-integration.md)
- [Face Mesh Template](docs/face-mesh-template.md)
- [Vertex Merging](docs/vertex-merging.md)

## License

MIT