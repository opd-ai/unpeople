// Package main provides tests for the unpeopled CLI tool.
package main

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/opd-ai/unpeople"
)

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

func init() {
	// Enable quiet mode by default for tests
	*quietFlag = true
}
