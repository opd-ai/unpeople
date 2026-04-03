//go:build !kaiju

// Package kaiju provides tests for the stub implementation.
package kaiju

import (
	"testing"

	"github.com/opd-ai/unpeople"
)

func TestKaijuGeneratorStub(t *testing.T) {
	gen := NewKaijuGenerator()
	if gen == nil {
		t.Fatal("NewKaijuGenerator returned nil")
	}

	params := unpeople.DefaultParams()
	params.Seed = 42

	mesh, err := gen.Generate(params)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}
	if mesh == nil {
		t.Fatal("Generate returned nil mesh")
	}
	if len(mesh.Vertices) == 0 {
		t.Error("Expected non-empty vertices")
	}
	if len(mesh.Indices) == 0 {
		t.Error("Expected non-empty indices")
	}
}

func TestDeterministicGeneration(t *testing.T) {
	gen := NewKaijuGenerator()
	params := unpeople.DefaultParams()
	params.Seed = 12345

	mesh1, _ := gen.Generate(params)
	mesh2, _ := gen.Generate(params)

	if len(mesh1.Vertices) != len(mesh2.Vertices) {
		t.Error("Same seed should produce same vertex count")
	}
	if len(mesh1.Indices) != len(mesh2.Indices) {
		t.Error("Same seed should produce same index count")
	}
	if mesh1.Key != mesh2.Key {
		t.Error("Same seed should produce same key")
	}
}
