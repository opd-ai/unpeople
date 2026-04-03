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

// TestKeyUniqueness verifies that changing any geometry-affecting parameter
// that was previously absent from the key produces a different key.
func TestKeyUniqueness(t *testing.T) {
	g := unpeople.NewGenerator()
	base := unpeople.DefaultParams()
	base.Seed = 0

	baseM, err := g.Generate(base)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	variants := []struct {
		name   string
		mutate func(*unpeople.Params)
	}{
		{"FaceShape", func(p *unpeople.Params) { p.FaceShape = unpeople.FaceShapeSquare }},
		{"Jaw", func(p *unpeople.Params) { p.Jaw = unpeople.JawProminent }},
		{"Brow", func(p *unpeople.Params) { p.Brow = unpeople.BrowHeavy }},
		{"Ears", func(p *unpeople.Params) { p.Ears = unpeople.EarsLarge }},
		{"ShoulderWidth", func(p *unpeople.Params) { p.ShoulderWidth = unpeople.ShoulderWidthBroad }},
		{"HipWidth", func(p *unpeople.Params) { p.HipWidth = unpeople.HipWidthWide }},
		{"LimbLength", func(p *unpeople.Params) { p.LimbLength = unpeople.LimbLengthLong }},
		{"NeckLength", func(p *unpeople.Params) { p.NeckLength = unpeople.NeckLengthLong }},
		{"HandSize", func(p *unpeople.Params) { p.HandSize = unpeople.HandSizeLarge }},
		{"FingerLength", func(p *unpeople.Params) { p.FingerLength = unpeople.FingerLengthLong }},
		{"FootSize", func(p *unpeople.Params) { p.FootSize = unpeople.FootSizeLarge }},
	}

	for _, v := range variants {
		p := unpeople.DefaultParams()
		v.mutate(&p)
		m, err := g.Generate(p)
		if err != nil {
			t.Errorf("%s: Generate failed: %v", v.name, err)
			continue
		}
		if m.Key == baseM.Key {
			t.Errorf("%s: key unchanged after param mutation: %q", v.name, m.Key)
		}
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

func TestAllProportions(t *testing.T) {
	g := unpeople.NewGenerator()
	for pr := unpeople.ProportionsHeroic; pr <= unpeople.ProportionsCaricature; pr++ {
		p := unpeople.DefaultParams()
		p.Proportions = pr
		if _, err := g.Generate(p); err != nil {
			t.Errorf("proportions=%d: %v", pr, err)
		}
	}
}

func TestAllPhenotypes(t *testing.T) {
	g := unpeople.NewGenerator()
	for ph := unpeople.PhenotypeMasculine; ph <= unpeople.PhenotypeFeminine; ph++ {
		p := unpeople.DefaultParams()
		p.Phenotype = ph
		if _, err := g.Generate(p); err != nil {
			t.Errorf("phenotype=%d: %v", ph, err)
		}
	}
}

func TestAllPostures(t *testing.T) {
	g := unpeople.NewGenerator()
	for po := unpeople.PostureUpright; po <= unpeople.PostureRigid; po++ {
		p := unpeople.DefaultParams()
		p.Posture = po
		if _, err := g.Generate(p); err != nil {
			t.Errorf("posture=%d: %v", po, err)
		}
	}
}

func TestAllFaceShapes(t *testing.T) {
	g := unpeople.NewGenerator()
	for fs := unpeople.FaceShapeOval; fs <= unpeople.FaceShapeOblong; fs++ {
		p := unpeople.DefaultParams()
		p.FaceShape = fs
		if _, err := g.Generate(p); err != nil {
			t.Errorf("faceshape=%d: %v", fs, err)
		}
	}
}

func TestAllJaws(t *testing.T) {
	g := unpeople.NewGenerator()
	for j := unpeople.JawProminent; j <= unpeople.JawRounded; j++ {
		p := unpeople.DefaultParams()
		p.Jaw = j
		if _, err := g.Generate(p); err != nil {
			t.Errorf("jaw=%d: %v", j, err)
		}
	}
}

func TestAllBrows(t *testing.T) {
	g := unpeople.NewGenerator()
	for br := unpeople.BrowHeavy; br <= unpeople.BrowArched; br++ {
		p := unpeople.DefaultParams()
		p.Brow = br
		if _, err := g.Generate(p); err != nil {
			t.Errorf("brow=%d: %v", br, err)
		}
	}
}

func TestAllEars(t *testing.T) {
	g := unpeople.NewGenerator()
	for e := unpeople.EarsSmall; e <= unpeople.EarsRounded; e++ {
		p := unpeople.DefaultParams()
		p.Ears = e
		if _, err := g.Generate(p); err != nil {
			t.Errorf("ears=%d: %v", e, err)
		}
	}
}

func TestAllShoulderWidths(t *testing.T) {
	g := unpeople.NewGenerator()
	for sw := unpeople.ShoulderWidthBroad; sw <= unpeople.ShoulderWidthNarrow; sw++ {
		p := unpeople.DefaultParams()
		p.ShoulderWidth = sw
		if _, err := g.Generate(p); err != nil {
			t.Errorf("shoulderwidth=%d: %v", sw, err)
		}
	}
}

func TestAllHipWidths(t *testing.T) {
	g := unpeople.NewGenerator()
	for hw := unpeople.HipWidthWide; hw <= unpeople.HipWidthNarrow; hw++ {
		p := unpeople.DefaultParams()
		p.HipWidth = hw
		if _, err := g.Generate(p); err != nil {
			t.Errorf("hipwidth=%d: %v", hw, err)
		}
	}
}

func TestAllLimbLengths(t *testing.T) {
	g := unpeople.NewGenerator()
	for ll := unpeople.LimbLengthLong; ll <= unpeople.LimbLengthShort; ll++ {
		p := unpeople.DefaultParams()
		p.LimbLength = ll
		if _, err := g.Generate(p); err != nil {
			t.Errorf("limblength=%d: %v", ll, err)
		}
	}
}

func TestAllNeckLengths(t *testing.T) {
	g := unpeople.NewGenerator()
	for nl := unpeople.NeckLengthLong; nl <= unpeople.NeckLengthThick; nl++ {
		p := unpeople.DefaultParams()
		p.NeckLength = nl
		if _, err := g.Generate(p); err != nil {
			t.Errorf("necklength=%d: %v", nl, err)
		}
	}
}

func TestAllHandSizes(t *testing.T) {
	g := unpeople.NewGenerator()
	for hs := unpeople.HandSizeLarge; hs <= unpeople.HandSizeSmall; hs++ {
		p := unpeople.DefaultParams()
		p.HandSize = hs
		if _, err := g.Generate(p); err != nil {
			t.Errorf("handsize=%d: %v", hs, err)
		}
	}
}

func TestAllFingerLengths(t *testing.T) {
	g := unpeople.NewGenerator()
	for fl := unpeople.FingerLengthLong; fl <= unpeople.FingerLengthShort; fl++ {
		p := unpeople.DefaultParams()
		p.FingerLength = fl
		if _, err := g.Generate(p); err != nil {
			t.Errorf("fingerlength=%d: %v", fl, err)
		}
	}
}

func TestAllFootSizes(t *testing.T) {
	g := unpeople.NewGenerator()
	for fs := unpeople.FootSizeLarge; fs <= unpeople.FootSizeSmall; fs++ {
		p := unpeople.DefaultParams()
		p.FootSize = fs
		if _, err := g.Generate(p); err != nil {
			t.Errorf("footsize=%d: %v", fs, err)
		}
	}
}

func TestAllSkinTones(t *testing.T) {
	g := unpeople.NewGenerator()
	for st := unpeople.SkinTonePale; st <= unpeople.SkinToneDark; st++ {
		p := unpeople.DefaultParams()
		p.SkinTone = st
		m, err := g.Generate(p)
		if err != nil {
			t.Errorf("skintone=%d: %v", st, err)
			continue
		}
		// Verify vertices have non-gray color
		if len(m.Vertices) > 0 {
			c := m.Vertices[0].Color
			if c[0] == 0.5 && c[1] == 0.5 && c[2] == 0.5 {
				t.Errorf("skintone=%d: vertex color should not be default gray", st)
			}
		}
	}
}

func TestAllSkinUndertones(t *testing.T) {
	g := unpeople.NewGenerator()
	for ut := unpeople.SkinUndertoneNeutral; ut <= unpeople.SkinUndertoneCool; ut++ {
		p := unpeople.DefaultParams()
		p.SkinUndertone = ut
		if _, err := g.Generate(p); err != nil {
			t.Errorf("skinundertone=%d: %v", ut, err)
		}
	}
}

func TestSkinToneAffectsKey(t *testing.T) {
	g := unpeople.NewGenerator()
	p1 := unpeople.DefaultParams()
	p1.SkinTone = unpeople.SkinTonePale

	p2 := unpeople.DefaultParams()
	p2.SkinTone = unpeople.SkinToneDark

	m1, _ := g.Generate(p1)
	m2, _ := g.Generate(p2)

	if m1.Key == m2.Key {
		t.Error("different skin tones should produce different mesh keys")
	}
}

func TestSkinUndertoneAffectsKey(t *testing.T) {
	g := unpeople.NewGenerator()
	p1 := unpeople.DefaultParams()
	p1.SkinUndertone = unpeople.SkinUndertoneWarm

	p2 := unpeople.DefaultParams()
	p2.SkinUndertone = unpeople.SkinUndertoneCool

	m1, _ := g.Generate(p1)
	m2, _ := g.Generate(p2)

	if m1.Key == m2.Key {
		t.Error("different skin undertones should produce different mesh keys")
	}
}

func TestSkinToneProducesValidColor(t *testing.T) {
	g := unpeople.NewGenerator()

	// Test all combinations of skin tone and undertone
	for st := unpeople.SkinTonePale; st <= unpeople.SkinToneDark; st++ {
		for ut := unpeople.SkinUndertoneNeutral; ut <= unpeople.SkinUndertoneCool; ut++ {
			p := unpeople.DefaultParams()
			p.SkinTone = st
			p.SkinUndertone = ut

			m, err := g.Generate(p)
			if err != nil {
				t.Errorf("tone=%d undertone=%d: %v", st, ut, err)
				continue
			}

			// Check that all vertices have valid color values (0-1 range)
			for i, v := range m.Vertices {
				if v.Color[0] < 0 || v.Color[0] > 1 ||
					v.Color[1] < 0 || v.Color[1] > 1 ||
					v.Color[2] < 0 || v.Color[2] > 1 ||
					v.Color[3] < 0 || v.Color[3] > 1 {
					t.Errorf("tone=%d undertone=%d: vertex[%d] has invalid color: %v",
						st, ut, i, v.Color)
					break
				}
			}
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
		{"bad SkinTone", func(p *unpeople.Params) { p.SkinTone = 99 }},
		{"bad SkinUndertone", func(p *unpeople.Params) { p.SkinUndertone = 50 }},
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
// Skipped when -short is set to avoid flakiness on slow/loaded CI runners.
func TestPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping timing test in short mode")
	}
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
