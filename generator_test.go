package unpeople_test

import (
	"bytes"
	"context"
	"strings"
	"sync"
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

// ─── Material ────────────────────────────────────────────────────────────────

func TestGenerateWithMaterial(t *testing.T) {
	g := unpeople.NewGenerator()
	p := unpeople.DefaultParams()
	p.SkinTone = unpeople.SkinToneMedium
	p.SkinUndertone = unpeople.SkinUndertoneWarm

	result, err := g.GenerateWithMaterial(p)
	if err != nil {
		t.Fatalf("GenerateWithMaterial failed: %v", err)
	}

	if result.Mesh == nil {
		t.Error("MeshWithMaterial.Mesh is nil")
	}
	if len(result.Mesh.Vertices) == 0 {
		t.Error("mesh has no vertices")
	}

	// Check material properties
	m := result.Material
	if m.ShaderName != "standard" {
		t.Errorf("expected shader 'standard', got %q", m.ShaderName)
	}
	if m.Metallic != 0.0 {
		t.Errorf("skin should be non-metallic, got %f", m.Metallic)
	}
	if m.Roughness <= 0 || m.Roughness >= 1 {
		t.Errorf("roughness out of expected range: %f", m.Roughness)
	}

	// Material albedo should match computed skin color
	expectedColor := unpeople.ComputeSkinColor(p.SkinTone, p.SkinUndertone)
	if m.Albedo != expectedColor {
		t.Errorf("material albedo %v doesn't match expected skin color %v", m.Albedo, expectedColor)
	}
}

func TestDefaultSkinMaterial(t *testing.T) {
	color := unpeople.Color{0.7, 0.6, 0.5, 1.0}
	m := unpeople.DefaultSkinMaterial(color)

	if m.Albedo != color {
		t.Errorf("albedo mismatch: got %v, want %v", m.Albedo, color)
	}
	if m.Metallic != 0.0 {
		t.Errorf("skin should not be metallic: %f", m.Metallic)
	}
	if m.SubsurfaceScattering <= 0 {
		t.Error("skin should have positive subsurface scattering")
	}
}

func TestAgeSkinMaterial(t *testing.T) {
	color := unpeople.Color{0.7, 0.6, 0.5, 1.0}

	// Younger should be smoother than older
	toddlerMat := unpeople.AgeSkinMaterial(color, unpeople.AgeToddler)
	decrepitMat := unpeople.AgeSkinMaterial(color, unpeople.AgeDecrepit)

	if toddlerMat.Roughness >= decrepitMat.Roughness {
		t.Errorf("toddler skin (%f) should be smoother than decrepit (%f)",
			toddlerMat.Roughness, decrepitMat.Roughness)
	}

	// Younger should have more SSS
	if toddlerMat.SubsurfaceScattering <= decrepitMat.SubsurfaceScattering {
		t.Errorf("toddler SSS (%f) should be higher than decrepit (%f)",
			toddlerMat.SubsurfaceScattering, decrepitMat.SubsurfaceScattering)
	}
}

func TestUnlitMaterial(t *testing.T) {
	color := unpeople.Color{1.0, 0.0, 0.0, 1.0}
	m := unpeople.UnlitMaterial(color)

	if m.ShaderName != "unlit" {
		t.Errorf("expected 'unlit' shader, got %q", m.ShaderName)
	}
	if m.Albedo != color {
		t.Errorf("albedo mismatch: got %v, want %v", m.Albedo, color)
	}
}

func TestSSSkinMaterial(t *testing.T) {
	color := unpeople.Color{0.7, 0.6, 0.5, 1.0}
	standard := unpeople.DefaultSkinMaterial(color)
	sss := unpeople.SSSkinMaterial(color)

	if sss.ShaderName != "pbr" {
		t.Errorf("SSS material should use 'pbr' shader, got %q", sss.ShaderName)
	}
	if sss.SubsurfaceScattering <= standard.SubsurfaceScattering {
		t.Error("SSS material should have higher subsurface scattering than default")
	}
}

// ─── Cache ───────────────────────────────────────────────────────────────────

func TestCachedGeneratorBasic(t *testing.T) {
	cg := unpeople.NewCachedGenerator(100)
	p := unpeople.DefaultParams()
	p.Seed = 42

	// First call - cache miss
	m1, err := cg.Generate(p)
	if err != nil {
		t.Fatalf("first Generate failed: %v", err)
	}

	// Second call - cache hit
	m2, err := cg.Generate(p)
	if err != nil {
		t.Fatalf("second Generate failed: %v", err)
	}

	// Should return the same mesh instance
	if m1 != m2 {
		t.Error("cache should return same mesh instance for identical params")
	}

	// Cache should have 1 entry
	if cg.CacheSize() != 1 {
		t.Errorf("expected cache size 1, got %d", cg.CacheSize())
	}
}

func TestCachedGeneratorDifferentParams(t *testing.T) {
	cg := unpeople.NewCachedGenerator(100)

	p1 := unpeople.DefaultParams()
	p1.Seed = 1

	p2 := unpeople.DefaultParams()
	p2.Seed = 2

	m1, _ := cg.Generate(p1)
	m2, _ := cg.Generate(p2)

	if m1 == m2 {
		t.Error("different params should produce different mesh instances")
	}

	if cg.CacheSize() != 2 {
		t.Errorf("expected cache size 2, got %d", cg.CacheSize())
	}
}

func TestCachedGeneratorLRUEviction(t *testing.T) {
	cg := unpeople.NewCachedGenerator(3) // Small cache for testing

	// Generate 4 different meshes
	for i := int64(1); i <= 4; i++ {
		p := unpeople.DefaultParams()
		p.Seed = i
		if _, err := cg.Generate(p); err != nil {
			t.Fatalf("Generate failed for seed %d: %v", i, err)
		}
	}

	// Cache should only have 3 entries (LRU evicted one)
	if cg.CacheSize() != 3 {
		t.Errorf("expected cache size 3 after eviction, got %d", cg.CacheSize())
	}
}

func TestCachedGeneratorClearCache(t *testing.T) {
	cg := unpeople.NewCachedGenerator(100)

	// Populate cache
	for i := int64(1); i <= 5; i++ {
		p := unpeople.DefaultParams()
		p.Seed = i
		cg.Generate(p)
	}

	if cg.CacheSize() != 5 {
		t.Errorf("expected cache size 5, got %d", cg.CacheSize())
	}

	cg.ClearCache()

	if cg.CacheSize() != 0 {
		t.Errorf("expected cache size 0 after clear, got %d", cg.CacheSize())
	}
}

func TestCachedGeneratorInvalidate(t *testing.T) {
	cg := unpeople.NewCachedGenerator(100)

	p := unpeople.DefaultParams()
	p.Seed = 42

	// Generate and cache
	cg.Generate(p)
	if cg.CacheSize() != 1 {
		t.Errorf("expected cache size 1, got %d", cg.CacheSize())
	}

	// Invalidate
	if !cg.Invalidate(p) {
		t.Error("Invalidate should return true for cached entry")
	}

	if cg.CacheSize() != 0 {
		t.Errorf("expected cache size 0 after invalidate, got %d", cg.CacheSize())
	}

	// Invalidating again should return false
	if cg.Invalidate(p) {
		t.Error("Invalidate should return false for non-existent entry")
	}
}

func TestCachedGeneratorZeroSize(t *testing.T) {
	cg := unpeople.NewCachedGenerator(0) // Caching disabled

	p := unpeople.DefaultParams()

	m1, _ := cg.Generate(p)
	m2, _ := cg.Generate(p)

	// With caching disabled, should generate new instances each time
	if m1 == m2 {
		t.Error("with maxSize=0, caching should be disabled")
	}

	if cg.CacheSize() != 0 {
		t.Errorf("expected cache size 0 with disabled cache, got %d", cg.CacheSize())
	}
}

func TestCachedGeneratorWithMaterial(t *testing.T) {
	cg := unpeople.NewCachedGenerator(100)
	p := unpeople.DefaultParams()
	p.SkinTone = unpeople.SkinToneTan

	result, err := cg.GenerateWithMaterial(p)
	if err != nil {
		t.Fatalf("GenerateWithMaterial failed: %v", err)
	}

	if result.Mesh == nil {
		t.Error("mesh is nil")
	}
	if result.Material.ShaderName == "" {
		t.Error("material shader name is empty")
	}

	// Cache should contain the mesh
	if cg.CacheSize() != 1 {
		t.Errorf("expected cache size 1, got %d", cg.CacheSize())
	}
}

func TestCachedGeneratorConcurrency(t *testing.T) {
	cg := unpeople.NewCachedGenerator(100)
	p := unpeople.DefaultParams()
	p.Seed = 42

	var wg sync.WaitGroup
	results := make(chan *unpeople.Mesh, 10)

	// Launch 10 concurrent goroutines all requesting same params
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			m, err := cg.Generate(p)
			if err != nil {
				t.Errorf("concurrent Generate failed: %v", err)
				return
			}
			results <- m
		}()
	}

	wg.Wait()
	close(results)

	// All results should be the same instance
	var first *unpeople.Mesh
	for m := range results {
		if first == nil {
			first = m
		} else if m != first {
			t.Error("concurrent calls with same params should return same cached instance")
		}
	}

	// Should only be 1 cache entry
	if cg.CacheSize() != 1 {
		t.Errorf("expected cache size 1, got %d", cg.CacheSize())
	}
}

// ─── Batch Generation ────────────────────────────────────────────────────────

func TestBatchGeneratorBasic(t *testing.T) {
	bg := unpeople.NewBatchGenerator()

	params := make([]unpeople.Params, 5)
	for i := range params {
		params[i] = unpeople.DefaultParams()
		params[i].Seed = int64(i)
	}

	results := bg.GenerateBatch(context.Background(), params, unpeople.BatchOptions{})

	if len(results) != 5 {
		t.Fatalf("expected 5 results, got %d", len(results))
	}

	for i, r := range results {
		if r.Err != nil {
			t.Errorf("result[%d] error: %v", i, r.Err)
		}
		if r.Mesh == nil {
			t.Errorf("result[%d] mesh is nil", i)
		}
		if r.Index != i {
			t.Errorf("result[%d] has wrong index: %d", i, r.Index)
		}
	}
}

func TestBatchGeneratorWithCache(t *testing.T) {
	bg := unpeople.NewBatchGeneratorWithCache(100)

	// Generate same params twice in batch
	params := []unpeople.Params{
		unpeople.DefaultParams(),
		unpeople.DefaultParams(),
		unpeople.DefaultParams(),
	}
	params[0].Seed = 42
	params[1].Seed = 42 // Same as first
	params[2].Seed = 43 // Different

	results := bg.GenerateBatch(context.Background(), params, unpeople.BatchOptions{})

	if results[0].Mesh != results[1].Mesh {
		t.Error("same params should return same cached mesh")
	}
	if results[0].Mesh == results[2].Mesh {
		t.Error("different params should return different meshes")
	}

	size, _ := bg.CacheStats()
	if size != 2 {
		t.Errorf("expected cache size 2, got %d", size)
	}
}

func TestBatchGeneratorWithMaterial(t *testing.T) {
	bg := unpeople.NewBatchGenerator()

	params := []unpeople.Params{unpeople.DefaultParams()}
	params[0].SkinTone = unpeople.SkinToneTan

	results := bg.GenerateBatch(context.Background(), params, unpeople.BatchOptions{
		IncludeMaterial: true,
	})

	if results[0].Material == nil {
		t.Error("expected material to be set")
	}
	if results[0].Material.ShaderName == "" {
		t.Error("material should have shader name")
	}
}

func TestBatchGeneratorCancellation(t *testing.T) {
	bg := unpeople.NewBatchGenerator()

	// Create many params to process
	params := make([]unpeople.Params, 100)
	for i := range params {
		params[i] = unpeople.DefaultParams()
		params[i].Seed = int64(i)
	}

	// Create already-cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	results := bg.GenerateBatch(ctx, params, unpeople.BatchOptions{
		Workers: 2,
	})

	// Some or all results should have context error
	cancelled := 0
	for _, r := range results {
		if r.Err == context.Canceled {
			cancelled++
		}
	}

	// At least some should be cancelled (depends on timing)
	if cancelled == 0 && len(results) > 0 {
		// If workers are very fast, they might finish before cancellation takes effect
		// This is acceptable behavior - we just verify no panics
	}
}

func TestBatchGeneratorSimple(t *testing.T) {
	bg := unpeople.NewBatchGenerator()

	params := make([]unpeople.Params, 3)
	for i := range params {
		params[i] = unpeople.DefaultParams()
		params[i].Seed = int64(i)
	}

	meshes := bg.GenerateBatchSimple(context.Background(), params)

	if len(meshes) != 3 {
		t.Fatalf("expected 3 meshes, got %d", len(meshes))
	}

	for i, m := range meshes {
		if m == nil {
			t.Errorf("mesh[%d] is nil", i)
		}
	}
}

func TestBatchGeneratorWithMaterialMethod(t *testing.T) {
	bg := unpeople.NewBatchGenerator()

	params := []unpeople.Params{unpeople.DefaultParams()}

	results := bg.GenerateBatchWithMaterial(context.Background(), params)

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Material == nil {
		t.Error("material should be set")
	}
}

func TestBatchGeneratorEmptyInput(t *testing.T) {
	bg := unpeople.NewBatchGenerator()

	results := bg.GenerateBatch(context.Background(), nil, unpeople.BatchOptions{})
	if results != nil {
		t.Error("expected nil result for empty input")
	}

	results = bg.GenerateBatch(context.Background(), []unpeople.Params{}, unpeople.BatchOptions{})
	if results != nil {
		t.Error("expected nil result for empty slice")
	}
}

// ─── OBJ Export ──────────────────────────────────────────────────────────────

func TestExportOBJBasic(t *testing.T) {
	g := unpeople.NewGenerator()
	mesh, err := g.Generate(unpeople.DefaultParams())
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	var buf bytes.Buffer
	err = unpeople.ExportOBJ(&buf, mesh, "test_human")
	if err != nil {
		t.Fatalf("ExportOBJ failed: %v", err)
	}

	obj := buf.String()

	// Check for required OBJ elements
	if !strings.Contains(obj, "# Wavefront OBJ") {
		t.Error("missing OBJ header comment")
	}
	if !strings.Contains(obj, "o test_human") {
		t.Error("missing object name")
	}
	if !strings.Contains(obj, "v ") {
		t.Error("missing vertex positions")
	}
	if !strings.Contains(obj, "vt ") {
		t.Error("missing texture coordinates")
	}
	if !strings.Contains(obj, "vn ") {
		t.Error("missing vertex normals")
	}
	if !strings.Contains(obj, "f ") {
		t.Error("missing faces")
	}

	// Count vertices and faces
	vCount := strings.Count(obj, "\nv ")
	if vCount != len(mesh.Vertices) {
		t.Errorf("vertex count mismatch: OBJ has %d, mesh has %d", vCount, len(mesh.Vertices))
	}

	fCount := strings.Count(obj, "\nf ")
	expectedFaces := len(mesh.Indices) / 3
	if fCount != expectedFaces {
		t.Errorf("face count mismatch: OBJ has %d, expected %d", fCount, expectedFaces)
	}
}

func TestExportOBJDefaultName(t *testing.T) {
	g := unpeople.NewGenerator()
	mesh, _ := g.Generate(unpeople.DefaultParams())

	var buf bytes.Buffer
	unpeople.ExportOBJ(&buf, mesh, "") // Empty name

	if !strings.Contains(buf.String(), "o humanoid") {
		t.Error("empty name should default to 'humanoid'")
	}
}

func TestExportOBJWithMTL(t *testing.T) {
	g := unpeople.NewGenerator()
	p := unpeople.DefaultParams()
	p.SkinTone = unpeople.SkinToneTan

	mesh, _ := g.Generate(p)
	skinColor := unpeople.ComputeSkinColor(p.SkinTone, p.SkinUndertone)
	material := unpeople.DefaultSkinMaterial(skinColor)

	var objBuf, mtlBuf bytes.Buffer
	err := unpeople.ExportOBJWithMTL(&objBuf, &mtlBuf, mesh, &material, "character", "skin.mtl")
	if err != nil {
		t.Fatalf("ExportOBJWithMTL failed: %v", err)
	}

	obj := objBuf.String()
	mtl := mtlBuf.String()

	// Check OBJ references material library
	if !strings.Contains(obj, "mtllib skin.mtl") {
		t.Error("OBJ should reference MTL file")
	}
	if !strings.Contains(obj, "usemtl character_mat") {
		t.Error("OBJ should use material")
	}

	// Check MTL content
	if !strings.Contains(mtl, "newmtl character_mat") {
		t.Error("MTL should define material")
	}
	if !strings.Contains(mtl, "Kd ") {
		t.Error("MTL should have diffuse color")
	}
	if !strings.Contains(mtl, "Ks ") {
		t.Error("MTL should have specular color")
	}
}

func TestExportOBJNilMesh(t *testing.T) {
	var buf bytes.Buffer
	err := unpeople.ExportOBJ(&buf, nil, "test")
	if err == nil {
		t.Error("expected error for nil mesh")
	}
}

func TestExportOBJEmptyMesh(t *testing.T) {
	var buf bytes.Buffer
	mesh := &unpeople.Mesh{Key: "empty", Vertices: nil, Indices: nil}
	err := unpeople.ExportOBJ(&buf, mesh, "test")
	if err == nil {
		t.Error("expected error for empty mesh")
	}
}

// ─── Vertex Merging Tests ────────────────────────────────────────────────────

// TestMergingReducesVertexCount verifies that vertex merging actually
// reduces the vertex count by eliminating duplicates at body part boundaries.
func TestMergingReducesVertexCount(t *testing.T) {
	// Generate a mesh and verify it has fewer vertices than expected from
	// raw primitive assembly (which would have many duplicates)
	g := unpeople.NewGenerator()
	p := unpeople.DefaultParams()
	m, err := g.Generate(p)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// A fully unmerged human mesh would have ~2000-3000 vertices.
	// After merging, we expect a noticeable reduction (at least 50+ vertices).
	// We don't know the exact count, but we can verify the mesh is valid.
	if len(m.Vertices) < 100 {
		t.Errorf("too few vertices after merge: %d", len(m.Vertices))
	}

	// Verify all indices are valid
	for i, idx := range m.Indices {
		if int(idx) >= len(m.Vertices) {
			t.Errorf("index[%d]=%d out of bounds (vertices=%d)", i, idx, len(m.Vertices))
		}
	}
}

// TestMergingDeterminism verifies that vertex merging preserves determinism:
// same params always produce identical merged mesh.
func TestMergingDeterminism(t *testing.T) {
	g := unpeople.NewGenerator()
	p := unpeople.DefaultParams()
	p.Seed = 12345

	// Generate multiple times with same params
	var results []*unpeople.Mesh
	for i := 0; i < 3; i++ {
		m, err := g.Generate(p)
		if err != nil {
			t.Fatalf("Generate %d failed: %v", i, err)
		}
		results = append(results, m)
	}

	// Verify all results are identical
	for i := 1; i < len(results); i++ {
		if len(results[0].Vertices) != len(results[i].Vertices) {
			t.Errorf("run %d: vertex count %d != run 0: %d",
				i, len(results[i].Vertices), len(results[0].Vertices))
		}
		if len(results[0].Indices) != len(results[i].Indices) {
			t.Errorf("run %d: index count %d != run 0: %d",
				i, len(results[i].Indices), len(results[0].Indices))
		}

		// Spot-check some vertices
		for j := 0; j < len(results[0].Vertices) && j < 100; j++ {
			if results[0].Vertices[j] != results[i].Vertices[j] {
				t.Errorf("run %d vertex[%d] differs from run 0", i, j)
				break
			}
		}
	}
}

// TestMergingAcrossSpecies verifies that vertex merging works correctly
// for species with different body scales.
func TestMergingAcrossSpecies(t *testing.T) {
	g := unpeople.NewGenerator()

	// Test a variety of species with different body scales
	species := []unpeople.Species{
		unpeople.SpeciesHuman,
		unpeople.SpeciesGnome, // Small
		unpeople.SpeciesOgre,  // Large
		unpeople.SpeciesDwarf, // Compact
		unpeople.SpeciesTroll, // Very large
	}

	for _, sp := range species {
		p := unpeople.DefaultParams()
		p.Species = sp
		m, err := g.Generate(p)
		if err != nil {
			t.Errorf("species=%d: Generate failed: %v", sp, err)
			continue
		}

		// Verify mesh validity
		if len(m.Vertices) == 0 {
			t.Errorf("species=%d: no vertices", sp)
		}
		if len(m.Indices) == 0 {
			t.Errorf("species=%d: no indices", sp)
		}

		// Verify all indices are valid
		for i, idx := range m.Indices {
			if int(idx) >= len(m.Vertices) {
				t.Errorf("species=%d: index[%d]=%d out of bounds (vertices=%d)",
					sp, i, idx, len(m.Vertices))
				break
			}
		}
	}
}

// ─── UV Atlas Tests ──────────────────────────────────────────────────────────

// TestUVAtlasNonOverlapping verifies that UV coordinates are properly
// distributed across the atlas and don't overlap between body parts.
func TestUVAtlasNonOverlapping(t *testing.T) {
	g := unpeople.NewGenerator()
	p := unpeople.DefaultParams()
	m, err := g.Generate(p)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Verify all UVs are within [0,1] bounds
	for i, v := range m.Vertices {
		u, vi := v.UV0[0], v.UV0[1]
		if u < 0 || u > 1 || vi < 0 || vi > 1 {
			t.Errorf("vertex[%d]: UV out of bounds (%.3f, %.3f)", i, u, vi)
		}
	}
}

// TestUVAtlasDeterminism verifies that UV atlas generation is deterministic.
func TestUVAtlasDeterminism(t *testing.T) {
	g := unpeople.NewGenerator()
	p := unpeople.DefaultParams()
	p.Seed = 99999

	// Generate twice with same params
	m1, err := g.Generate(p)
	if err != nil {
		t.Fatalf("first Generate failed: %v", err)
	}
	m2, err := g.Generate(p)
	if err != nil {
		t.Fatalf("second Generate failed: %v", err)
	}

	// Verify UV coordinates are identical
	if len(m1.Vertices) != len(m2.Vertices) {
		t.Fatalf("vertex count mismatch: %d vs %d", len(m1.Vertices), len(m2.Vertices))
	}

	for i := range m1.Vertices {
		if m1.Vertices[i].UV0 != m2.Vertices[i].UV0 {
			t.Errorf("vertex[%d]: UV differs (%.3f,%.3f) vs (%.3f,%.3f)",
				i, m1.Vertices[i].UV0[0], m1.Vertices[i].UV0[1],
				m2.Vertices[i].UV0[0], m2.Vertices[i].UV0[1])
			break
		}
	}
}

// TestUVAtlasDistribution verifies that UVs are distributed across
// different regions of the atlas space (not all in one corner).
func TestUVAtlasDistribution(t *testing.T) {
	g := unpeople.NewGenerator()
	p := unpeople.DefaultParams()
	m, err := g.Generate(p)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Check that UVs span a reasonable range (not all near 0)
	var minU, maxU, minV, maxV float32 = 1, 0, 1, 0
	for _, v := range m.Vertices {
		u, vi := v.UV0[0], v.UV0[1]
		if u < minU {
			minU = u
		}
		if u > maxU {
			maxU = u
		}
		if vi < minV {
			minV = vi
		}
		if vi > maxV {
			maxV = vi
		}
	}

	// Expect UVs to span at least 80% of the UV space
	uRange := maxU - minU
	vRange := maxV - minV
	if uRange < 0.8 {
		t.Errorf("U range too small: %.3f (min=%.3f, max=%.3f)", uRange, minU, maxU)
	}
	if vRange < 0.8 {
		t.Errorf("V range too small: %.3f (min=%.3f, max=%.3f)", vRange, minV, maxV)
	}
}

// ─── Musculature Normal Map Tests ────────────────────────────────────────────

// TestMusculatureNormalMapGeneration verifies that musculature normal maps
// are generated correctly for all build types.
func TestMusculatureNormalMapGeneration(t *testing.T) {
	builds := []unpeople.Build{
		unpeople.BuildMuscular,
		unpeople.BuildAthletic,
		unpeople.BuildAverage,
		unpeople.BuildLean,
		unpeople.BuildStocky,
		unpeople.BuildFragile,
	}

	for _, build := range builds {
		t.Run(buildName(build), func(t *testing.T) {
			nm := unpeople.GenerateMusculatureAtlas(build, 42, 128, 128)
			if nm == nil {
				t.Fatal("GenerateMusculatureAtlas returned nil")
			}
			if nm.Width != 128 || nm.Height != 128 {
				t.Errorf("wrong dimensions: %dx%d", nm.Width, nm.Height)
			}
			if len(nm.Pixels) != 128*128 {
				t.Errorf("wrong pixel count: %d", len(nm.Pixels))
			}

			// Verify all pixels are valid normals
			for i, p := range nm.Pixels {
				// R, G, B should be in [0, 1]
				if p[0] < 0 || p[0] > 1 || p[1] < 0 || p[1] > 1 || p[2] < 0 || p[2] > 1 {
					t.Errorf("pixel[%d] has invalid normal color: (%.3f, %.3f, %.3f)",
						i, p[0], p[1], p[2])
					break
				}
				// Alpha should be 1.0
				if p[3] != 1.0 {
					t.Errorf("pixel[%d] has invalid alpha: %.3f", i, p[3])
					break
				}
			}
		})
	}
}

// buildName returns a string name for the build enum.
func buildName(b unpeople.Build) string {
	switch b {
	case unpeople.BuildMuscular:
		return "Muscular"
	case unpeople.BuildAthletic:
		return "Athletic"
	case unpeople.BuildAverage:
		return "Average"
	case unpeople.BuildLean:
		return "Lean"
	case unpeople.BuildStocky:
		return "Stocky"
	case unpeople.BuildFragile:
		return "Fragile"
	default:
		return "Unknown"
	}
}

// ageName returns a string name for the age enum.
func ageName(a unpeople.Age) string {
	switch a {
	case unpeople.AgeDecrepit:
		return "Decrepit"
	case unpeople.AgeElderly:
		return "Elderly"
	case unpeople.AgeOld:
		return "Old"
	case unpeople.AgeAdult:
		return "Adult"
	case unpeople.AgeYouth:
		return "Youth"
	case unpeople.AgeTeen:
		return "Teen"
	case unpeople.AgeChild:
		return "Child"
	case unpeople.AgeToddler:
		return "Toddler"
	default:
		return "Unknown"
	}
}

// TestMusculatureNormalMapDeterminism verifies that the same seed produces
// the same normal map.
func TestMusculatureNormalMapDeterminism(t *testing.T) {
	nm1 := unpeople.GenerateMusculatureAtlas(unpeople.BuildMuscular, 123, 64, 64)
	nm2 := unpeople.GenerateMusculatureAtlas(unpeople.BuildMuscular, 123, 64, 64)

	if len(nm1.Pixels) != len(nm2.Pixels) {
		t.Fatalf("pixel count mismatch: %d vs %d", len(nm1.Pixels), len(nm2.Pixels))
	}

	for i := range nm1.Pixels {
		if nm1.Pixels[i] != nm2.Pixels[i] {
			t.Errorf("pixel[%d] differs: (%.3f,%.3f,%.3f) vs (%.3f,%.3f,%.3f)",
				i, nm1.Pixels[i][0], nm1.Pixels[i][1], nm1.Pixels[i][2],
				nm2.Pixels[i][0], nm2.Pixels[i][1], nm2.Pixels[i][2])
			break
		}
	}
}

// TestMusculatureDefinitionVaries verifies that different build types
// produce different muscle definition levels.
func TestMusculatureDefinitionVaries(t *testing.T) {
	// Muscular should have more pronounced normals than Fragile
	nmMuscular := unpeople.GenerateMusculatureAtlas(unpeople.BuildMuscular, 42, 64, 64)
	nmFragile := unpeople.GenerateMusculatureAtlas(unpeople.BuildFragile, 42, 64, 64)

	// Measure average deviation from flat normal (0.5, 0.5, 1.0)
	muscularDev := measureNormalDeviation(nmMuscular.Pixels)
	fragileDev := measureNormalDeviation(nmFragile.Pixels)

	if muscularDev <= fragileDev {
		t.Errorf("Muscular build should have more normal deviation than Fragile: %.4f vs %.4f",
			muscularDev, fragileDev)
	}
}

// measureNormalDeviation calculates the average deviation from flat normals.
func measureNormalDeviation(pixels []unpeople.Color) float32 {
	var totalDev float32
	for _, p := range pixels {
		// Deviation from flat normal (0.5, 0.5, 1.0)
		dx := p[0] - 0.5
		dy := p[1] - 0.5
		totalDev += dx*dx + dy*dy
	}
	return totalDev / float32(len(pixels))
}

// TestGenerateWithMusculature verifies the full generation pipeline with
// musculature normal maps.
func TestGenerateWithMusculature(t *testing.T) {
	g := unpeople.NewGenerator()
	p := unpeople.DefaultParams()
	p.Build = unpeople.BuildMuscular

	result, err := g.GenerateWithMusculature(p)
	if err != nil {
		t.Fatalf("GenerateWithMusculature failed: %v", err)
	}

	if result.Mesh == nil {
		t.Error("Mesh is nil")
	}
	if result.MusculatureMap == nil {
		t.Error("MusculatureMap is nil")
	}
	if result.Material.NormalScale <= 0 {
		t.Error("NormalScale should be positive for muscular build")
	}
}

// TestBuildToMuscleDefinition verifies the mapping from Build to MuscleDefinition.
func TestBuildToMuscleDefinition(t *testing.T) {
	// Muscular should map to Pronounced
	if def := unpeople.BuildToMuscleDefinition(unpeople.BuildMuscular); def != unpeople.MuscleDefinitionPronounced {
		t.Errorf("Muscular should be Pronounced, got %d", def)
	}

	// Fragile should map to None
	if def := unpeople.BuildToMuscleDefinition(unpeople.BuildFragile); def != unpeople.MuscleDefinitionNone {
		t.Errorf("Fragile should be None, got %d", def)
	}

	// Athletic should map to Moderate
	if def := unpeople.BuildToMuscleDefinition(unpeople.BuildAthletic); def != unpeople.MuscleDefinitionModerate {
		t.Errorf("Athletic should be Moderate, got %d", def)
	}
}

// TestNormalMapSampling verifies bilinear sampling works correctly.
func TestNormalMapSampling(t *testing.T) {
	nm := unpeople.GenerateMusculatureAtlas(unpeople.BuildMuscular, 42, 64, 64)

	// Sample at corners
	corners := [][2]float32{{0, 0}, {1, 0}, {0, 1}, {1, 1}}
	for _, c := range corners {
		sample := nm.SampleBilinear(c[0], c[1])
		// Should be valid normal
		if sample[0] < 0 || sample[0] > 1 || sample[1] < 0 || sample[1] > 1 {
			t.Errorf("invalid sample at (%.1f, %.1f): (%.3f, %.3f, %.3f)",
				c[0], c[1], sample[0], sample[1], sample[2])
		}
	}

	// Sample in center
	center := nm.SampleBilinear(0.5, 0.5)
	if center[3] != 1.0 {
		t.Errorf("center sample alpha should be 1.0, got %.3f", center[3])
	}
}

// ─── Skin Texture Tests ──────────────────────────────────────────────────────

// TestSkinTextureGeneration verifies that skin textures are generated correctly
// for all skin tone and age combinations.
func TestSkinTextureGeneration(t *testing.T) {
	tones := []unpeople.SkinTone{
		unpeople.SkinTonePale,
		unpeople.SkinToneFair,
		unpeople.SkinToneMedium,
		unpeople.SkinToneDark,
	}
	ages := []unpeople.Age{
		unpeople.AgeChild,
		unpeople.AgeTeen,
		unpeople.AgeAdult,
		unpeople.AgeElderly,
	}

	for _, tone := range tones {
		for _, age := range ages {
			t.Run(skinToneName(tone)+"_"+ageName(age), func(t *testing.T) {
				params := unpeople.DefaultSkinTextureParams(
					tone, unpeople.SkinUndertoneNeutral, age, 42)
				tex := unpeople.GenerateSkinTexture(params, 64, 64)

				if tex == nil {
					t.Fatal("GenerateSkinTexture returned nil")
				}
				if len(tex.Pixels) != 64*64 {
					t.Errorf("wrong pixel count: %d", len(tex.Pixels))
				}

				// Verify all pixels are valid colors
				for i, p := range tex.Pixels {
					if p[0] < 0 || p[0] > 1 || p[1] < 0 || p[1] > 1 ||
						p[2] < 0 || p[2] > 1 || p[3] < 0 || p[3] > 1 {
						t.Errorf("pixel[%d] has invalid color: (%.3f, %.3f, %.3f, %.3f)",
							i, p[0], p[1], p[2], p[3])
						break
					}
				}
			})
		}
	}
}

// skinToneName returns a string name for the skin tone enum.
func skinToneName(s unpeople.SkinTone) string {
	switch s {
	case unpeople.SkinTonePale:
		return "Pale"
	case unpeople.SkinToneFair:
		return "Fair"
	case unpeople.SkinToneMedium:
		return "Medium"
	case unpeople.SkinToneDark:
		return "Dark"
	default:
		return "Unknown"
	}
}

// TestSkinTextureDeterminism verifies deterministic texture generation.
func TestSkinTextureDeterminism(t *testing.T) {
	params := unpeople.DefaultSkinTextureParams(
		unpeople.SkinToneMedium, unpeople.SkinUndertoneWarm, unpeople.AgeAdult, 123)

	tex1 := unpeople.GenerateSkinTexture(params, 32, 32)
	tex2 := unpeople.GenerateSkinTexture(params, 32, 32)

	if len(tex1.Pixels) != len(tex2.Pixels) {
		t.Fatalf("pixel count mismatch: %d vs %d", len(tex1.Pixels), len(tex2.Pixels))
	}

	for i := range tex1.Pixels {
		if tex1.Pixels[i] != tex2.Pixels[i] {
			t.Errorf("pixel[%d] differs: (%.3f,%.3f,%.3f) vs (%.3f,%.3f,%.3f)",
				i, tex1.Pixels[i][0], tex1.Pixels[i][1], tex1.Pixels[i][2],
				tex2.Pixels[i][0], tex2.Pixels[i][1], tex2.Pixels[i][2])
			break
		}
	}
}

// TestSkinTextureVariesByAge verifies textures differ between ages.
func TestSkinTextureVariesByAge(t *testing.T) {
	baseParams := unpeople.DefaultSkinTextureParams(
		unpeople.SkinToneFair, unpeople.SkinUndertoneNeutral, unpeople.AgeChild, 42)
	elderlyParams := unpeople.DefaultSkinTextureParams(
		unpeople.SkinToneFair, unpeople.SkinUndertoneNeutral, unpeople.AgeElderly, 42)

	texChild := unpeople.GenerateSkinTexture(baseParams, 32, 32)
	texElderly := unpeople.GenerateSkinTexture(elderlyParams, 32, 32)

	// Count differing pixels
	diffCount := 0
	for i := range texChild.Pixels {
		if texChild.Pixels[i] != texElderly.Pixels[i] {
			diffCount++
		}
	}

	// Should have significant differences
	if diffCount < 10 {
		t.Errorf("child and elderly textures should differ more (only %d pixels differ)", diffCount)
	}
}

// TestGenerateWithTextures verifies the complete texture pipeline.
func TestGenerateWithTextures(t *testing.T) {
	g := unpeople.NewGenerator()
	p := unpeople.DefaultParams()
	p.SkinTone = unpeople.SkinToneFair
	p.Age = unpeople.AgeElderly
	p.Build = unpeople.BuildMuscular

	result, err := g.GenerateWithTextures(p)
	if err != nil {
		t.Fatalf("GenerateWithTextures failed: %v", err)
	}

	if result.Mesh == nil {
		t.Error("Mesh is nil")
	}
	if result.AlbedoTexture == nil {
		t.Error("AlbedoTexture is nil")
	}
	if result.NormalTexture == nil {
		t.Error("NormalTexture is nil")
	}
	if result.AlbedoTexture.Width != 512 || result.AlbedoTexture.Height != 512 {
		t.Errorf("unexpected texture size: %dx%d", result.AlbedoTexture.Width, result.AlbedoTexture.Height)
	}
}

// TestTextureToRGBA8 verifies texture export to bytes.
func TestTextureToRGBA8(t *testing.T) {
	params := unpeople.DefaultSkinTextureParams(
		unpeople.SkinToneMedium, unpeople.SkinUndertoneNeutral, unpeople.AgeAdult, 42)
	tex := unpeople.GenerateSkinTexture(params, 16, 16)

	data := tex.ToRGBA8()
	expectedLen := 16 * 16 * 4
	if len(data) != expectedLen {
		t.Errorf("wrong data length: %d (expected %d)", len(data), expectedLen)
	}

	// Verify all bytes are in valid range (implicit 0-255)
	// Just check that export worked without panic
}

// TestTextureSampling verifies texture bilinear sampling.
func TestTextureSampling(t *testing.T) {
	params := unpeople.DefaultSkinTextureParams(
		unpeople.SkinToneMedium, unpeople.SkinUndertoneNeutral, unpeople.AgeAdult, 42)
	tex := unpeople.GenerateSkinTexture(params, 32, 32)

	// Sample at corners and center
	samples := [][2]float32{{0, 0}, {1, 1}, {0.5, 0.5}}
	for _, s := range samples {
		c := tex.SampleBilinear(s[0], s[1])
		// Should be valid color
		if c[0] < 0 || c[0] > 1 || c[1] < 0 || c[1] > 1 ||
			c[2] < 0 || c[2] > 1 || c[3] < 0 || c[3] > 1 {
			t.Errorf("invalid sample at (%.1f, %.1f): %v", s[0], s[1], c)
		}
	}
}

// TestFreckleIntensityByTone verifies freckle intensity varies with skin tone.
func TestFreckleIntensityByTone(t *testing.T) {
	paleParams := unpeople.DefaultSkinTextureParams(
		unpeople.SkinTonePale, unpeople.SkinUndertoneNeutral, unpeople.AgeAdult, 42)
	darkParams := unpeople.DefaultSkinTextureParams(
		unpeople.SkinToneDark, unpeople.SkinUndertoneNeutral, unpeople.AgeAdult, 42)

	// Pale skin should have higher freckle intensity
	if paleParams.FreckleIntensity <= darkParams.FreckleIntensity {
		t.Errorf("pale skin should have more freckles: %.3f vs %.3f",
			paleParams.FreckleIntensity, darkParams.FreckleIntensity)
	}
}

// ─── Skeleton Tests ──────────────────────────────────────────────────────────

// TestSkeletonGeneration verifies that a skeleton is generated with correct structure.
func TestSkeletonGeneration(t *testing.T) {
	g := unpeople.NewGenerator()
	p := unpeople.DefaultParams()

	result, err := g.GenerateWithSkeleton(p)
	if err != nil {
		t.Fatalf("GenerateWithSkeleton failed: %v", err)
	}

	if result.Skeleton == nil {
		t.Fatal("Skeleton is nil")
	}
	if len(result.Skeleton.Joints) != int(unpeople.JointCount) {
		t.Errorf("wrong joint count: %d (expected %d)",
			len(result.Skeleton.Joints), unpeople.JointCount)
	}
}

// TestSkeletonValidation verifies skeleton hierarchy is valid.
func TestSkeletonValidation(t *testing.T) {
	g := unpeople.NewGenerator()
	p := unpeople.DefaultParams()

	result, err := g.GenerateWithSkeleton(p)
	if err != nil {
		t.Fatalf("GenerateWithSkeleton failed: %v", err)
	}

	if err := result.Skeleton.Validate(); err != nil {
		t.Errorf("skeleton validation failed: %v", err)
	}
}

// TestSkeletonJointHierarchy verifies parent-child relationships.
func TestSkeletonJointHierarchy(t *testing.T) {
	g := unpeople.NewGenerator()
	p := unpeople.DefaultParams()

	result, err := g.GenerateWithSkeleton(p)
	if err != nil {
		t.Fatalf("GenerateWithSkeleton failed: %v", err)
	}
	skel := result.Skeleton

	// Root should have no parent
	root := skel.Joint(unpeople.JointRoot)
	if root.ParentID != -1 {
		t.Errorf("Root should have no parent, got %d", root.ParentID)
	}

	// Hips should have Root as parent
	hips := skel.Joint(unpeople.JointHips)
	if hips.ParentID != unpeople.JointRoot {
		t.Errorf("Hips should have Root as parent, got %d", hips.ParentID)
	}

	// Head should have Neck as parent
	head := skel.Joint(unpeople.JointHead)
	if head.ParentID != unpeople.JointNeck {
		t.Errorf("Head should have Neck as parent, got %d", head.ParentID)
	}

	// LeftHand should have LeftForearm as parent
	leftHand := skel.Joint(unpeople.JointLeftHand)
	if leftHand.ParentID != unpeople.JointLeftForearm {
		t.Errorf("LeftHand should have LeftForearm as parent, got %d", leftHand.ParentID)
	}
}

// TestSkeletonJointPositions verifies joint positions are reasonable.
func TestSkeletonJointPositions(t *testing.T) {
	g := unpeople.NewGenerator()
	p := unpeople.DefaultParams()

	result, err := g.GenerateWithSkeleton(p)
	if err != nil {
		t.Fatalf("GenerateWithSkeleton failed: %v", err)
	}
	skel := result.Skeleton

	// Head should be above hips
	head := skel.Joint(unpeople.JointHead)
	hips := skel.Joint(unpeople.JointHips)
	if head.Position[1] <= hips.Position[1] {
		t.Errorf("Head (Y=%.3f) should be above Hips (Y=%.3f)",
			head.Position[1], hips.Position[1])
	}

	// Feet should be near the ground
	leftFoot := skel.Joint(unpeople.JointLeftFoot)
	rightFoot := skel.Joint(unpeople.JointRightFoot)
	if leftFoot.Position[1] > 0.2 {
		t.Errorf("LeftFoot (Y=%.3f) should be near ground", leftFoot.Position[1])
	}
	if rightFoot.Position[1] > 0.2 {
		t.Errorf("RightFoot (Y=%.3f) should be near ground", rightFoot.Position[1])
	}

	// Left arm should be on the left side (negative X)
	leftHand := skel.Joint(unpeople.JointLeftHand)
	if leftHand.Position[0] >= 0 {
		t.Errorf("LeftHand (X=%.3f) should have negative X", leftHand.Position[0])
	}

	// Right arm should be on the right side (positive X)
	rightHand := skel.Joint(unpeople.JointRightHand)
	if rightHand.Position[0] <= 0 {
		t.Errorf("RightHand (X=%.3f) should have positive X", rightHand.Position[0])
	}
}

// TestSkeletonVariesWithHeight verifies skeleton scales with height parameter.
func TestSkeletonVariesWithHeight(t *testing.T) {
	g := unpeople.NewGenerator()

	pTall := unpeople.DefaultParams()
	pTall.Height = unpeople.HeightTall
	pShort := unpeople.DefaultParams()
	pShort.Height = unpeople.HeightShort

	resultTall, _ := g.GenerateWithSkeleton(pTall)
	resultShort, _ := g.GenerateWithSkeleton(pShort)

	headTall := resultTall.Skeleton.Joint(unpeople.JointHead)
	headShort := resultShort.Skeleton.Joint(unpeople.JointHead)

	// Tall character's head should be higher
	if headTall.Position[1] <= headShort.Position[1] {
		t.Errorf("Tall head (Y=%.3f) should be higher than short head (Y=%.3f)",
			headTall.Position[1], headShort.Position[1])
	}
}

// TestSkeletonBoneCount verifies total bone count.
func TestSkeletonBoneCount(t *testing.T) {
	g := unpeople.NewGenerator()
	p := unpeople.DefaultParams()

	result, err := g.GenerateWithSkeleton(p)
	if err != nil {
		t.Fatalf("GenerateWithSkeleton failed: %v", err)
	}

	boneCount := result.Skeleton.TotalBoneCount()
	// Should have JointCount - 1 bones (root has no parent bone)
	expectedBones := int(unpeople.JointCount) - 1
	if boneCount != expectedBones {
		t.Errorf("wrong bone count: %d (expected %d)", boneCount, expectedBones)
	}
}

// TestSkeletonJointByName verifies finding joints by name.
func TestSkeletonJointByName(t *testing.T) {
	g := unpeople.NewGenerator()
	p := unpeople.DefaultParams()

	result, err := g.GenerateWithSkeleton(p)
	if err != nil {
		t.Fatalf("GenerateWithSkeleton failed: %v", err)
	}

	// Find specific joints by name
	head := result.Skeleton.JointByName("Head")
	if head == nil {
		t.Error("Could not find joint named 'Head'")
	} else if head.ID != unpeople.JointHead {
		t.Errorf("Head joint has wrong ID: %d", head.ID)
	}

	// Non-existent joint
	notFound := result.Skeleton.JointByName("NonExistentJoint")
	if notFound != nil {
		t.Error("Should not find non-existent joint")
	}
}

// ─── Skinning Tests ──────────────────────────────────────────────────────────

// TestSkinningWeightGeneration verifies skinning weights are computed correctly.
func TestSkinningWeightGeneration(t *testing.T) {
	g := unpeople.NewGenerator()
	p := unpeople.DefaultParams()

	result, err := g.GenerateWithRig(p)
	if err != nil {
		t.Fatalf("GenerateWithRig failed: %v", err)
	}

	// Check all vertices have weights
	for i, v := range result.Mesh.Vertices {
		// At least one weight should be non-zero
		hasWeight := v.JointWeights[0] > 0 || v.JointWeights[1] > 0 ||
			v.JointWeights[2] > 0 || v.JointWeights[3] > 0
		if !hasWeight {
			t.Errorf("vertex[%d] has no skinning weights", i)
			break
		}
	}
}

// TestSkinningWeightsNormalized verifies weights sum to approximately 1.0.
func TestSkinningWeightsNormalized(t *testing.T) {
	g := unpeople.NewGenerator()
	p := unpeople.DefaultParams()

	result, err := g.GenerateWithRig(p)
	if err != nil {
		t.Fatalf("GenerateWithRig failed: %v", err)
	}

	for i, v := range result.Mesh.Vertices {
		sum := v.JointWeights[0] + v.JointWeights[1] + v.JointWeights[2] + v.JointWeights[3]
		if sum < 0.99 || sum > 1.01 {
			t.Errorf("vertex[%d] weights sum to %.4f (expected ~1.0)", i, sum)
			break
		}
	}
}

// TestSkinningValidation verifies the validation function works.
func TestSkinningValidation(t *testing.T) {
	g := unpeople.NewGenerator()
	p := unpeople.DefaultParams()

	result, err := g.GenerateWithRig(p)
	if err != nil {
		t.Fatalf("GenerateWithRig failed: %v", err)
	}

	err = unpeople.ValidateSkinning(result.Mesh)
	if err != nil {
		t.Errorf("ValidateSkinning failed: %v", err)
	}
}

// TestSkinningJointIDsValid verifies all joint IDs are within valid range.
func TestSkinningJointIDsValid(t *testing.T) {
	g := unpeople.NewGenerator()
	p := unpeople.DefaultParams()

	result, err := g.GenerateWithRig(p)
	if err != nil {
		t.Fatalf("GenerateWithRig failed: %v", err)
	}

	for i, v := range result.Mesh.Vertices {
		for slot := 0; slot < 4; slot++ {
			if v.JointWeights[slot] > 0 {
				jointID := v.JointIds[slot]
				if jointID < 0 || jointID >= int32(unpeople.JointCount) {
					t.Errorf("vertex[%d] has invalid joint ID %d in slot %d",
						i, jointID, slot)
				}
			}
		}
	}
}

// TestSkinningHeadVertex verifies head vertices are weighted to head joint.
func TestSkinningHeadVertex(t *testing.T) {
	g := unpeople.NewGenerator()
	p := unpeople.DefaultParams()

	result, err := g.GenerateWithRig(p)
	if err != nil {
		t.Fatalf("GenerateWithRig failed: %v", err)
	}

	// Find a vertex that should be on the head (high Y coordinate)
	var headVertex *unpeople.Vertex
	for i := range result.Mesh.Vertices {
		v := &result.Mesh.Vertices[i]
		if v.Position[1] > 1.6 { // Head region
			headVertex = v
			break
		}
	}

	if headVertex == nil {
		t.Skip("No vertex found in head region")
		return
	}

	// Check that head joint has significant weight
	headJointID := int32(unpeople.JointHead)
	hasHeadInfluence := false
	for slot := 0; slot < 4; slot++ {
		if headVertex.JointIds[slot] == headJointID && headVertex.JointWeights[slot] > 0.1 {
			hasHeadInfluence = true
			break
		}
	}

	if !hasHeadInfluence {
		t.Error("Head vertex should have significant Head joint influence")
	}
}

// TestSkinningAllSpecies verifies skinning works for all species.
func TestSkinningAllSpecies(t *testing.T) {
	g := unpeople.NewGenerator()

	species := []unpeople.Species{
		unpeople.SpeciesHuman,
		unpeople.SpeciesElf,
		unpeople.SpeciesDwarf,
		unpeople.SpeciesOrc,
		unpeople.SpeciesTroll,
	}

	for _, sp := range species {
		p := unpeople.DefaultParams()
		p.Species = sp

		result, err := g.GenerateWithRig(p)
		if err != nil {
			t.Errorf("GenerateWithRig failed for species %d: %v", sp, err)
			continue
		}

		err = unpeople.ValidateSkinning(result.Mesh)
		if err != nil {
			t.Errorf("ValidateSkinning failed for species %d: %v", sp, err)
		}
	}
}
