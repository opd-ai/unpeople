// Package unpeople – material descriptor
//
// Material provides a Kaiju-compatible material description for humanoid meshes.
// The Material struct is layout-compatible with Kaiju's standard shader data
// structures, allowing downstream systems to convert it directly to
// rendering.ShaderDataStandard or similar types.
package unpeople

// Material describes the surface properties of a generated humanoid mesh.
// It is designed to be compatible with Kaiju's PBR-based rendering pipeline
// while remaining engine-agnostic for use in other contexts.
type Material struct {
	// ShaderName indicates which shader pipeline should render this material.
	// Common values: "standard", "standard_skinned", "pbr", "unlit"
	ShaderName string

	// Albedo is the base color of the material in linear RGB space.
	// For skin, this matches the SkinTone and SkinUndertone parameters.
	// Alpha channel controls overall opacity (typically 1.0 for skin).
	Albedo Color

	// Metallic controls how metallic the surface appears (0.0 = dielectric, 1.0 = metal).
	// Skin is non-metallic, so this is typically 0.0.
	Metallic float32

	// Roughness controls the microsurface roughness (0.0 = mirror, 1.0 = rough).
	// Skin typically has roughness in the 0.3-0.6 range depending on moisture.
	Roughness float32

	// AmbientOcclusion modulates indirect lighting (1.0 = no occlusion).
	AmbientOcclusion float32

	// SubsurfaceScattering controls light transmission through thin surfaces.
	// Skin benefits from SSS for a more realistic appearance.
	// Value in [0, 1] where 0 = no SSS, 1 = full SSS.
	SubsurfaceScattering float32

	// SubsurfaceColor is the tint applied to scattered light under the surface.
	// For skin, this is typically a warm red/orange tone.
	SubsurfaceColor Color

	// NormalScale multiplies the normal map intensity (1.0 = normal, 0.0 = flat).
	// Useful for controlling detail prominence.
	NormalScale float32

	// EmissiveColor adds self-illumination to the material.
	// Typically {0,0,0,0} for skin unless special effects are desired.
	EmissiveColor Color

	// Properties contains additional named float parameters.
	// Allows extending the material without modifying the struct.
	Properties map[string]float32

	// TextureSlots contains named texture references for texture-based materials.
	// Keys are slot names (e.g., "albedo", "normal", "roughness").
	// Values are texture identifiers/paths (interpretation depends on renderer).
	TextureSlots map[string]string
}

// DefaultSkinMaterial returns a Material configured for typical human skin
// using the given skin color. The material is optimized for PBR rendering
// with appropriate subsurface scattering values.
func DefaultSkinMaterial(skinColor Color) Material {
	return Material{
		ShaderName:           "standard",
		Albedo:               skinColor,
		Metallic:             0.0,                       // Skin is non-metallic
		Roughness:            0.45,                      // Moderately rough skin
		AmbientOcclusion:     1.0,                       // No baked AO
		SubsurfaceScattering: 0.3,                       // Light SSS for skin translucency
		SubsurfaceColor:      Color{0.8, 0.3, 0.2, 1.0}, // Warm red undertone
		NormalScale:          1.0,
		EmissiveColor:        Color{0, 0, 0, 0}, // No emission
		Properties:           nil,
		TextureSlots:         nil,
	}
}

// SSSkinMaterial returns a Material configured for skin with enhanced
// subsurface scattering for more realistic appearance in engines that
// support advanced SSS techniques.
func SSSkinMaterial(skinColor Color) Material {
	m := DefaultSkinMaterial(skinColor)
	m.ShaderName = "pbr"
	m.SubsurfaceScattering = 0.5 // Enhanced SSS
	m.Roughness = 0.4            // Slightly smoother
	return m
}

// UnlitMaterial returns a simple unlit material using only the albedo color.
// Useful for preview rendering or stylized visuals.
func UnlitMaterial(color Color) Material {
	return Material{
		ShaderName:       "unlit",
		Albedo:           color,
		Metallic:         0.0,
		Roughness:        1.0,
		AmbientOcclusion: 1.0,
		Properties:       nil,
		TextureSlots:     nil,
	}
}

// MeshWithMaterial bundles a generated mesh with its corresponding material.
// This is the recommended output format when material information is needed
// alongside the geometry.
type MeshWithMaterial struct {
	Mesh     *Mesh
	Material Material
}

// GenerateWithMaterial extends Generate to also produce material data.
// The material's albedo color is set to match the skin tone parameters.
func (g *Generator) GenerateWithMaterial(p Params) (*MeshWithMaterial, error) {
	mesh, err := g.Generate(p)
	if err != nil {
		return nil, err
	}

	skinColor := ComputeSkinColor(p.SkinTone, p.SkinUndertone)
	material := DefaultSkinMaterial(skinColor)

	return &MeshWithMaterial{
		Mesh:     mesh,
		Material: material,
	}, nil
}

// AgeSkinMaterial adjusts the material roughness based on character age.
// Younger skin is smoother; older skin is rougher with less SSS.
func AgeSkinMaterial(skinColor Color, age Age) Material {
	m := DefaultSkinMaterial(skinColor)

	// Adjust roughness based on age
	switch age {
	case AgeToddler, AgeChild:
		m.Roughness = 0.35 // Smoother, more supple skin
		m.SubsurfaceScattering = 0.4
	case AgeTeen, AgeYouth:
		m.Roughness = 0.40
		m.SubsurfaceScattering = 0.35
	case AgeAdult:
		m.Roughness = 0.45
		m.SubsurfaceScattering = 0.30
	case AgeOld:
		m.Roughness = 0.55
		m.SubsurfaceScattering = 0.25
	case AgeElderly:
		m.Roughness = 0.65 // More weathered
		m.SubsurfaceScattering = 0.20
	case AgeDecrepit:
		m.Roughness = 0.75 // Very rough, aged skin
		m.SubsurfaceScattering = 0.15
	}

	return m
}
