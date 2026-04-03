// Package unpeople – OBJ export
//
// ExportOBJ writes a Mesh to the Wavefront OBJ format, a simple text-based
// 3D model format widely supported by DCC tools (Blender, Maya, 3ds Max, etc.).
// This enables debugging and visualization of generated humanoids outside
// the Kaiju engine.
package unpeople

import (
	"bufio"
	"fmt"
	"io"
)

// validateMeshForExport checks that a mesh is suitable for OBJ export.
func validateMeshForExport(mesh *Mesh) error {
	if mesh == nil {
		return fmt.Errorf("mesh is nil")
	}
	if len(mesh.Vertices) == 0 {
		return fmt.Errorf("mesh has no vertices")
	}
	if len(mesh.Indices) == 0 {
		return fmt.Errorf("mesh has no indices")
	}
	if len(mesh.Indices)%3 != 0 {
		return fmt.Errorf("index count (%d) is not a multiple of 3", len(mesh.Indices))
	}
	return nil
}

// writeOBJHeader writes the standard OBJ file header comment block.
func writeOBJHeader(bw *bufio.Writer, mesh *Mesh) error {
	if _, err := fmt.Fprintf(bw, "# Wavefront OBJ exported by unpeople\n"); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(bw, "# Mesh key: %s\n", mesh.Key); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(bw, "# Vertices: %d, Triangles: %d\n\n",
		len(mesh.Vertices), len(mesh.Indices)/3); err != nil {
		return err
	}
	return nil
}

// writeOBJVertexData writes vertex positions, UVs, and normals to the OBJ file.
func writeOBJVertexData(bw *bufio.Writer, vertices []Vertex) error {
	// Vertex positions
	for _, v := range vertices {
		if _, err := fmt.Fprintf(bw, "v %f %f %f\n",
			v.Position[0], v.Position[1], v.Position[2]); err != nil {
			return err
		}
	}
	if _, err := bw.WriteString("\n"); err != nil {
		return err
	}

	// Texture coordinates
	for _, v := range vertices {
		if _, err := fmt.Fprintf(bw, "vt %f %f\n", v.UV0[0], v.UV0[1]); err != nil {
			return err
		}
	}
	if _, err := bw.WriteString("\n"); err != nil {
		return err
	}

	// Vertex normals
	for _, v := range vertices {
		if _, err := fmt.Fprintf(bw, "vn %f %f %f\n",
			v.Normal[0], v.Normal[1], v.Normal[2]); err != nil {
			return err
		}
	}
	if _, err := bw.WriteString("\n"); err != nil {
		return err
	}

	return nil
}

// writeOBJFaces writes triangle face definitions to the OBJ file.
func writeOBJFaces(bw *bufio.Writer, indices []uint32) error {
	for i := 0; i < len(indices); i += 3 {
		// Convert 0-based to 1-based indices
		i1 := indices[i] + 1
		i2 := indices[i+1] + 1
		i3 := indices[i+2] + 1
		if _, err := fmt.Fprintf(bw, "f %d/%d/%d %d/%d/%d %d/%d/%d\n",
			i1, i1, i1, i2, i2, i2, i3, i3, i3); err != nil {
			return err
		}
	}
	return nil
}

// resolveObjectName returns the provided name or "humanoid" as default.
func resolveObjectName(name string) string {
	if name == "" {
		return "humanoid"
	}
	return name
}

// ExportOBJ writes the mesh to w in Wavefront OBJ format.
// The optional objectName sets the 'o' line; if empty, "humanoid" is used.
//
// The OBJ format includes:
//   - Vertex positions (v)
//   - Vertex normals (vn)
//   - Texture coordinates (vt)
//   - Triangular faces (f) with position/texture/normal indices
//
// Vertex colors are not standard OBJ but are exported as comments for reference.
func ExportOBJ(w io.Writer, mesh *Mesh, objectName string) error {
	if err := validateMeshForExport(mesh); err != nil {
		return err
	}

	bw := bufio.NewWriter(w)

	if err := writeOBJHeader(bw, mesh); err != nil {
		return err
	}

	name := resolveObjectName(objectName)
	if _, err := fmt.Fprintf(bw, "o %s\n\n", name); err != nil {
		return err
	}

	if err := writeOBJVertexData(bw, mesh.Vertices); err != nil {
		return err
	}

	if err := writeOBJFaces(bw, mesh.Indices); err != nil {
		return err
	}

	return bw.Flush()
}

// writeMaterialReference writes the MTL library reference and usemtl directive.
func writeMaterialReference(bw *bufio.Writer, mtlW io.Writer, name, mtlName string) error {
	if mtlW != nil && mtlName != "" {
		if _, err := fmt.Fprintf(bw, "mtllib %s\n\n", mtlName); err != nil {
			return err
		}
	}

	if _, err := fmt.Fprintf(bw, "o %s\n", name); err != nil {
		return err
	}

	if mtlW != nil {
		if _, err := fmt.Fprintf(bw, "usemtl %s_mat\n\n", name); err != nil {
			return err
		}
	} else {
		if _, err := bw.WriteString("\n"); err != nil {
			return err
		}
	}
	return nil
}

// ExportOBJWithMTL writes the mesh to objW and an accompanying material
// library to mtlW. The material includes basic skin tone information.
//
// Parameters:
//   - objW: Writer for the .obj file
//   - mtlW: Writer for the .mtl file (may be nil to skip material)
//   - mesh: The mesh to export
//   - material: Material properties (may be nil for default)
//   - objectName: Name for the object (defaults to "humanoid")
//   - mtlName: Name of the material library file (e.g., "skin.mtl")
func ExportOBJWithMTL(objW, mtlW io.Writer, mesh *Mesh, material *Material, objectName, mtlName string) error {
	if mesh == nil {
		return fmt.Errorf("mesh is nil")
	}

	name := resolveObjectName(objectName)

	// Write material library if writer provided
	if mtlW != nil && material != nil {
		if err := writeMTL(mtlW, name+"_mat", material); err != nil {
			return fmt.Errorf("writing MTL: %w", err)
		}
	}

	bw := bufio.NewWriter(objW)

	if err := writeOBJHeader(bw, mesh); err != nil {
		return err
	}

	if err := writeMaterialReference(bw, mtlW, name, mtlName); err != nil {
		return err
	}

	if err := writeOBJVertexData(bw, mesh.Vertices); err != nil {
		return err
	}

	// Smoothing group
	if _, err := bw.WriteString("s 1\n"); err != nil {
		return err
	}

	if err := writeOBJFaces(bw, mesh.Indices); err != nil {
		return err
	}

	return bw.Flush()
}

// writeMTLColor writes a color property line to the MTL file.
func writeMTLColor(bw *bufio.Writer, prefix string, r, g, b float32) error {
	_, err := fmt.Fprintf(bw, "%s %f %f %f\n", prefix, r, g, b)
	return err
}

// writeMTL writes a Wavefront MTL material file.
func writeMTL(w io.Writer, matName string, mat *Material) error {
	bw := bufio.NewWriter(w)

	// Header
	if _, err := fmt.Fprintf(bw, "# Wavefront MTL exported by unpeople\n\n"); err != nil {
		return err
	}

	// Material definition
	if _, err := fmt.Fprintf(bw, "newmtl %s\n", matName); err != nil {
		return err
	}

	// Ambient color (use albedo scaled down)
	if err := writeMTLColor(bw, "Ka", mat.Albedo[0]*0.1, mat.Albedo[1]*0.1, mat.Albedo[2]*0.1); err != nil {
		return err
	}

	// Diffuse color (albedo)
	if err := writeMTLColor(bw, "Kd", mat.Albedo[0], mat.Albedo[1], mat.Albedo[2]); err != nil {
		return err
	}

	// Specular color (based on roughness - rougher = less specular)
	spec := (1.0 - mat.Roughness) * 0.3
	if err := writeMTLColor(bw, "Ks", spec, spec, spec); err != nil {
		return err
	}

	// Specular exponent (shininess) - inverse of roughness
	shininess := (1.0-mat.Roughness)*128.0 + 1.0
	if _, err := fmt.Fprintf(bw, "Ns %f\n", shininess); err != nil {
		return err
	}

	// Transparency (dissolve)
	if _, err := fmt.Fprintf(bw, "d %f\n", mat.Albedo[3]); err != nil {
		return err
	}

	// Illumination model (2 = highlight on)
	if _, err := bw.WriteString("illum 2\n"); err != nil {
		return err
	}

	return bw.Flush()
}
