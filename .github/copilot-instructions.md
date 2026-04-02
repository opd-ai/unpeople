# Project Overview

`unpeople` is a Go library for deterministic procedural generation of humanoid character meshes. Given a seed and a descriptive parameter set (species, height, build, age, facial features, etc.), the `Generator` always produces an identical 3D mesh, making it ideal for open-world games where characters must be reproducible from a saved seed. The output `Mesh` type is layout-compatible with the Kaiju game engine's `rendering.Vertex` / `rendering.Mesh` structures for direct integration with the Kaiju rendering pipeline.

The target audience is game developers building procedural content for the Kaiju engine or similar Go-based game engines. Typical use cases include populating open worlds with varied NPC humanoids, generating deterministic character appearances from compact seed data, and prototyping character silhouettes across fantasy species (Human, Elf, Dwarf, Gnome, Halfling, Goblin, Kobold, Orc, Troll, Ogre). The library is pure Go with zero external dependencies—it uses only the standard library and a custom SplitMix64 PRNG to guarantee cross-version determinism.

## Technical Stack
- **Primary Language**: Go 1.21+
- **Frameworks**: None — stdlib only; no external module dependencies
- **Testing**: Go's built-in `testing` package; table-driven tests and benchmarks in `generator_test.go`
- **Build/Deploy**: `go build ./...` to build, `go test ./...` to test, `go vet ./...` to lint; no Makefile or CI config required beyond standard Go tooling
- **License**: MIT

## Code Assistance Guidelines

1. **Zero External Dependencies**: This project intentionally has no third-party dependencies (`go.mod` lists only the module path and Go version). Never add external modules. All functionality must be implemented using the Go standard library or package-local code. The custom `splitmix64` PRNG in `rng.go` exists specifically to avoid coupling determinism to `math/rand` internals.

2. **Determinism Is Sacred**: The same `Params` (including `Seed`) must always produce a bit-identical `Mesh`. Never introduce non-deterministic operations (e.g., map iteration order, goroutine scheduling, `math/rand` global state, time-based values) into the generation path. The PRNG (`splitmix64`) must be created fresh per `Generate` call from the seed and used sequentially.

3. **Kaiju Vertex Layout Compatibility**: The `Vertex` struct in `mesh.go` mirrors `kaijuengine.com/rendering.Vertex` field-for-field. Never reorder, rename, or add fields to `Vertex` without verifying compatibility with the Kaiju engine's expected memory layout. All vertices must set `Color` to `ColorGray` (mid-grey, fully opaque) as the default material colour.

4. **Enum-Based Parameter Design**: All character traits (`Species`, `Height`, `Build`, `Proportions`, `Phenotype`, `Age`, `Posture`, facial features, body dimensions) are `int`-typed enums with `iota` constants. When adding new enum values, always append them to the end of the existing `const` block to preserve existing ordinal values. Update `Validate()` in `params.go` to cover the new range. Include the new value in the mesh key format string in `generator.go`.

5. **Primitive-Based Geometry Assembly**: The mesh is built from geometric primitives—`generateEllipsoid`, `generateCylinder`, and `generateBox` in `primitive.go`. Body parts are positioned and sized via the `bodyLayout` struct (`basemesh.go`). Transformations are applied as multiplicative scale factors on layout fields in `transforms.go`. When adding new body parts, add layout fields to `bodyLayout`, update `defaultBodyLayout()`, update both `scaleAll` and `scaleHeight` helpers, and append geometry in `buildMesh`.

6. **Table-Driven Tests for All Enum Ranges**: Every enum type must have a test that iterates through all valid values and verifies `Generate` succeeds. Follow the pattern in `TestAllSpecies`, `TestAllHeights`, `TestAllBuilds`, and `TestAllAges`. When adding new parameters or enum values, add corresponding test coverage in `generator_test.go`.

7. **Complete Feature Implementations**: Always prefer completing the full implementation of any feature rather than leaving partial or placeholder code. When a complete implementation is not feasible, insert clear inline `TODO` comments describing what remains, why it was deferred, and any known constraints (e.g., `// TODO: Implement retry logic once the error categorization schema is finalized`). Never leave code in a silently incomplete state.

## Project Context

- **Domain**: Procedural character generation for fantasy game engines. Key concepts include species-based body scaling, phenotype dimorphism, age-based proportion changes, posture deformation, and facial feature parameterization. All measurements are in metres in a right-handed Y-up coordinate system (feet at Y≈0, X lateral, Z forward).

- **Architecture**: Single-package library (`package unpeople`) with a stateless `Generator` that accepts a `Params` struct and returns a `*Mesh`. The generation pipeline is: validate params → create seeded PRNG → `computeBodyLayout` (applies all transform functions) → `buildMesh` (assembles primitives into vertex/index buffers). The `Generator` is safe for concurrent use.

- **Key Directories**:
  - Root (`*.go`): All library source — `params.go` (enums & validation), `generator.go` (public API), `basemesh.go` (body layout & mesh assembly), `transforms.go` (parameter-to-geometry transforms), `primitive.go` (geometric primitives), `mesh.go` (Vertex/Mesh types & helpers), `rng.go` (deterministic PRNG)
  - `example/`: Reference CLI binary demonstrating parameter variety and mesh statistics output
  - `ROADMAP.md`: Phased development plan (Phase 1 complete; Phases 2–6 pending)
  - `GAPS.md`: Known limitations, missing features, and technical debt

- **Configuration**: No runtime configuration files or environment variables. All behaviour is controlled through `Params` struct fields. The mesh key format string in `generator.go` must encode every geometry-affecting parameter for Kaiju cache correctness.

- **Known Limitations** (see `GAPS.md`): Body is assembled from disconnected primitives with visible seams; facial features only adjust head ellipsoid radii (no dedicated face mesh); `JointIds`/`JointWeights`/`MorphTarget` fields are zeroed (no animation support yet); hands are flat boxes without finger geometry; no UV atlas for texturing.

## Quality Standards

- **Testing**: Maintain comprehensive test coverage using Go's built-in testing package. All tests live in `generator_test.go` as an external test package (`package unpeople_test`). Write table-driven tests for parameter validation and enum iteration tests for all parameter types. Include a performance test ensuring `Generate` completes in under 100 ms. Run benchmarks with `go test -bench=.` to catch performance regressions.

- **Code Review Criteria**: Changes must preserve determinism (same seed → same mesh). New enum values must not shift existing ordinals. The `scaleAll`/`scaleHeight`/`scaleLimbs` helpers must be updated in lockstep when adding layout fields. All index values in generated meshes must be within vertex buffer bounds (validated by `TestMeshIsValid`).

- **Documentation**: Every exported type, function, and constant must have a Go doc comment. Internal functions should have brief comments explaining their geometric purpose. The `ROADMAP.md` and `GAPS.md` files should be updated when features are completed or new limitations are discovered.

## Networking Best Practices (for Go projects)

When declaring network variables, always use interface types:
- Never use `net.UDPAddr`, `net.IPAddr`, or `net.TCPAddr`. Use `net.Addr` only instead.
- Never use `net.UDPConn`, use `net.PacketConn` instead
- Never use `net.TCPConn`, use `net.Conn` instead
- Never use `net.UDPListener` or `net.TCPListener`, use `net.Listener` instead
- Never use a type switch or type assertion to convert from an interface type to a concrete type. Use the interface methods instead.

This approach enhances testability and flexibility when working with different network implementations or mocks.
