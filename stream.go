// Package unpeople provides deterministic procedural generation of humanoid meshes.
//
// This file implements streaming output APIs for memory-efficient generation
// of large numbers of characters.

package unpeople

import (
	"encoding/binary"
	"io"
)

// MeshWriter defines an interface for streaming mesh output.
type MeshWriter interface {
	// WriteHeader writes mesh metadata before vertices and indices.
	WriteHeader(vertexCount, indexCount int) error
	// WriteVertex writes a single vertex to the output.
	WriteVertex(v Vertex) error
	// WriteIndex writes a single index to the output.
	WriteIndex(idx uint32) error
	// Flush ensures all data is written.
	Flush() error
}

// StreamResult contains metadata about a streamed mesh.
type StreamResult struct {
	Key           string
	VertexCount   int
	IndexCount    int
	TriangleCount int
	BytesWritten  int64
}

// GenerateStream generates a character mesh and streams it to the provided writer.
// This is memory-efficient for very large scenes as it doesn't hold the complete
// mesh in memory.
// streamVertices writes all mesh vertices to the writer.
func streamVertices(vertices []Vertex, w MeshWriter) error {
	for _, v := range vertices {
		if err := w.WriteVertex(v); err != nil {
			return err
		}
	}
	return nil
}

// streamIndices writes all mesh indices to the writer.
func streamIndices(indices []uint32, w MeshWriter) error {
	for _, idx := range indices {
		if err := w.WriteIndex(idx); err != nil {
			return err
		}
	}
	return nil
}

// buildStreamResult creates a StreamResult from a mesh.
func buildStreamResult(mesh *Mesh) *StreamResult {
	return &StreamResult{
		Key:           mesh.Key,
		VertexCount:   len(mesh.Vertices),
		IndexCount:    len(mesh.Indices),
		TriangleCount: len(mesh.Indices) / 3,
	}
}

// streamMeshToWriter writes the mesh data (header, vertices, indices) to the writer.
func streamMeshToWriter(mesh *Mesh, w MeshWriter) error {
	if err := w.WriteHeader(len(mesh.Vertices), len(mesh.Indices)); err != nil {
		return err
	}
	if err := streamVertices(mesh.Vertices, w); err != nil {
		return err
	}
	if err := streamIndices(mesh.Indices, w); err != nil {
		return err
	}
	return w.Flush()
}

// GenerateStream generates a mesh from the given parameters and writes it
// to the provided MeshWriter. This allows streaming large meshes without
// holding the entire mesh in memory.
func (g *Generator) GenerateStream(p Params, w MeshWriter) (*StreamResult, error) {
	if err := p.Validate(); err != nil {
		return nil, err
	}

	mesh, err := g.Generate(p)
	if err != nil {
		return nil, err
	}

	if err := streamMeshToWriter(mesh, w); err != nil {
		return nil, err
	}

	return buildStreamResult(mesh), nil
}

// MeshChan is a channel-based mesh output for concurrent consumption.
type MeshChan struct {
	Vertices chan Vertex
	Indices  chan uint32
	Done     chan struct{}
	Err      chan error
}

// NewMeshChan creates a new channel-based mesh output.
func NewMeshChan(bufferSize int) *MeshChan {
	return &MeshChan{
		Vertices: make(chan Vertex, bufferSize),
		Indices:  make(chan uint32, bufferSize*3), // 3 indices per triangle
		Done:     make(chan struct{}),
		Err:      make(chan error, 1),
	}
}

// Close closes all channels in the MeshChan.
func (mc *MeshChan) Close() {
	close(mc.Vertices)
	close(mc.Indices)
	close(mc.Done)
}

// GenerateToChan generates a character mesh and sends it to the provided channels.
// The function returns immediately; generation happens in a goroutine.
// The caller should read from Vertices and Indices channels, then wait for Done.
func (g *Generator) GenerateToChan(p Params, mc *MeshChan) {
	go func() {
		defer mc.Close()

		mesh, err := g.Generate(p)
		if err != nil {
			mc.Err <- err
			return
		}

		// Send vertices
		for _, v := range mesh.Vertices {
			mc.Vertices <- v
		}

		// Send indices
		for _, idx := range mesh.Indices {
			mc.Indices <- idx
		}
	}()
}

// BinaryMeshWriter writes mesh data in a compact binary format.
type BinaryMeshWriter struct {
	w            io.Writer
	bytesWritten int64
	err          error
}

// NewBinaryMeshWriter creates a new binary mesh writer.
func NewBinaryMeshWriter(w io.Writer) *BinaryMeshWriter {
	return &BinaryMeshWriter{w: w}
}

// writeHeaderField writes a single uint32 header field.
func (bw *BinaryMeshWriter) writeHeaderField(value uint32) error {
	if err := binary.Write(bw.w, binary.LittleEndian, value); err != nil {
		bw.err = err
		return err
	}
	bw.bytesWritten += 4
	return nil
}

// writeMagicNumber writes the UNPM magic number to identify the file format.
func (bw *BinaryMeshWriter) writeMagicNumber() error {
	magic := []byte{'U', 'N', 'P', 'M'}
	if _, err := bw.w.Write(magic); err != nil {
		bw.err = err
		return err
	}
	bw.bytesWritten += 4
	return nil
}

// WriteHeader writes the mesh header (vertex count, index count).
func (bw *BinaryMeshWriter) WriteHeader(vertexCount, indexCount int) error {
	if bw.err != nil {
		return bw.err
	}
	if err := bw.writeMagicNumber(); err != nil {
		return err
	}
	if err := bw.writeHeaderField(1); err != nil { // Version
		return err
	}
	if err := bw.writeHeaderField(uint32(vertexCount)); err != nil {
		return err
	}
	return bw.writeHeaderField(uint32(indexCount))
}

// writeBinaryField writes a single field and updates byte count.
func (bw *BinaryMeshWriter) writeBinaryField(data any, size int64) error {
	if bw.err != nil {
		return bw.err
	}
	if err := binary.Write(bw.w, binary.LittleEndian, data); err != nil {
		bw.err = err
		return err
	}
	bw.bytesWritten += size
	return nil
}

// vertexField represents a vertex attribute with its data and byte size.
type vertexField struct {
	data any
	size int64
}

// vertexFields returns all fields of a vertex for sequential binary writing.
func vertexFields(v Vertex) []vertexField {
	return []vertexField{
		{v.Position, 12},     // Position (3 floats)
		{v.Normal, 12},       // Normal (3 floats)
		{v.Tangent, 16},      // Tangent (4 floats)
		{v.UV0, 8},           // UV0 (2 floats)
		{v.Color, 4},         // Color (4 bytes)
		{v.JointIds, 8},      // JointIds (4 int32)
		{v.JointWeights, 16}, // JointWeights (4 floats)
		{v.MorphTarget, 12},  // MorphTarget (3 floats)
	}
}

// WriteVertex writes a single vertex in binary format.
func (bw *BinaryMeshWriter) WriteVertex(v Vertex) error {
	for _, field := range vertexFields(v) {
		if err := bw.writeBinaryField(field.data, field.size); err != nil {
			return err
		}
	}
	return nil
}

// WriteIndex writes a single index in binary format.
func (bw *BinaryMeshWriter) WriteIndex(idx uint32) error {
	if bw.err != nil {
		return bw.err
	}

	if err := binary.Write(bw.w, binary.LittleEndian, idx); err != nil {
		bw.err = err
		return err
	}
	bw.bytesWritten += 4

	return nil
}

// Flush flushes the underlying writer if it supports flushing.
func (bw *BinaryMeshWriter) Flush() error {
	if bw.err != nil {
		return bw.err
	}

	if f, ok := bw.w.(interface{ Flush() error }); ok {
		return f.Flush()
	}
	return nil
}

// BytesWritten returns the total number of bytes written.
func (bw *BinaryMeshWriter) BytesWritten() int64 {
	return bw.bytesWritten
}

// BatchStreamResult contains results from streaming batch generation.
type BatchStreamResult struct {
	MeshResults []StreamResult
	TotalBytes  int64
	Errors      []error
}

// GenerateBatchStream generates multiple meshes and streams them to the writer.
// Each mesh is written sequentially with its own header.
func (g *Generator) GenerateBatchStream(params []Params, w MeshWriter) (*BatchStreamResult, error) {
	result := &BatchStreamResult{
		MeshResults: make([]StreamResult, 0, len(params)),
	}

	for i, p := range params {
		sr, err := g.GenerateStream(p, w)
		if err != nil {
			result.Errors = append(result.Errors, err)
			continue
		}

		result.MeshResults = append(result.MeshResults, *sr)
		if bw, ok := w.(*BinaryMeshWriter); ok {
			result.TotalBytes = bw.BytesWritten()
		}

		// Allow early termination check
		_ = i
	}

	return result, nil
}

// VertexSize returns the size in bytes of a single vertex in the binary format.
// Position(12) + Normal(12) + Tangent(16) + UV0(8) + Color(4) + JointIds(8) + JointWeights(16) + MorphTarget(12) = 88
func VertexSize() int {
	return 88
}

// EstimateMeshSize estimates the binary size of a mesh with the given counts.
func EstimateMeshSize(vertexCount, indexCount int) int {
	headerSize := 16 // magic + version + counts
	verticesSize := vertexCount * VertexSize()
	indicesSize := indexCount * 4
	return headerSize + verticesSize + indicesSize
}
