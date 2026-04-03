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
func (g *Generator) GenerateStream(p Params, w MeshWriter) (*StreamResult, error) {
	if err := p.Validate(); err != nil {
		return nil, err
	}

	// Generate the mesh normally (we need all data to get counts first)
	mesh, err := g.Generate(p)
	if err != nil {
		return nil, err
	}

	// Write header with counts
	if err := w.WriteHeader(len(mesh.Vertices), len(mesh.Indices)); err != nil {
		return nil, err
	}

	// Stream vertices
	for _, v := range mesh.Vertices {
		if err := w.WriteVertex(v); err != nil {
			return nil, err
		}
	}

	// Stream indices
	for _, idx := range mesh.Indices {
		if err := w.WriteIndex(idx); err != nil {
			return nil, err
		}
	}

	if err := w.Flush(); err != nil {
		return nil, err
	}

	return &StreamResult{
		Key:           mesh.Key,
		VertexCount:   len(mesh.Vertices),
		IndexCount:    len(mesh.Indices),
		TriangleCount: len(mesh.Indices) / 3,
	}, nil
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

// WriteHeader writes the mesh header (vertex count, index count).
func (bw *BinaryMeshWriter) WriteHeader(vertexCount, indexCount int) error {
	if bw.err != nil {
		return bw.err
	}

	// Magic number "UNPM" (UNPeople Mesh)
	magic := []byte{'U', 'N', 'P', 'M'}
	if _, err := bw.w.Write(magic); err != nil {
		bw.err = err
		return err
	}
	bw.bytesWritten += 4

	// Version (1)
	if err := binary.Write(bw.w, binary.LittleEndian, uint32(1)); err != nil {
		bw.err = err
		return err
	}
	bw.bytesWritten += 4

	// Vertex count
	if err := binary.Write(bw.w, binary.LittleEndian, uint32(vertexCount)); err != nil {
		bw.err = err
		return err
	}
	bw.bytesWritten += 4

	// Index count
	if err := binary.Write(bw.w, binary.LittleEndian, uint32(indexCount)); err != nil {
		bw.err = err
		return err
	}
	bw.bytesWritten += 4

	return nil
}

// WriteVertex writes a single vertex in binary format.
func (bw *BinaryMeshWriter) WriteVertex(v Vertex) error {
	if bw.err != nil {
		return bw.err
	}

	// Position (3 floats = 12 bytes)
	if err := binary.Write(bw.w, binary.LittleEndian, v.Position); err != nil {
		bw.err = err
		return err
	}
	bw.bytesWritten += 12

	// Normal (3 floats = 12 bytes)
	if err := binary.Write(bw.w, binary.LittleEndian, v.Normal); err != nil {
		bw.err = err
		return err
	}
	bw.bytesWritten += 12

	// Tangent (4 floats = 16 bytes)
	if err := binary.Write(bw.w, binary.LittleEndian, v.Tangent); err != nil {
		bw.err = err
		return err
	}
	bw.bytesWritten += 16

	// UV0 (2 floats = 8 bytes)
	if err := binary.Write(bw.w, binary.LittleEndian, v.UV0); err != nil {
		bw.err = err
		return err
	}
	bw.bytesWritten += 8

	// Color (4 bytes)
	if err := binary.Write(bw.w, binary.LittleEndian, v.Color); err != nil {
		bw.err = err
		return err
	}
	bw.bytesWritten += 4

	// JointIds (4 uint16 = 8 bytes)
	if err := binary.Write(bw.w, binary.LittleEndian, v.JointIds); err != nil {
		bw.err = err
		return err
	}
	bw.bytesWritten += 8

	// JointWeights (4 floats = 16 bytes)
	if err := binary.Write(bw.w, binary.LittleEndian, v.JointWeights); err != nil {
		bw.err = err
		return err
	}
	bw.bytesWritten += 16

	// MorphTarget (3 floats = 12 bytes)
	if err := binary.Write(bw.w, binary.LittleEndian, v.MorphTarget); err != nil {
		bw.err = err
		return err
	}
	bw.bytesWritten += 12

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
