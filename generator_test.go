package unpeople_test

import (
	"testing"
	"time"

	"github.com/opd-ai/unpeople"
)

// ─── Determinism ─────────────────────────────────────────────────────────────

func TestGenerateDeterministic(t *testing.T) {
	g := unpeople.NewGenerator()
	p := unpeople.DefaultParams()
	p.Seed = 42

	m1, err := g.Generate(p)
	if err != nil {
		t.Fatalf("first Generate failed: %v", err)
	}
	m2, err := g.Generate(p)
	if err != nil {
		t.Fatalf("second Generate failed: %v", err)
	}

	if len(m1.Vertices) != len(m2.Vertices) {
		t.Fatalf("vertex count mismatch: %d vs %d", len(m1.Vertices), len(m2.Vertices))
	}
	if len(m1.Indices) != len(m2.Indices) {
		t.Fatalf("index count mismatch: %d vs %d", len(m1.Indices), len(m2.Indices))
	}
	for i := range m1.Vertices {
		if m1.Vertices[i] != m2.Vertices[i] {
			t.Errorf("vertex[%d] differs between two calls with same params", i)
		}
	}
	for i := range m1.Indices {
		if m1.Indices[i] != m2.Indices[i] {
			t.Errorf("index[%d] differs between two calls with same params", i)
		}
	}
}

// Different seeds must produce different meshes (at least the key differs).
func TestDifferentSeedsDifferentKey(t *testing.T) {
	g := unpeople.NewGenerator()
	p1 := unpeople.DefaultParams()
	p1.Seed = 1

	p2 := unpeople.DefaultParams()
	p2.Seed = 2

	m1, _ := g.Generate(p1)
	m2, _ := g.Generate(p2)

	if m1.Key == m2.Key {
		t.Errorf("expected different keys for different seeds, got %q", m1.Key)
	}
}

// ─── Mesh validity ───────────────────────────────────────────────────────────

func TestMeshIsValid(t *testing.T) {
	g := unpeople.NewGenerator()
	m, err := g.Generate(unpeople.DefaultParams())
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if len(m.Vertices) == 0 {
		t.Error("mesh has no vertices")
	}
	if len(m.Indices) == 0 {
		t.Error("mesh has no indices")
	}
	if len(m.Indices)%3 != 0 {
		t.Errorf("index count %d is not a multiple of 3", len(m.Indices))
	}

	maxIdx := uint32(len(m.Vertices))
	for i, idx := range m.Indices {
		if idx >= maxIdx {
			t.Errorf("index[%d]=%d is out of range (vertices=%d)", i, idx, maxIdx)
		}
	}

	if m.Key == "" {
		t.Error("mesh key must not be empty")
	}
}

// ─── Species ─────────────────────────────────────────────────────────────────

func TestAllSpecies(t *testing.T) {
	g := unpeople.NewGenerator()
	species := []unpeople.Species{
		unpeople.SpeciesHuman,
		unpeople.SpeciesElf,
		unpeople.SpeciesDwarf,
		unpeople.SpeciesGnome,
		unpeople.SpeciesHalfling,
		unpeople.SpeciesGoblin,
		unpeople.SpeciesKobold,
		unpeople.SpeciesOrc,
		unpeople.SpeciesTroll,
		unpeople.SpeciesOgre,
	}
	for _, s := range species {
		p := unpeople.DefaultParams()
		p.Species = s
		m, err := g.Generate(p)
		if err != nil {
			t.Errorf("species=%d: Generate error: %v", s, err)
			continue
		}
		if len(m.Vertices) == 0 || len(m.Indices) == 0 {
			t.Errorf("species=%d: empty mesh", s)
		}
		// Sanity check indices
		for i, idx := range m.Indices {
			if idx >= uint32(len(m.Vertices)) {
				t.Errorf("species=%d: index[%d]=%d out of range", s, i, idx)
			}
		}
	}
}

// ─── Parameter extremes ──────────────────────────────────────────────────────

func TestAllHeights(t *testing.T) {
	g := unpeople.NewGenerator()
	for h := unpeople.HeightGiant; h <= unpeople.HeightTiny; h++ {
		p := unpeople.DefaultParams()
		p.Height = h
		if _, err := g.Generate(p); err != nil {
			t.Errorf("height=%d: %v", h, err)
		}
	}
}

func TestAllBuilds(t *testing.T) {
	g := unpeople.NewGenerator()
	for b := unpeople.BuildMuscular; b <= unpeople.BuildFragile; b++ {
		p := unpeople.DefaultParams()
		p.Build = b
		if _, err := g.Generate(p); err != nil {
			t.Errorf("build=%d: %v", b, err)
		}
	}
}

func TestAllAges(t *testing.T) {
	g := unpeople.NewGenerator()
	for a := unpeople.AgeDecrepit; a <= unpeople.AgeToddler; a++ {
		p := unpeople.DefaultParams()
		p.Age = a
		if _, err := g.Generate(p); err != nil {
			t.Errorf("age=%d: %v", a, err)
		}
	}
}

// ─── Validation ──────────────────────────────────────────────────────────────

func TestValidateRejectsOutOfRange(t *testing.T) {
	g := unpeople.NewGenerator()
	tests := []struct {
		name   string
		mutate func(*unpeople.Params)
	}{
		{"bad Species", func(p *unpeople.Params) { p.Species = 999 }},
		{"bad Height", func(p *unpeople.Params) { p.Height = -1 }},
		{"bad Build", func(p *unpeople.Params) { p.Build = 100 }},
		{"bad Proportions", func(p *unpeople.Params) { p.Proportions = 50 }},
		{"bad Phenotype", func(p *unpeople.Params) { p.Phenotype = 99 }},
		{"bad Age", func(p *unpeople.Params) { p.Age = -5 }},
		{"bad Posture", func(p *unpeople.Params) { p.Posture = 10 }},
	}
	for _, tc := range tests {
		p := unpeople.DefaultParams()
		tc.mutate(&p)
		_, err := g.Generate(p)
		if err == nil {
			t.Errorf("%s: expected validation error but got nil", tc.name)
		}
	}
}

// ─── Performance ─────────────────────────────────────────────────────────────

// TestPerformance ensures a single Generate call completes in under 100 ms.
func TestPerformance(t *testing.T) {
	g := unpeople.NewGenerator()
	p := unpeople.DefaultParams()
	p.Seed = 7777

	start := time.Now()
	_, err := g.Generate(p)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}
	if elapsed > 100*time.Millisecond {
		t.Errorf("Generate took %v, want < 100 ms", elapsed)
	}
}

func BenchmarkGenerate(b *testing.B) {
	g := unpeople.NewGenerator()
	p := unpeople.DefaultParams()
	p.Seed = 42
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := g.Generate(p); err != nil {
			b.Fatal(err)
		}
	}
}
