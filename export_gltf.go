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
	// IncludeSlots includes attachment slot nodes (default: false)
	IncludeSlots bool
	// AssetName sets the mesh name in the glTF asset
	AssetName string
	// Slots provides attachment slot data when IncludeSlots is true
	Slots *AttachmentSlots
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
	Name        string      `json:"name,omitempty"`
	Mesh        *int        `json:"mesh,omitempty"`
	Translation *[3]float32 `json:"translation,omitempty"`
	Rotation    *[4]float32 `json:"rotation,omitempty"`
	Scale       *[3]float32 `json:"scale,omitempty"`
	Children    []int       `json:"children,omitempty"`
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

// gltfBuildResult holds the output of buffer building for glTF export.
type gltfBuildResult struct {
	buffer      []byte
	bufferViews []gltfBufView
	accessors   []gltfAccessor
	attributes  map[string]int
	indicesIdx  int
	minPos      [3]float64
	maxPos      [3]float64
}

// buildGLTFBuffers creates the binary buffer and all accessors/views for a mesh.
func buildGLTFBuffers(mesh *Mesh, opts GLTFExportOptions) *gltfBuildResult {
	result := &gltfBuildResult{
		attributes: make(map[string]int),
		minPos:     [3]float64{math.MaxFloat64, math.MaxFloat64, math.MaxFloat64},
		maxPos:     [3]float64{-math.MaxFloat64, -math.MaxFloat64, -math.MaxFloat64},
	}

	// Position buffer with bounds tracking
	result.appendPositions(mesh)

	if opts.IncludeNormals {
		result.appendNormals(mesh)
	}
	if opts.IncludeUVs {
		result.appendUVs(mesh)
	}
	if opts.IncludeColors {
		result.appendColors(mesh)
	}
	if opts.IncludeTangents {
		result.appendTangents(mesh)
	}
	if opts.IncludeSkinning {
		result.appendSkinning(mesh)
	}

	// Align and append indices
	result.alignTo4Bytes()
	result.appendIndices(mesh)

	return result
}

// updateBounds updates the min/max position bounds for a vertex position.
func (r *gltfBuildResult) updateBounds(pos [3]float32) {
	for i := 0; i < 3; i++ {
		val := float64(pos[i])
		if val < r.minPos[i] {
			r.minPos[i] = val
		}
		if val > r.maxPos[i] {
			r.maxPos[i] = val
		}
	}
}

// addPositionAccessor adds the accessor for position data with bounds.
func (r *gltfBuildResult) addPositionAccessor(vertexCount int) {
	r.accessors = append(r.accessors, gltfAccessor{
		BufferView:    len(r.bufferViews) - 1,
		ComponentType: 5126,
		Count:         vertexCount,
		Type:          "VEC3",
		Min:           []float64{r.minPos[0], r.minPos[1], r.minPos[2]},
		Max:           []float64{r.maxPos[0], r.maxPos[1], r.maxPos[2]},
	})
	r.attributes["POSITION"] = len(r.accessors) - 1
}

func (r *gltfBuildResult) appendPositions(mesh *Mesh) {
	offset := len(r.buffer)
	for _, v := range mesh.Vertices {
		r.buffer = appendVec3(r.buffer, v.Position)
		r.updateBounds(v.Position)
	}
	r.addBufferView(offset, len(r.buffer)-offset, 34962)
	r.addPositionAccessor(len(mesh.Vertices))
}

func (r *gltfBuildResult) appendNormals(mesh *Mesh) {
	offset := len(r.buffer)
	for _, v := range mesh.Vertices {
		r.buffer = appendVec3(r.buffer, v.Normal)
	}
	r.addBufferView(offset, len(r.buffer)-offset, 34962)
	r.addAccessorVec3(len(mesh.Vertices))
	r.attributes["NORMAL"] = len(r.accessors) - 1
}

func (r *gltfBuildResult) appendUVs(mesh *Mesh) {
	offset := len(r.buffer)
	for _, v := range mesh.Vertices {
		r.buffer = appendVec2(r.buffer, v.UV0)
	}
	r.addBufferView(offset, len(r.buffer)-offset, 34962)
	r.addAccessorVec2(len(mesh.Vertices))
	r.attributes["TEXCOORD_0"] = len(r.accessors) - 1
}

func (r *gltfBuildResult) appendColors(mesh *Mesh) {
	offset := len(r.buffer)
	for _, v := range mesh.Vertices {
		r.buffer = append(r.buffer,
			byte(clampFloat32(v.Color[0], 0, 1)*255),
			byte(clampFloat32(v.Color[1], 0, 1)*255),
			byte(clampFloat32(v.Color[2], 0, 1)*255),
			byte(clampFloat32(v.Color[3], 0, 1)*255),
		)
	}
	r.addBufferView(offset, len(r.buffer)-offset, 34962)
	r.accessors = append(r.accessors, gltfAccessor{
		BufferView:    len(r.bufferViews) - 1,
		ComponentType: 5121, // UNSIGNED_BYTE
		Count:         len(mesh.Vertices),
		Type:          "VEC4",
	})
	r.attributes["COLOR_0"] = len(r.accessors) - 1
}

func (r *gltfBuildResult) appendTangents(mesh *Mesh) {
	offset := len(r.buffer)
	for _, v := range mesh.Vertices {
		r.buffer = appendVec4(r.buffer, v.Tangent)
	}
	r.addBufferView(offset, len(r.buffer)-offset, 34962)
	r.addAccessorVec4(len(mesh.Vertices))
	r.attributes["TANGENT"] = len(r.accessors) - 1
}

func (r *gltfBuildResult) appendSkinning(mesh *Mesh) {
	// Joint indices
	jointsOffset := len(r.buffer)
	for _, v := range mesh.Vertices {
		jointData := make([]byte, 8)
		binary.LittleEndian.PutUint16(jointData[0:2], uint16(v.JointIds[0]))
		binary.LittleEndian.PutUint16(jointData[2:4], uint16(v.JointIds[1]))
		binary.LittleEndian.PutUint16(jointData[4:6], uint16(v.JointIds[2]))
		binary.LittleEndian.PutUint16(jointData[6:8], uint16(v.JointIds[3]))
		r.buffer = append(r.buffer, jointData...)
	}
	r.addBufferView(jointsOffset, len(r.buffer)-jointsOffset, 34962)
	r.accessors = append(r.accessors, gltfAccessor{
		BufferView:    len(r.bufferViews) - 1,
		ComponentType: 5123, // UNSIGNED_SHORT
		Count:         len(mesh.Vertices),
		Type:          "VEC4",
	})
	r.attributes["JOINTS_0"] = len(r.accessors) - 1

	// Joint weights
	weightsOffset := len(r.buffer)
	for _, v := range mesh.Vertices {
		r.buffer = appendVec4(r.buffer, v.JointWeights)
	}
	r.addBufferView(weightsOffset, len(r.buffer)-weightsOffset, 34962)
	r.addAccessorVec4(len(mesh.Vertices))
	r.attributes["WEIGHTS_0"] = len(r.accessors) - 1
}

func (r *gltfBuildResult) appendIndices(mesh *Mesh) {
	offset := len(r.buffer)
	for _, idx := range mesh.Indices {
		idxData := make([]byte, 4)
		binary.LittleEndian.PutUint32(idxData, idx)
		r.buffer = append(r.buffer, idxData...)
	}
	r.addBufferView(offset, len(r.buffer)-offset, 34963)
	r.accessors = append(r.accessors, gltfAccessor{
		BufferView:    len(r.bufferViews) - 1,
		ComponentType: 5125, // UNSIGNED_INT
		Count:         len(mesh.Indices),
		Type:          "SCALAR",
	})
	r.indicesIdx = len(r.accessors) - 1
}

func (r *gltfBuildResult) addBufferView(offset, length, target int) {
	r.bufferViews = append(r.bufferViews, gltfBufView{
		Buffer:     0,
		ByteOffset: offset,
		ByteLength: length,
		Target:     target,
	})
}

func (r *gltfBuildResult) addAccessorVec2(count int) {
	r.accessors = append(r.accessors, gltfAccessor{
		BufferView:    len(r.bufferViews) - 1,
		ComponentType: 5126,
		Count:         count,
		Type:          "VEC2",
	})
}

func (r *gltfBuildResult) addAccessorVec3(count int) {
	r.accessors = append(r.accessors, gltfAccessor{
		BufferView:    len(r.bufferViews) - 1,
		ComponentType: 5126,
		Count:         count,
		Type:          "VEC3",
	})
}

func (r *gltfBuildResult) addAccessorVec4(count int) {
	r.accessors = append(r.accessors, gltfAccessor{
		BufferView:    len(r.bufferViews) - 1,
		ComponentType: 5126,
		Count:         count,
		Type:          "VEC4",
	})
}

func (r *gltfBuildResult) alignTo4Bytes() {
	for len(r.buffer)%4 != 0 {
		r.buffer = append(r.buffer, 0)
	}
}

// appendVec2 appends a 2-component float vector to the buffer.
func appendVec2(buf []byte, v [2]float32) []byte {
	data := make([]byte, 8)
	binary.LittleEndian.PutUint32(data[0:4], math.Float32bits(v[0]))
	binary.LittleEndian.PutUint32(data[4:8], math.Float32bits(v[1]))
	return append(buf, data...)
}

// appendVec3 appends a 3-component float vector to the buffer.
func appendVec3(buf []byte, v [3]float32) []byte {
	data := make([]byte, 12)
	binary.LittleEndian.PutUint32(data[0:4], math.Float32bits(v[0]))
	binary.LittleEndian.PutUint32(data[4:8], math.Float32bits(v[1]))
	binary.LittleEndian.PutUint32(data[8:12], math.Float32bits(v[2]))
	return append(buf, data...)
}

// appendVec4 appends a 4-component float vector to the buffer.
func appendVec4(buf []byte, v [4]float32) []byte {
	data := make([]byte, 16)
	binary.LittleEndian.PutUint32(data[0:4], math.Float32bits(v[0]))
	binary.LittleEndian.PutUint32(data[4:8], math.Float32bits(v[1]))
	binary.LittleEndian.PutUint32(data[8:12], math.Float32bits(v[2]))
	binary.LittleEndian.PutUint32(data[12:16], math.Float32bits(v[3]))
	return append(buf, data...)
}

// buildGLTFBase creates the glTF root structure with asset, scene, and mesh data.
func buildGLTFBase(br *gltfBuildResult, opts GLTFExportOptions) *gltfRoot {
	meshIdx := 0
	rootNode := gltfNode{
		Name: opts.AssetName,
		Mesh: &meshIdx,
	}

	nodes := []gltfNode{rootNode}
	sceneNodes := []int{0}

	// Add attachment slot nodes if requested
	if opts.IncludeSlots && opts.Slots != nil {
		for i, slot := range opts.Slots.Slots {
			pos := [3]float32{slot.Position[0], slot.Position[1], slot.Position[2]}
			rot := [4]float32{slot.Rotation[0], slot.Rotation[1], slot.Rotation[2], slot.Rotation[3]}
			scale := [3]float32{slot.Scale[0], slot.Scale[1], slot.Scale[2]}

			slotNode := gltfNode{
				Name:        "Slot_" + slot.Name,
				Translation: &pos,
				Rotation:    &rot,
				Scale:       &scale,
			}
			nodes = append(nodes, slotNode)
			sceneNodes = append(sceneNodes, i+1)
		}
	}

	return &gltfRoot{
		Asset: gltfAsset{
			Version:   "2.0",
			Generator: "unpeople",
		},
		Scene: 0,
		Scenes: []gltfScene{{
			Name:  "Scene",
			Nodes: sceneNodes,
		}},
		Nodes: nodes,
		Meshes: []gltfMesh{{
			Name: opts.AssetName,
			Primitives: []gltfPrimitive{{
				Attributes: br.attributes,
				Indices:    br.indicesIdx,
				Mode:       4, // TRIANGLES
			}},
		}},
		Accessors:   br.accessors,
		BufferViews: br.bufferViews,
		Buffers: []gltfBuffer{{
			ByteLength: len(br.buffer),
		}},
	}
}

// buildDefaultSkinMaterial creates the default PBR skin material.
func buildDefaultSkinMaterial() gltfMaterial {
	return gltfMaterial{
		Name: "skin",
		PBRMetallicRough: &gltfPBRMetallicRough{
			BaseColorFactor: []float64{0.8, 0.6, 0.5, 1.0},
			MetallicFactor:  0.0,
			RoughnessFactor: 0.8,
		},
	}
}

func buildGLTF(mesh *Mesh, opts GLTFExportOptions) (*gltfRoot, error) {
	br := buildGLTFBuffers(mesh, opts)
	gltf := buildGLTFBase(br, opts)

	// Add default material
	materialIdx := 0
	gltf.Materials = []gltfMaterial{buildDefaultSkinMaterial()}
	gltf.Meshes[0].Primitives[0].Material = &materialIdx

	if opts.EmbedBuffers {
		encoded := base64.StdEncoding.EncodeToString(br.buffer)
		gltf.Buffers[0].URI = fmt.Sprintf("data:application/octet-stream;base64,%s", encoded)
	}

	return gltf, nil
}

// ExportGLB writes the mesh in glTF Binary format (.glb).
// GLB is more efficient than .gltf+.bin as everything is in one file.
func ExportGLB(w io.Writer, mesh *Mesh, opts GLTFExportOptions) error {
	opts.EmbedBuffers = false
	gltf, err := buildGLTF(mesh, opts)
	if err != nil {
		return err
	}

	// Get the binary buffer from a fresh build
	br := buildGLTFBuffers(mesh, opts)

	return writeGLBBinary(w, gltf, br.buffer)
}

// writeGLBBinary writes the GLB file format with JSON and binary chunks.
func writeGLBBinary(w io.Writer, gltf *gltfRoot, binBuf []byte) error {
	jsonBytes, err := json.Marshal(gltf)
	if err != nil {
		return err
	}
	// Pad JSON to 4 bytes with spaces
	for len(jsonBytes)%4 != 0 {
		jsonBytes = append(jsonBytes, ' ')
	}

	totalSize := 12 + 8 + len(jsonBytes) + 8 + len(binBuf)

	if err := writeGLBHeader(w, totalSize); err != nil {
		return err
	}
	if err := writeGLBChunk(w, "JSON", jsonBytes); err != nil {
		return err
	}
	return writeGLBChunk(w, "BIN\x00", binBuf)
}

// writeGLBHeader writes the 12-byte GLB header.
func writeGLBHeader(w io.Writer, totalSize int) error {
	header := make([]byte, 12)
	copy(header[0:4], "glTF")
	binary.LittleEndian.PutUint32(header[4:8], 2)
	binary.LittleEndian.PutUint32(header[8:12], uint32(totalSize))
	_, err := w.Write(header)
	return err
}

// writeGLBChunk writes a single GLB chunk with the given type and data.
func writeGLBChunk(w io.Writer, chunkType string, data []byte) error {
	chunkHeader := make([]byte, 8)
	binary.LittleEndian.PutUint32(chunkHeader[0:4], uint32(len(data)))
	copy(chunkHeader[4:8], chunkType)
	if _, err := w.Write(chunkHeader); err != nil {
		return err
	}
	_, err := w.Write(data)
	return err
}
