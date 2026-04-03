// Package main provides tests for the unpeopled CLI tool.
package main

import (
	"bytes"
	"errors"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/opd-ai/unpeople"
)

// errorWriter always returns an error on Write.
type errorWriter struct{}

func (e errorWriter) Write(p []byte) (n int, err error) {
	return 0, errors.New("simulated write error")
}

func TestGenerateOBJ(t *testing.T) {
	g := unpeople.NewGenerator()
	p := unpeople.DefaultParams()
	p.Seed = 42

	var buf bytes.Buffer
	// Reset format flag for test
	*formatFlag = "obj"
	err := generateOBJ(g, p, &buf)
	if err != nil {
		t.Fatalf("generateOBJ failed: %v", err)
	}

	output := buf.String()
	if !strings.HasPrefix(output, "# Wavefront OBJ exported by unpeople") {
		t.Errorf("OBJ output should start with Wavefront header, got: %.60s", output)
	}
	if !strings.Contains(output, "v ") {
		t.Error("OBJ output should contain vertices")
	}
	if !strings.Contains(output, "f ") {
		t.Error("OBJ output should contain faces")
	}
}

func TestGenerateGLTF(t *testing.T) {
	g := unpeople.NewGenerator()
	p := unpeople.DefaultParams()
	p.Seed = 42

	var buf bytes.Buffer
	err := generateGLTF(g, p, &buf)
	if err != nil {
		t.Fatalf("generateGLTF failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, `"version": "2.0"`) && !strings.Contains(output, `"version":"2.0"`) {
		t.Error("glTF output should contain version 2.0")
	}
	if !strings.Contains(output, `"generator": "unpeople"`) && !strings.Contains(output, `"generator":"unpeople"`) {
		t.Error("glTF output should identify generator as unpeople")
	}
}

func TestGenerateGLB(t *testing.T) {
	g := unpeople.NewGenerator()
	p := unpeople.DefaultParams()
	p.Seed = 42

	var buf bytes.Buffer
	err := generateGLB(g, p, &buf)
	if err != nil {
		t.Fatalf("generateGLB failed: %v", err)
	}

	data := buf.Bytes()
	if len(data) < 12 {
		t.Fatal("GLB output too small")
	}
	// Check GLB magic
	if string(data[0:4]) != "glTF" {
		t.Errorf("GLB output should start with 'glTF' magic, got: %q", data[0:4])
	}
}

func TestGenerateBinary(t *testing.T) {
	g := unpeople.NewGenerator()
	p := unpeople.DefaultParams()
	p.Seed = 42

	var buf bytes.Buffer
	err := generateBinary(g, p, &buf)
	if err != nil {
		t.Fatalf("generateBinary failed: %v", err)
	}

	data := buf.Bytes()
	if len(data) < 8 {
		t.Fatal("Binary output too small")
	}
	// Check UNPM magic
	if string(data[0:4]) != "UNPM" {
		t.Errorf("Binary output should start with 'UNPM' magic, got: %q", data[0:4])
	}
}

func TestGenerateLOD(t *testing.T) {
	g := unpeople.NewGenerator()
	p := unpeople.DefaultParams()
	p.Seed = 42

	for level := 0; level < 3; level++ {
		*lodFlag = level
		var buf bytes.Buffer
		err := generateLOD(g, p, &buf)
		if err != nil {
			t.Fatalf("generateLOD (level %d) failed: %v", level, err)
		}

		data := buf.Bytes()
		if len(data) < 8 {
			t.Fatalf("LOD level %d output too small", level)
		}
		if string(data[0:4]) != "UNPM" {
			t.Errorf("LOD output should start with 'UNPM' magic, got: %q", data[0:4])
		}
	}
}

func TestGenerateLODInvalidLevel(t *testing.T) {
	g := unpeople.NewGenerator()
	p := unpeople.DefaultParams()
	p.Seed = 42

	*lodFlag = 5 // Invalid level
	var buf bytes.Buffer
	err := generateLOD(g, p, &buf)
	if err == nil {
		t.Error("Expected error for invalid LOD level, got nil")
	}
	if !strings.Contains(err.Error(), "invalid LOD level") {
		t.Errorf("Expected 'invalid LOD level' error, got: %v", err)
	}
}

func TestGenerateInvalidFormat(t *testing.T) {
	p := unpeople.DefaultParams()
	p.Seed = 42

	*formatFlag = "invalid"
	var buf bytes.Buffer
	err := generate(p, &buf)
	if err == nil {
		t.Error("Expected error for invalid format, got nil")
	}
	if !strings.Contains(err.Error(), "unknown format") {
		t.Errorf("Expected 'unknown format' error, got: %v", err)
	}
}

func TestHelpEnumConsistency(t *testing.T) {
	// Verify the help text matches the actual enum definitions
	tests := []struct {
		name     string
		helpText string
	}{
		{
			"Height",
			"0=Giant, 1=Tall, 2=Medium, 3=Short, 4=Tiny",
		},
		{
			"Build",
			"0=Muscular, 1=Athletic, 2=Average, 3=Lean, 4=Stocky, 5=Fragile",
		},
		{
			"Age",
			"0=Decrepit, 1=Elderly, 2=Old, 3=Adult, 4=Youth, 5=Teen, 6=Child, 7=Toddler",
		},
		{
			"Species",
			"0=Human, 1=Elf, 2=Dwarf, 3=Gnome, 4=Halfling",
		},
	}

	// Capture help output
	old := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	printUsage()

	w.Close()
	os.Stderr = old
	var buf bytes.Buffer
	io.Copy(&buf, r)
	helpOutput := buf.String()

	for _, tt := range tests {
		if !strings.Contains(helpOutput, tt.helpText) {
			t.Errorf("Help text missing or incorrect %s values.\nExpected to contain: %s\nGot: %s",
				tt.name, tt.helpText, helpOutput)
		}
	}
}

func TestLoadParamsWithSeedFlag(t *testing.T) {
	// Test that seed flag overrides stdin
	*seedFlag = 12345
	defer func() { *seedFlag = 0 }()

	p, err := loadParams()
	if err != nil {
		t.Fatalf("loadParams failed: %v", err)
	}
	if p.Seed != 12345 {
		t.Errorf("Expected seed 12345, got %d", p.Seed)
	}
}

func TestLoadParamsFromJSON(t *testing.T) {
	*seedFlag = 0 // Ensure we read from stdin

	// Save original stdin
	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }()

	// Create pipe for test input
	r, w, _ := os.Pipe()
	os.Stdin = r

	// Write test JSON
	go func() {
		w.Write([]byte(`{"seed": 99999, "species": 2}`))
		w.Close()
	}()

	p, err := loadParams()
	if err != nil {
		t.Fatalf("loadParams failed: %v", err)
	}
	if p.Seed != 99999 {
		t.Errorf("Expected seed 99999, got %d", p.Seed)
	}
	if p.Species != 2 {
		t.Errorf("Expected species 2, got %d", p.Species)
	}
}

func TestLoadParamsInvalidJSON(t *testing.T) {
	*seedFlag = 0

	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }()

	r, w, _ := os.Pipe()
	os.Stdin = r

	go func() {
		w.Write([]byte(`{invalid json`))
		w.Close()
	}()

	_, err := loadParams()
	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
	if !strings.Contains(err.Error(), "parsing JSON") {
		t.Errorf("Expected 'parsing JSON' error, got: %v", err)
	}
}

func TestAllFormats(t *testing.T) {
	formats := []string{"obj", "gltf", "glb", "binary"}
	p := unpeople.DefaultParams()
	p.Seed = 42

	for _, fmt := range formats {
		t.Run(fmt, func(t *testing.T) {
			*formatFlag = fmt
			var buf bytes.Buffer
			err := generate(p, &buf)
			if err != nil {
				t.Errorf("Format %s failed: %v", fmt, err)
			}
			if buf.Len() == 0 {
				t.Errorf("Format %s produced empty output", fmt)
			}
		})
	}
}

func TestDeterministicOutput(t *testing.T) {
	p := unpeople.DefaultParams()
	p.Seed = 42

	*formatFlag = "obj"

	// Generate twice with same seed
	var buf1, buf2 bytes.Buffer
	g := unpeople.NewGenerator()

	if err := generateOBJ(g, p, &buf1); err != nil {
		t.Fatalf("First generation failed: %v", err)
	}
	if err := generateOBJ(g, p, &buf2); err != nil {
		t.Fatalf("Second generation failed: %v", err)
	}

	if buf1.String() != buf2.String() {
		t.Error("Same seed should produce identical output")
	}
}

func TestLoadParamsEmptyStdin(t *testing.T) {
	*seedFlag = 0

	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }()

	r, w, _ := os.Pipe()
	os.Stdin = r

	// Write nothing, just close
	go func() {
		w.Close()
	}()

	p, err := loadParams()
	if err != nil {
		t.Fatalf("loadParams with empty stdin failed: %v", err)
	}
	// Should get default params
	defaults := unpeople.DefaultParams()
	if p.Species != defaults.Species {
		t.Errorf("Expected default species %d, got %d", defaults.Species, p.Species)
	}
}

func TestInfoFunction(t *testing.T) {
	// Test with quiet mode disabled
	*quietFlag = false
	defer func() { *quietFlag = true }()

	// Capture stderr
	old := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	info("test message %d", 42)

	w.Close()
	os.Stderr = old
	var buf bytes.Buffer
	io.Copy(&buf, r)

	if !strings.Contains(buf.String(), "test message 42") {
		t.Errorf("info() output missing, got: %s", buf.String())
	}
}

func TestInfoFunctionQuiet(t *testing.T) {
	// Test with quiet mode enabled
	*quietFlag = true

	old := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	info("should not appear")

	w.Close()
	os.Stderr = old
	var buf bytes.Buffer
	io.Copy(&buf, r)

	if buf.Len() > 0 {
		t.Errorf("info() should produce no output in quiet mode, got: %s", buf.String())
	}
}

func TestValidateLODLevel(t *testing.T) {
	tests := []struct {
		level   int
		wantErr bool
	}{
		{0, false},
		{1, false},
		{2, false},
		{-1, true},
		{3, true},
		{100, true},
	}

	for _, tt := range tests {
		err := validateLODLevel(tt.level)
		if (err != nil) != tt.wantErr {
			t.Errorf("validateLODLevel(%d) error = %v, wantErr %v", tt.level, err, tt.wantErr)
		}
	}
}

func TestWriteLODMesh(t *testing.T) {
	g := unpeople.NewGenerator()
	p := unpeople.DefaultParams()
	p.Seed = 42

	result, err := g.GenerateWithLOD(p)
	if err != nil {
		t.Fatalf("GenerateWithLOD failed: %v", err)
	}

	var buf bytes.Buffer
	err = writeLODMesh(result.LODSet.Meshes[0], &buf)
	if err != nil {
		t.Fatalf("writeLODMesh failed: %v", err)
	}

	// Verify UNPM magic
	if buf.Len() < 4 {
		t.Fatal("Output too small")
	}
	if string(buf.Bytes()[0:4]) != "UNPM" {
		t.Errorf("Expected UNPM magic, got: %q", buf.Bytes()[0:4])
	}
}

func TestPrintUsage(t *testing.T) {
	// Capture stderr
	old := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	printUsage()

	w.Close()
	os.Stderr = old
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Verify key sections are present
	checks := []string{
		"unpeopled",
		"Usage:",
		"Options:",
		"Output Formats:",
		"obj",
		"gltf",
		"glb",
		"binary",
		"lod",
	}

	for _, check := range checks {
		if !strings.Contains(output, check) {
			t.Errorf("printUsage() missing %q", check)
		}
	}
}

func TestGenerateLODNegativeLevel(t *testing.T) {
	g := unpeople.NewGenerator()
	p := unpeople.DefaultParams()
	p.Seed = 42

	*lodFlag = -1
	var buf bytes.Buffer
	err := generateLOD(g, p, &buf)
	if err == nil {
		t.Error("Expected error for negative LOD level, got nil")
	}
}

func TestGenerateDispatch(t *testing.T) {
	// Test generate() function dispatch for all formats
	p := unpeople.DefaultParams()
	p.Seed = 42

	tests := []struct {
		format string
		magic  string
	}{
		{"obj", "# Wavefront OBJ"},
		{"gltf", `"asset"`},
		{"glb", "glTF"},
		{"binary", "UNPM"},
	}

	for _, tt := range tests {
		t.Run(tt.format, func(t *testing.T) {
			*formatFlag = tt.format
			var buf bytes.Buffer
			err := generate(p, &buf)
			if err != nil {
				t.Fatalf("generate() with format %s failed: %v", tt.format, err)
			}
			if !strings.Contains(buf.String(), tt.magic) && !bytes.Contains(buf.Bytes(), []byte(tt.magic)) {
				t.Errorf("Output for format %s missing expected content %q", tt.format, tt.magic)
			}
		})
	}
}

func TestGenerateLODDispatch(t *testing.T) {
	p := unpeople.DefaultParams()
	p.Seed = 42

	*formatFlag = "lod"
	*lodFlag = 1

	var buf bytes.Buffer
	err := generate(p, &buf)
	if err != nil {
		t.Fatalf("generate() with format lod failed: %v", err)
	}

	if buf.Len() < 4 {
		t.Fatal("LOD output too small")
	}
	if string(buf.Bytes()[0:4]) != "UNPM" {
		t.Errorf("Expected UNPM magic, got: %q", buf.Bytes()[0:4])
	}
}

func TestGenerateOBJWriteError(t *testing.T) {
	g := unpeople.NewGenerator()
	p := unpeople.DefaultParams()
	p.Seed = 42

	err := generateOBJ(g, p, errorWriter{})
	if err == nil {
		t.Error("Expected error for write failure, got nil")
	}
}

func TestGenerateGLTFWriteError(t *testing.T) {
	g := unpeople.NewGenerator()
	p := unpeople.DefaultParams()
	p.Seed = 42

	err := generateGLTF(g, p, errorWriter{})
	if err == nil {
		t.Error("Expected error for write failure, got nil")
	}
}

func TestGenerateGLBWriteError(t *testing.T) {
	g := unpeople.NewGenerator()
	p := unpeople.DefaultParams()
	p.Seed = 42

	err := generateGLB(g, p, errorWriter{})
	if err == nil {
		t.Error("Expected error for write failure, got nil")
	}
}

func TestWriteLODMeshWriteError(t *testing.T) {
	g := unpeople.NewGenerator()
	p := unpeople.DefaultParams()
	p.Seed = 42

	result, err := g.GenerateWithLOD(p)
	if err != nil {
		t.Fatalf("GenerateWithLOD failed: %v", err)
	}

	err = writeLODMesh(result.LODSet.Meshes[0], errorWriter{})
	if err == nil {
		t.Error("Expected error for write failure, got nil")
	}
}

func TestInfoWithFormatArgs(t *testing.T) {
	*quietFlag = false
	defer func() { *quietFlag = true }()

	old := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	info("vertices: %d, triangles: %d", 100, 50)

	w.Close()
	os.Stderr = old
	var buf bytes.Buffer
	io.Copy(&buf, r)

	if !strings.Contains(buf.String(), "vertices: 100, triangles: 50") {
		t.Errorf("info() format args not working, got: %s", buf.String())
	}
}

func TestLoadParamsWithZeroSeedInJSON(t *testing.T) {
	*seedFlag = 0

	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }()

	r, w, _ := os.Pipe()
	os.Stdin = r

	// JSON with seed=0 should use default seed (also 0)
	go func() {
		w.Write([]byte(`{"seed": 0, "species": 5}`))
		w.Close()
	}()

	p, err := loadParams()
	if err != nil {
		t.Fatalf("loadParams failed: %v", err)
	}
	// seed=0 in JSON should fall back to default seed
	defaults := unpeople.DefaultParams()
	if p.Seed != defaults.Seed {
		t.Errorf("Expected default seed %d, got %d", defaults.Seed, p.Seed)
	}
	if p.Species != 5 {
		t.Errorf("Expected species 5, got %d", p.Species)
	}
}

func TestAllLODLevels(t *testing.T) {
	p := unpeople.DefaultParams()
	p.Seed = 42

	for level := 0; level < 3; level++ {
		t.Run(string(rune('0'+level)), func(t *testing.T) {
			*formatFlag = "lod"
			*lodFlag = level

			var buf bytes.Buffer
			err := generate(p, &buf)
			if err != nil {
				t.Fatalf("generate() with LOD level %d failed: %v", level, err)
			}
			if buf.Len() == 0 {
				t.Errorf("LOD level %d produced empty output", level)
			}
		})
	}
}

func init() {
	// Enable quiet mode by default for tests
	*quietFlag = true
}
