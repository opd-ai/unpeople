// Package main provides the unpeopled CLI tool for generating humanoid meshes.
//
// Usage:
//
//	unpeopled < params.json > character.obj
//	unpeopled -format obj < params.json > character.obj
//	unpeopled -format binary < params.json > character.unpm
//	unpeopled -format lod < params.json > character_lod.unpm
//	unpeopled -seed 12345 > character.obj
//
// The tool reads a JSON parameters object from stdin (or uses defaults with
// -seed flag) and writes the generated mesh to stdout. Supported output formats
// are "obj" (Wavefront OBJ) and "binary" (UNPM binary format).
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/opd-ai/unpeople"
)

var (
	formatFlag = flag.String("format", "obj", "Output format: obj, gltf, glb, binary, lod")
	seedFlag   = flag.Int64("seed", 0, "Random seed (uses stdin params if 0)")
	lodFlag    = flag.Int("lod", 0, "LOD level for lod format: 0, 1, or 2")
	poseFlag   = flag.String("pose", "tpose", "Skeleton pose: tpose or apose")
	quietFlag  = flag.Bool("q", false, "Quiet mode: suppress stderr messages")
	helpFlag   = flag.Bool("h", false, "Show help")
)

func main() {
	flag.Parse()

	if *helpFlag {
		printUsage()
		os.Exit(0)
	}

	params, err := loadParams()
	if err != nil {
		fatal("Failed to load parameters: %v", err)
	}

	if err := generate(params, os.Stdout); err != nil {
		fatal("Generation failed: %v", err)
	}
}

func printUsage() {
	fmt.Fprintln(os.Stderr, `unpeopled - Procedural humanoid mesh generator

Usage:
  unpeopled [options] < params.json > output.obj
  unpeopled -seed 12345 > output.obj
  unpeopled -seed 12345 -format gltf > output.gltf
  unpeopled -seed 12345 -format glb > output.glb
  unpeopled -seed 12345 -pose apose -format gltf > output_apose.gltf
  echo '{"seed":42}' | unpeopled > output.obj

Options:`)
	flag.PrintDefaults()
	fmt.Fprintln(os.Stderr, `
Input:
  Reads JSON parameters from stdin. If -seed is provided, uses default
  parameters with the specified seed instead of reading from stdin.

  Example params.json:
  {
    "seed": 12345,
    "species": 0,
    "height": 2,
    "build": 2,
    "age": 2,
    "skeletonpose": 1
  }

Output Formats:
  obj     - Wavefront OBJ text format (default)
  gltf    - glTF 2.0 JSON with embedded buffers
  glb     - glTF 2.0 Binary format (single file)
  binary  - UNPM binary format (compact, fast loading)
  lod     - Binary format with LOD level selection (-lod flag)

Skeleton Pose:
  tpose   - T-pose with arms horizontal (default)
  apose   - A-pose with arms angled ~45° down (better for animation)

Species Values:
  0=Human, 1=Elf, 2=Dwarf, 3=Gnome, 4=Halfling,
  5=Goblin, 6=Kobold, 7=Orc, 8=Troll, 9=Ogre

Height Values:
  0=Giant, 1=Tall, 2=Medium, 3=Short, 4=Tiny

Build Values:
  0=Muscular, 1=Athletic, 2=Average, 3=Lean, 4=Stocky, 5=Fragile

Age Values:
  0=Decrepit, 1=Elderly, 2=Old, 3=Adult, 4=Youth, 5=Teen, 6=Child, 7=Toddler`)
}

// readParamsFromStdin reads and parses JSON parameters from stdin.
func readParamsFromStdin() (unpeople.Params, error) {
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return unpeople.Params{}, fmt.Errorf("reading stdin: %w", err)
	}

	// If empty, use defaults
	if len(data) == 0 {
		info("No input provided, using default parameters")
		p := unpeople.DefaultParams()
		applyPoseFlag(&p)
		return p, nil
	}

	return parseParamsJSON(data)
}

// parseParamsJSON unmarshals JSON data into Params and applies defaults.
func parseParamsJSON(data []byte) (unpeople.Params, error) {
	var p unpeople.Params
	if err := json.Unmarshal(data, &p); err != nil {
		return unpeople.Params{}, fmt.Errorf("parsing JSON: %w", err)
	}

	// Apply defaults for zero values
	if p.Seed == 0 {
		defaults := unpeople.DefaultParams()
		p.Seed = defaults.Seed
	}

	// Apply pose flag if specified (CLI flag overrides JSON)
	applyPoseFlag(&p)

	return p, nil
}

func loadParams() (unpeople.Params, error) {
	// If seed flag is set, use defaults with that seed
	if *seedFlag != 0 {
		p := unpeople.DefaultParams()
		p.Seed = *seedFlag
		applyPoseFlag(&p)
		return p, nil
	}

	return readParamsFromStdin()
}

// applyPoseFlag applies the -pose flag to the params if specified.
func applyPoseFlag(p *unpeople.Params) {
	switch *poseFlag {
	case "apose", "a-pose", "a":
		p.SkeletonPose = unpeople.SkeletonPoseAPose
	case "tpose", "t-pose", "t":
		p.SkeletonPose = unpeople.SkeletonPoseTPose
		// Default: leave as-is (from JSON or default)
	}
}

func generate(p unpeople.Params, w io.Writer) error {
	g := unpeople.NewGenerator()

	switch *formatFlag {
	case "obj":
		return generateOBJ(g, p, w)
	case "gltf":
		return generateGLTF(g, p, w)
	case "glb":
		return generateGLB(g, p, w)
	case "binary":
		return generateBinary(g, p, w)
	case "lod":
		return generateLOD(g, p, w)
	default:
		return fmt.Errorf("unknown format: %s", *formatFlag)
	}
}

func generateOBJ(g *unpeople.Generator, p unpeople.Params, w io.Writer) error {
	mesh, err := g.Generate(p)
	if err != nil {
		return err
	}

	info("Generated mesh: %d vertices, %d triangles", len(mesh.Vertices), len(mesh.Indices)/3)
	return unpeople.ExportOBJ(w, mesh, "character")
}

func generateGLTF(g *unpeople.Generator, p unpeople.Params, w io.Writer) error {
	mesh, err := g.Generate(p)
	if err != nil {
		return err
	}

	info("Generated mesh: %d vertices, %d triangles", len(mesh.Vertices), len(mesh.Indices)/3)
	return unpeople.ExportGLTFDefault(w, mesh)
}

func generateGLB(g *unpeople.Generator, p unpeople.Params, w io.Writer) error {
	mesh, err := g.Generate(p)
	if err != nil {
		return err
	}

	info("Generated mesh: %d vertices, %d triangles", len(mesh.Vertices), len(mesh.Indices)/3)
	return unpeople.ExportGLB(w, mesh, unpeople.DefaultGLTFOptions())
}

func generateBinary(g *unpeople.Generator, p unpeople.Params, w io.Writer) error {
	bw := unpeople.NewBinaryMeshWriter(w)
	result, err := g.GenerateStream(p, bw)
	if err != nil {
		return err
	}

	info("Generated mesh: %d vertices, %d triangles", result.VertexCount, result.TriangleCount)
	return nil
}

// validateLODLevel checks if the LOD level is valid.
func validateLODLevel(level int) error {
	if level < 0 || level >= int(unpeople.LODCount) {
		return fmt.Errorf("invalid LOD level: %d (must be 0, 1, or 2)", level)
	}
	return nil
}

// writeVertices writes all vertices to the binary writer.
func writeVertices(bw *unpeople.BinaryMeshWriter, vertices []unpeople.Vertex) error {
	for _, v := range vertices {
		if err := bw.WriteVertex(v); err != nil {
			return err
		}
	}
	return nil
}

// writeIndices writes all indices to the binary writer.
func writeIndices(bw *unpeople.BinaryMeshWriter, indices []uint32) error {
	for _, idx := range indices {
		if err := bw.WriteIndex(idx); err != nil {
			return err
		}
	}
	return nil
}

// writeLODMesh writes the LOD mesh to the writer in binary format.
func writeLODMesh(lodMesh *unpeople.LODMesh, w io.Writer) error {
	bw := unpeople.NewBinaryMeshWriter(w)
	if err := bw.WriteHeader(len(lodMesh.Mesh.Vertices), len(lodMesh.Mesh.Indices)); err != nil {
		return err
	}
	if err := writeVertices(bw, lodMesh.Mesh.Vertices); err != nil {
		return err
	}
	if err := writeIndices(bw, lodMesh.Mesh.Indices); err != nil {
		return err
	}
	return bw.Flush()
}

func generateLOD(g *unpeople.Generator, p unpeople.Params, w io.Writer) error {
	if err := validateLODLevel(*lodFlag); err != nil {
		return err
	}

	result, err := g.GenerateWithLOD(p)
	if err != nil {
		return err
	}

	level := unpeople.LODLevel(*lodFlag)
	lodMesh := result.LODSet.Meshes[level]
	info("Generated LOD%d: %d triangles (%.0f%% of LOD0)",
		level, lodMesh.TriangleCount, lodMesh.TriangleRatio*100)

	return writeLODMesh(lodMesh, w)
}

func info(format string, args ...any) {
	if !*quietFlag {
		fmt.Fprintf(os.Stderr, format+"\n", args...)
	}
}

func fatal(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "error: "+format+"\n", args...)
	os.Exit(1)
}
