package unpeople

// ─── UV Atlas Generation ─────────────────────────────────────────────────────
//
// This file implements UV atlas generation for the humanoid mesh. Instead of
// each body part using the full [0,1]² UV space (which causes texture overlap),
// this system partitions the UV space into non-overlapping regions for each
// body part, enabling proper texture mapping.
//
// The layout is deterministic and follows a consistent pattern:
// - Top row: Head, face, ears
// - Middle row: Torso (neck, chest, abdomen, hips)
// - Lower rows: Arms and legs

// UVRegion defines a rectangular region in UV space for a single body part.
type UVRegion struct {
	UMin, UMax float32 // Horizontal bounds [0,1]
	VMin, VMax float32 // Vertical bounds [0,1]
}

// Width returns the horizontal span of the region.
func (r UVRegion) Width() float32 { return r.UMax - r.UMin }

// Height returns the vertical span of the region.
func (r UVRegion) Height() float32 { return r.VMax - r.VMin }

// Transform maps a local UV coordinate [0,1]² into the atlas region.
func (r UVRegion) Transform(localU, localV float32) Vec2 {
	return Vec2{
		r.UMin + localU*r.Width(),
		r.VMin + localV*r.Height(),
	}
}

// UVAtlas maps body part names to their UV regions.
// The atlas guarantees non-overlapping regions for all body parts.
type UVAtlas struct {
	Head      UVRegion
	Face      UVRegion
	SkullCap  UVRegion
	Neck      UVRegion
	Chest     UVRegion
	Abdomen   UVRegion
	Hips      UVRegion
	UpperArmL UVRegion
	UpperArmR UVRegion
	ForearmL  UVRegion
	ForearmR  UVRegion
	HandL     UVRegion
	HandR     UVRegion
	UpperLegL UVRegion
	UpperLegR UVRegion
	LowerLegL UVRegion
	LowerLegR UVRegion
	FootL     UVRegion
	FootR     UVRegion
	EarL      UVRegion
	EarR      UVRegion
}

// defaultUVAtlas returns the standard UV atlas layout for a humanoid mesh.
// The layout is designed for efficient texture packing with minimal waste:
//
//	Row 1 (V: 0.70-1.00): Head, Face, SkullCap, Ears
//	Row 2 (V: 0.50-0.70): Neck, Chest, Abdomen, Hips
//	Row 3 (V: 0.25-0.50): Arms (Upper + Forearm for each side)
//	Row 4 (V: 0.00-0.25): Legs (Upper + Lower for each side) + Hands/Feet
func defaultUVAtlas() UVAtlas {
	return UVAtlas{
		// Row 1: Head region (top)
		Head:     UVRegion{0.00, 0.40, 0.70, 1.00}, // Large: ellipsoid
		Face:     UVRegion{0.40, 0.65, 0.75, 1.00}, // Face overlay
		SkullCap: UVRegion{0.65, 0.80, 0.85, 1.00}, // Hair slot
		EarL:     UVRegion{0.80, 0.90, 0.85, 1.00}, // Left ear
		EarR:     UVRegion{0.90, 1.00, 0.85, 1.00}, // Right ear

		// Row 2: Torso (middle-upper)
		Neck:    UVRegion{0.00, 0.15, 0.50, 0.70}, // Short cylinder
		Chest:   UVRegion{0.15, 0.45, 0.50, 0.70}, // Large torso
		Abdomen: UVRegion{0.45, 0.70, 0.50, 0.70}, // Medium
		Hips:    UVRegion{0.70, 1.00, 0.50, 0.70}, // Pelvis

		// Row 3: Arms (middle-lower)
		UpperArmL: UVRegion{0.00, 0.25, 0.25, 0.50}, // Left upper arm
		ForearmL:  UVRegion{0.25, 0.50, 0.25, 0.50}, // Left forearm
		UpperArmR: UVRegion{0.50, 0.75, 0.25, 0.50}, // Right upper arm
		ForearmR:  UVRegion{0.75, 1.00, 0.25, 0.50}, // Right forearm

		// Row 4: Legs + Hands/Feet (bottom)
		UpperLegL: UVRegion{0.00, 0.20, 0.00, 0.25}, // Left upper leg
		LowerLegL: UVRegion{0.20, 0.35, 0.00, 0.25}, // Left lower leg
		HandL:     UVRegion{0.35, 0.50, 0.00, 0.25}, // Left hand + fingers
		UpperLegR: UVRegion{0.50, 0.70, 0.00, 0.25}, // Right upper leg
		LowerLegR: UVRegion{0.70, 0.85, 0.00, 0.25}, // Right lower leg
		HandR:     UVRegion{0.85, 1.00, 0.00, 0.25}, // Right hand + fingers
		FootL:     UVRegion{0.00, 0.15, 0.70, 0.85}, // Left foot + toes (tucked)
		FootR:     UVRegion{0.15, 0.30, 0.70, 0.85}, // Right foot + toes
	}
}

// remapUVs transforms all vertex UV coordinates from local [0,1]² space
// into the specified atlas region. This is applied after primitive generation.
func remapUVs(verts []Vertex, region UVRegion) {
	for i := range verts {
		v := &verts[i]
		v.UV0 = region.Transform(v.UV0[0], v.UV0[1])
	}
}
