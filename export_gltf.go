// Package unpeople provides deterministic procedural generation of humanoid meshes.
//
// This file implements glTF 2.0 export for interoperability with standard 3D
// tools and engines. glTF (GL Transmission Format) is the "JPEG of 3D" -
// a royalty-free specification for efficient transmission and loading of 3D
// scenes and models.

package unpeople

import (
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"math"
)

// GLTFExportOptions configures glTF export behavior.
type GLTFExportOptions struct {
	// EmbedBuffers embeds binary data as base64 data URIs (default: true)
	EmbedBuffers bool
	// IncludeNormals includes vertex normals (default: true)
	IncludeNormals bool
	// IncludeUVs includes texture coordinates (default: true)
	IncludeUVs bool
	// IncludeColors includes vertex colors (default: true)
	IncludeColors bool
	// IncludeTangents includes tangent vectors (default: false)
	IncludeTangents bool
	// IncludeSkinning includes joint IDs and weights (default: false)
	IncludeSkinning bool
	// AssetName sets the mesh name in the glTF asset
	AssetName string
}

// DefaultGLTFOptions returns sensible default export options.
func DefaultGLTFOptions() GLTFExportOptions {
	return GLTFExportOptions{
		EmbedBuffers:    true,
		IncludeNormals:  true,
		IncludeUVs:      true,
		IncludeColors:   true,
		IncludeTangents: false,
		IncludeSkinning: false,
		AssetName:       "character",
	}
}

// ExportGLTF writes the mesh to w in glTF 2.0 JSON format with embedded buffers.
func ExportGLTF(w io.Writer, mesh *Mesh, opts GLTFExportOptions) error {
	if opts.AssetName == "" {
		opts.AssetName = "character"
	}

	gltf, err := buildGLTF(mesh, opts)
	if err != nil {
		return err
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(gltf)
}

// ExportGLTFDefault writes the mesh with default options.
func ExportGLTFDefault(w io.Writer, mesh *Mesh) error {
	return ExportGLTF(w, mesh, DefaultGLTFOptions())
}

// glTF structure types (JSON schema)
type gltfRoot struct {
	Asset       gltfAsset      `json:"asset"`
	Scene       int            `json:"scene"`
	Scenes      []gltfScene    `json:"scenes"`
	Nodes       []gltfNode     `json:"nodes"`
	Meshes      []gltfMesh     `json:"meshes"`
	Accessors   []gltfAccessor `json:"accessors"`
	BufferViews []gltfBufView  `json:"bufferViews"`
	Buffers     []gltfBuffer   `json:"buffers"`
	Materials   []gltfMaterial `json:"materials,omitempty"`
}

type gltfAsset struct {
	Version   string `json:"version"`
	Generator string `json:"generator"`
}

type gltfScene struct {
	Name  string `json:"name,omitempty"`
	Nodes []int  `json:"nodes"`
}

type gltfNode struct {
	Name string `json:"name,omitempty"`
	Mesh int    `json:"mesh"`
}

type gltfMesh struct {
	Name       string          `json:"name,omitempty"`
	Primitives []gltfPrimitive `json:"primitives"`
}

type gltfPrimitive struct {
	Attributes map[string]int `json:"attributes"`
	Indices    int            `json:"indices"`
	Mode       int            `json:"mode"` // 4 = TRIANGLES
	Material   *int           `json:"material,omitempty"`
}

type gltfAccessor struct {
	BufferView    int       `json:"bufferView"`
	ByteOffset    int       `json:"byteOffset,omitempty"`
	ComponentType int       `json:"componentType"` // 5123=UNSIGNED_SHORT, 5126=FLOAT, etc.
	Count         int       `json:"count"`
	Type          string    `json:"type"` // SCALAR, VEC2, VEC3, VEC4
	Min           []float64 `json:"min,omitempty"`
	Max           []float64 `json:"max,omitempty"`
}

type gltfBufView struct {
	Buffer     int `json:"buffer"`
	ByteLength int `json:"byteLength"`
	ByteOffset int `json:"byteOffset,omitempty"`
	ByteStride int `json:"byteStride,omitempty"`
	Target     int `json:"target,omitempty"` // 34962=ARRAY_BUFFER, 34963=ELEMENT_ARRAY_BUFFER
}

type gltfBuffer struct {
	ByteLength int    `json:"byteLength"`
	URI        string `json:"uri,omitempty"`
}

type gltfMaterial struct {
	Name             string                `json:"name,omitempty"`
	PBRMetallicRough *gltfPBRMetallicRough `json:"pbrMetallicRoughness,omitempty"`
}

type gltfPBRMetallicRough struct {
	BaseColorFactor []float64 `json:"baseColorFactor,omitempty"`
	MetallicFactor  float64   `json:"metallicFactor"`
	RoughnessFactor float64   `json:"roughnessFactor"`
}

func buildGLTF(mesh *Mesh, opts GLTFExportOptions) (*gltfRoot, error) {
	// Build binary buffer
	var buf []byte
	var bufferViews []gltfBufView
	var accessors []gltfAccessor
	attributes := make(map[string]int)

	// Track bounds for position
	var minPos, maxPos [3]float64
	minPos = [3]float64{math.MaxFloat64, math.MaxFloat64, math.MaxFloat64}
	maxPos = [3]float64{-math.MaxFloat64, -math.MaxFloat64, -math.MaxFloat64}

	// Position buffer
	posOffset := len(buf)
	for _, v := range mesh.Vertices {
		posData := make([]byte, 12)
		binary.LittleEndian.PutUint32(posData[0:4], math.Float32bits(v.Position[0]))
		binary.LittleEndian.PutUint32(posData[4:8], math.Float32bits(v.Position[1]))
		binary.LittleEndian.PutUint32(posData[8:12], math.Float32bits(v.Position[2]))
		buf = append(buf, posData...)

		// Update bounds
		for i := 0; i < 3; i++ {
			val := float64(v.Position[i])
			if val < minPos[i] {
				minPos[i] = val
			}
			if val > maxPos[i] {
				maxPos[i] = val
			}
		}
	}
	posLen := len(buf) - posOffset

	bufferViews = append(bufferViews, gltfBufView{
		Buffer:     0,
		ByteOffset: posOffset,
		ByteLength: posLen,
		Target:     34962, // ARRAY_BUFFER
	})
	accessors = append(accessors, gltfAccessor{
		BufferView:    len(bufferViews) - 1,
		ComponentType: 5126, // FLOAT
		Count:         len(mesh.Vertices),
		Type:          "VEC3",
		Min:           []float64{minPos[0], minPos[1], minPos[2]},
		Max:           []float64{maxPos[0], maxPos[1], maxPos[2]},
	})
	attributes["POSITION"] = len(accessors) - 1

	// Normal buffer
	if opts.IncludeNormals {
		normOffset := len(buf)
		for _, v := range mesh.Vertices {
			normData := make([]byte, 12)
			binary.LittleEndian.PutUint32(normData[0:4], math.Float32bits(v.Normal[0]))
			binary.LittleEndian.PutUint32(normData[4:8], math.Float32bits(v.Normal[1]))
			binary.LittleEndian.PutUint32(normData[8:12], math.Float32bits(v.Normal[2]))
			buf = append(buf, normData...)
		}
		normLen := len(buf) - normOffset

		bufferViews = append(bufferViews, gltfBufView{
			Buffer:     0,
			ByteOffset: normOffset,
			ByteLength: normLen,
			Target:     34962,
		})
		accessors = append(accessors, gltfAccessor{
			BufferView:    len(bufferViews) - 1,
			ComponentType: 5126,
			Count:         len(mesh.Vertices),
			Type:          "VEC3",
		})
		attributes["NORMAL"] = len(accessors) - 1
	}

	// UV buffer
	if opts.IncludeUVs {
		uvOffset := len(buf)
		for _, v := range mesh.Vertices {
			uvData := make([]byte, 8)
			binary.LittleEndian.PutUint32(uvData[0:4], math.Float32bits(v.UV0[0]))
			binary.LittleEndian.PutUint32(uvData[4:8], math.Float32bits(v.UV0[1]))
			buf = append(buf, uvData...)
		}
		uvLen := len(buf) - uvOffset

		bufferViews = append(bufferViews, gltfBufView{
			Buffer:     0,
			ByteOffset: uvOffset,
			ByteLength: uvLen,
			Target:     34962,
		})
		accessors = append(accessors, gltfAccessor{
			BufferView:    len(bufferViews) - 1,
			ComponentType: 5126,
			Count:         len(mesh.Vertices),
			Type:          "VEC2",
		})
		attributes["TEXCOORD_0"] = len(accessors) - 1
	}

	// Color buffer (as VEC4 normalized unsigned bytes)
	if opts.IncludeColors {
		colorOffset := len(buf)
		for _, v := range mesh.Vertices {
			// Convert float32 [0,1] to byte [0,255]
			r := byte(clampFloat32(v.Color[0], 0, 1) * 255)
			g := byte(clampFloat32(v.Color[1], 0, 1) * 255)
			b := byte(clampFloat32(v.Color[2], 0, 1) * 255)
			a := byte(clampFloat32(v.Color[3], 0, 1) * 255)
			colorData := []byte{r, g, b, a}
			buf = append(buf, colorData...)
		}
		colorLen := len(buf) - colorOffset

		bufferViews = append(bufferViews, gltfBufView{
			Buffer:     0,
			ByteOffset: colorOffset,
			ByteLength: colorLen,
			Target:     34962,
		})
		accessors = append(accessors, gltfAccessor{
			BufferView:    len(bufferViews) - 1,
			ComponentType: 5121, // UNSIGNED_BYTE
			Count:         len(mesh.Vertices),
			Type:          "VEC4",
		})
		attributes["COLOR_0"] = len(accessors) - 1
	}

	// Tangent buffer
	if opts.IncludeTangents {
		tangentOffset := len(buf)
		for _, v := range mesh.Vertices {
			tangentData := make([]byte, 16)
			binary.LittleEndian.PutUint32(tangentData[0:4], math.Float32bits(v.Tangent[0]))
			binary.LittleEndian.PutUint32(tangentData[4:8], math.Float32bits(v.Tangent[1]))
			binary.LittleEndian.PutUint32(tangentData[8:12], math.Float32bits(v.Tangent[2]))
			binary.LittleEndian.PutUint32(tangentData[12:16], math.Float32bits(v.Tangent[3]))
			buf = append(buf, tangentData...)
		}
		tangentLen := len(buf) - tangentOffset

		bufferViews = append(bufferViews, gltfBufView{
			Buffer:     0,
			ByteOffset: tangentOffset,
			ByteLength: tangentLen,
			Target:     34962,
		})
		accessors = append(accessors, gltfAccessor{
			BufferView:    len(bufferViews) - 1,
			ComponentType: 5126,
			Count:         len(mesh.Vertices),
			Type:          "VEC4",
		})
		attributes["TANGENT"] = len(accessors) - 1
	}

	// Skinning data (joints + weights)
	if opts.IncludeSkinning {
		// Joint indices as VEC4 unsigned shorts
		jointsOffset := len(buf)
		for _, v := range mesh.Vertices {
			jointData := make([]byte, 8)
			binary.LittleEndian.PutUint16(jointData[0:2], uint16(v.JointIds[0]))
			binary.LittleEndian.PutUint16(jointData[2:4], uint16(v.JointIds[1]))
			binary.LittleEndian.PutUint16(jointData[4:6], uint16(v.JointIds[2]))
			binary.LittleEndian.PutUint16(jointData[6:8], uint16(v.JointIds[3]))
			buf = append(buf, jointData...)
		}
		jointsLen := len(buf) - jointsOffset

		bufferViews = append(bufferViews, gltfBufView{
			Buffer:     0,
			ByteOffset: jointsOffset,
			ByteLength: jointsLen,
			Target:     34962,
		})
		accessors = append(accessors, gltfAccessor{
			BufferView:    len(bufferViews) - 1,
			ComponentType: 5123, // UNSIGNED_SHORT
			Count:         len(mesh.Vertices),
			Type:          "VEC4",
		})
		attributes["JOINTS_0"] = len(accessors) - 1

		// Joint weights as VEC4 floats
		weightsOffset := len(buf)
		for _, v := range mesh.Vertices {
			weightData := make([]byte, 16)
			binary.LittleEndian.PutUint32(weightData[0:4], math.Float32bits(v.JointWeights[0]))
			binary.LittleEndian.PutUint32(weightData[4:8], math.Float32bits(v.JointWeights[1]))
			binary.LittleEndian.PutUint32(weightData[8:12], math.Float32bits(v.JointWeights[2]))
			binary.LittleEndian.PutUint32(weightData[12:16], math.Float32bits(v.JointWeights[3]))
			buf = append(buf, weightData...)
		}
		weightsLen := len(buf) - weightsOffset

		bufferViews = append(bufferViews, gltfBufView{
			Buffer:     0,
			ByteOffset: weightsOffset,
			ByteLength: weightsLen,
			Target:     34962,
		})
		accessors = append(accessors, gltfAccessor{
			BufferView:    len(bufferViews) - 1,
			ComponentType: 5126,
			Count:         len(mesh.Vertices),
			Type:          "VEC4",
		})
		attributes["WEIGHTS_0"] = len(accessors) - 1
	}

	// Align buffer to 4 bytes before indices
	for len(buf)%4 != 0 {
		buf = append(buf, 0)
	}

	// Index buffer (as unsigned 32-bit integers)
	idxOffset := len(buf)
	for _, idx := range mesh.Indices {
		idxData := make([]byte, 4)
		binary.LittleEndian.PutUint32(idxData, idx)
		buf = append(buf, idxData...)
	}
	idxLen := len(buf) - idxOffset

	bufferViews = append(bufferViews, gltfBufView{
		Buffer:     0,
		ByteOffset: idxOffset,
		ByteLength: idxLen,
		Target:     34963, // ELEMENT_ARRAY_BUFFER
	})
	indicesAccessor := gltfAccessor{
		BufferView:    len(bufferViews) - 1,
		ComponentType: 5125, // UNSIGNED_INT
		Count:         len(mesh.Indices),
		Type:          "SCALAR",
	}
	accessors = append(accessors, indicesAccessor)
	indicesIdx := len(accessors) - 1

	// Build glTF structure
	gltf := &gltfRoot{
		Asset: gltfAsset{
			Version:   "2.0",
			Generator: "unpeople",
		},
		Scene: 0,
		Scenes: []gltfScene{{
			Name:  "Scene",
			Nodes: []int{0},
		}},
		Nodes: []gltfNode{{
			Name: opts.AssetName,
			Mesh: 0,
		}},
		Meshes: []gltfMesh{{
			Name: opts.AssetName,
			Primitives: []gltfPrimitive{{
				Attributes: attributes,
				Indices:    indicesIdx,
				Mode:       4, // TRIANGLES
			}},
		}},
		Accessors:   accessors,
		BufferViews: bufferViews,
		Buffers: []gltfBuffer{{
			ByteLength: len(buf),
		}},
	}

	// Add default material
	materialIdx := 0
	gltf.Materials = []gltfMaterial{{
		Name: "skin",
		PBRMetallicRough: &gltfPBRMetallicRough{
			BaseColorFactor: []float64{0.8, 0.6, 0.5, 1.0}, // Skin-like color
			MetallicFactor:  0.0,
			RoughnessFactor: 0.8,
		},
	}}
	gltf.Meshes[0].Primitives[0].Material = &materialIdx

	// Embed buffer as data URI
	if opts.EmbedBuffers {
		encoded := base64.StdEncoding.EncodeToString(buf)
		gltf.Buffers[0].URI = fmt.Sprintf("data:application/octet-stream;base64,%s", encoded)
	}

	return gltf, nil
}

// ExportGLB writes the mesh in glTF Binary format (.glb).
// GLB is more efficient than .gltf+.bin as everything is in one file.
func ExportGLB(w io.Writer, mesh *Mesh, opts GLTFExportOptions) error {
	// Build glTF with external buffer reference
	opts.EmbedBuffers = false
	gltf, err := buildGLTF(mesh, opts)
	if err != nil {
		return err
	}

	// Rebuild binary buffer
	var binBuf []byte
	for _, v := range mesh.Vertices {
		posData := make([]byte, 12)
		binary.LittleEndian.PutUint32(posData[0:4], math.Float32bits(v.Position[0]))
		binary.LittleEndian.PutUint32(posData[4:8], math.Float32bits(v.Position[1]))
		binary.LittleEndian.PutUint32(posData[8:12], math.Float32bits(v.Position[2]))
		binBuf = append(binBuf, posData...)
	}
	if opts.IncludeNormals {
		for _, v := range mesh.Vertices {
			normData := make([]byte, 12)
			binary.LittleEndian.PutUint32(normData[0:4], math.Float32bits(v.Normal[0]))
			binary.LittleEndian.PutUint32(normData[4:8], math.Float32bits(v.Normal[1]))
			binary.LittleEndian.PutUint32(normData[8:12], math.Float32bits(v.Normal[2]))
			binBuf = append(binBuf, normData...)
		}
	}
	if opts.IncludeUVs {
		for _, v := range mesh.Vertices {
			uvData := make([]byte, 8)
			binary.LittleEndian.PutUint32(uvData[0:4], math.Float32bits(v.UV0[0]))
			binary.LittleEndian.PutUint32(uvData[4:8], math.Float32bits(v.UV0[1]))
			binBuf = append(binBuf, uvData...)
		}
	}
	if opts.IncludeColors {
		for _, v := range mesh.Vertices {
			// Convert float32 [0,1] to byte [0,255]
			r := byte(clampFloat32(v.Color[0], 0, 1) * 255)
			g := byte(clampFloat32(v.Color[1], 0, 1) * 255)
			b := byte(clampFloat32(v.Color[2], 0, 1) * 255)
			a := byte(clampFloat32(v.Color[3], 0, 1) * 255)
			colorData := []byte{r, g, b, a}
			binBuf = append(binBuf, colorData...)
		}
	}
	if opts.IncludeTangents {
		for _, v := range mesh.Vertices {
			tangentData := make([]byte, 16)
			binary.LittleEndian.PutUint32(tangentData[0:4], math.Float32bits(v.Tangent[0]))
			binary.LittleEndian.PutUint32(tangentData[4:8], math.Float32bits(v.Tangent[1]))
			binary.LittleEndian.PutUint32(tangentData[8:12], math.Float32bits(v.Tangent[2]))
			binary.LittleEndian.PutUint32(tangentData[12:16], math.Float32bits(v.Tangent[3]))
			binBuf = append(binBuf, tangentData...)
		}
	}
	if opts.IncludeSkinning {
		for _, v := range mesh.Vertices {
			jointData := make([]byte, 8)
			binary.LittleEndian.PutUint16(jointData[0:2], uint16(v.JointIds[0]))
			binary.LittleEndian.PutUint16(jointData[2:4], uint16(v.JointIds[1]))
			binary.LittleEndian.PutUint16(jointData[4:6], uint16(v.JointIds[2]))
			binary.LittleEndian.PutUint16(jointData[6:8], uint16(v.JointIds[3]))
			binBuf = append(binBuf, jointData...)
		}
		for _, v := range mesh.Vertices {
			weightData := make([]byte, 16)
			binary.LittleEndian.PutUint32(weightData[0:4], math.Float32bits(v.JointWeights[0]))
			binary.LittleEndian.PutUint32(weightData[4:8], math.Float32bits(v.JointWeights[1]))
			binary.LittleEndian.PutUint32(weightData[8:12], math.Float32bits(v.JointWeights[2]))
			binary.LittleEndian.PutUint32(weightData[12:16], math.Float32bits(v.JointWeights[3]))
			binBuf = append(binBuf, weightData...)
		}
	}

	// Pad to 4 bytes
	for len(binBuf)%4 != 0 {
		binBuf = append(binBuf, 0)
	}

	// Indices
	for _, idx := range mesh.Indices {
		idxData := make([]byte, 4)
		binary.LittleEndian.PutUint32(idxData, idx)
		binBuf = append(binBuf, idxData...)
	}

	// JSON chunk
	jsonBytes, err := json.Marshal(gltf)
	if err != nil {
		return err
	}
	// Pad JSON to 4 bytes with spaces
	for len(jsonBytes)%4 != 0 {
		jsonBytes = append(jsonBytes, ' ')
	}

	// Calculate total size
	totalSize := 12 + // GLB header
		8 + len(jsonBytes) + // JSON chunk
		8 + len(binBuf) // BIN chunk

	// Write GLB header
	header := make([]byte, 12)
	copy(header[0:4], "glTF")                     // magic
	binary.LittleEndian.PutUint32(header[4:8], 2) // version
	binary.LittleEndian.PutUint32(header[8:12], uint32(totalSize))
	if _, err := w.Write(header); err != nil {
		return err
	}

	// Write JSON chunk
	jsonChunkHeader := make([]byte, 8)
	binary.LittleEndian.PutUint32(jsonChunkHeader[0:4], uint32(len(jsonBytes)))
	copy(jsonChunkHeader[4:8], "JSON")
	if _, err := w.Write(jsonChunkHeader); err != nil {
		return err
	}
	if _, err := w.Write(jsonBytes); err != nil {
		return err
	}

	// Write BIN chunk
	binChunkHeader := make([]byte, 8)
	binary.LittleEndian.PutUint32(binChunkHeader[0:4], uint32(len(binBuf)))
	copy(binChunkHeader[4:8], "BIN\x00")
	if _, err := w.Write(binChunkHeader); err != nil {
		return err
	}
	if _, err := w.Write(binBuf); err != nil {
		return err
	}

	return nil
}
