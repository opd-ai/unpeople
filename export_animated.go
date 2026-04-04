// Package unpeople – animated glTF export with skeleton and BVH animation
//
// This file extends the glTF export functionality to include skeleton nodes,
// skin data, and animation channels from BVH motion capture data.
package unpeople

import (
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"math"
)

// AnimatedMeshResult contains a mesh with its skeleton and optional animation.
type AnimatedMeshResult struct {
	Mesh      *Mesh
	Skeleton  *Skeleton
	Animation *Animation // nil if no animation
}

// GenerateAnimated produces a humanoid mesh with skeleton and animation from BVH data.
// If bvhData is nil, returns just the mesh with skeleton in bind pose.
func (g *Generator) GenerateAnimated(p Params, bvhData *BVHFile) (*AnimatedMeshResult, error) {
	if err := p.Validate(); err != nil {
		return nil, fmt.Errorf("unpeople: invalid params: %w", err)
	}

	// Generate the base mesh with skinning enabled
	originalMerge := p.MergeVertices
	p.MergeVertices = false // Disable for skinning
	mesh, err := g.Generate(p)
	p.MergeVertices = originalMerge
	if err != nil {
		return nil, err
	}

	// Generate skeleton matching the mesh layout
	rng := newSplitmix64(p.Seed)
	layout := computeBodyLayout(&p, rng)
	skeleton := GenerateSkeleton(layout)

	// Compute skinning weights for the mesh
	ComputeSkinningWeights(mesh, skeleton, DefaultSkinningParams())

	result := &AnimatedMeshResult{
		Mesh:     mesh,
		Skeleton: skeleton,
	}

	// Convert BVH to animation if provided
	if bvhData != nil {
		result.Animation = BVHToAnimation(bvhData, "animation")
	}

	return result, nil
}

// ─── Animated glTF Export ────────────────────────────────────────────────────

// gltfAnimatedRoot extends gltfRoot with animation and skin data.
type gltfAnimatedRoot struct {
	Asset       gltfAsset       `json:"asset"`
	Scene       int             `json:"scene"`
	Scenes      []gltfScene     `json:"scenes"`
	Nodes       []gltfAnimNode  `json:"nodes"`
	Meshes      []gltfMesh      `json:"meshes"`
	Accessors   []gltfAccessor  `json:"accessors"`
	BufferViews []gltfBufView   `json:"bufferViews"`
	Buffers     []gltfBuffer    `json:"buffers"`
	Materials   []gltfMaterial  `json:"materials,omitempty"`
	Skins       []gltfSkin      `json:"skins,omitempty"`
	Animations  []gltfAnimation `json:"animations,omitempty"`
}

// gltfAnimNode extends gltfNode for skeleton hierarchy.
type gltfAnimNode struct {
	Name        string      `json:"name,omitempty"`
	Mesh        *int        `json:"mesh,omitempty"`
	Skin        *int        `json:"skin,omitempty"`
	Children    []int       `json:"children,omitempty"`
	Translation *[3]float32 `json:"translation,omitempty"`
	Rotation    *[4]float32 `json:"rotation,omitempty"`
	Scale       *[3]float32 `json:"scale,omitempty"`
}

// gltfSkin represents a skeleton for mesh skinning.
type gltfSkin struct {
	Name                string `json:"name,omitempty"`
	Joints              []int  `json:"joints"`
	InverseBindMatrices int    `json:"inverseBindMatrices"`
	Skeleton            *int   `json:"skeleton,omitempty"`
}

// gltfAnimation represents a single animation clip.
type gltfAnimation struct {
	Name     string        `json:"name,omitempty"`
	Samplers []gltfSampler `json:"samplers"`
	Channels []gltfChannel `json:"channels"`
}

// gltfSampler defines how to interpolate animation data.
type gltfSampler struct {
	Input         int    `json:"input"`         // Accessor for keyframe times
	Output        int    `json:"output"`        // Accessor for keyframe values
	Interpolation string `json:"interpolation"` // LINEAR, STEP, CUBICSPLINE
}

// gltfChannel connects a sampler to a node property.
type gltfChannel struct {
	Sampler int               `json:"sampler"`
	Target  gltfChannelTarget `json:"target"`
}

// gltfChannelTarget specifies which node/property to animate.
type gltfChannelTarget struct {
	Node int    `json:"node"`
	Path string `json:"path"` // translation, rotation, scale
}

// AnimatedGLTFOptions configures animated glTF export.
type AnimatedGLTFOptions struct {
	GLTFExportOptions
	IncludeAnimation bool
}

// DefaultAnimatedGLTFOptions returns sensible default export options.
func DefaultAnimatedGLTFOptions() AnimatedGLTFOptions {
	return AnimatedGLTFOptions{
		GLTFExportOptions: GLTFExportOptions{
			EmbedBuffers:    true,
			IncludeNormals:  true,
			IncludeUVs:      true,
			IncludeColors:   true,
			IncludeTangents: false,
			IncludeSkinning: true,
			AssetName:       "animated_character",
		},
		IncludeAnimation: true,
	}
}

// ExportAnimatedGLTF writes an animated mesh to w in glTF 2.0 JSON format.
func ExportAnimatedGLTF(w io.Writer, result *AnimatedMeshResult, opts AnimatedGLTFOptions) error {
	if opts.AssetName == "" {
		opts.AssetName = "animated_character"
	}

	gltf, err := buildAnimatedGLTF(result, opts)
	if err != nil {
		return err
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(gltf)
}

// ExportAnimatedGLB writes an animated mesh in glTF Binary format (.glb).
func ExportAnimatedGLB(w io.Writer, result *AnimatedMeshResult, opts AnimatedGLTFOptions) error {
	opts.EmbedBuffers = false
	gltf, buffer, err := buildAnimatedGLTFWithBuffer(result, opts)
	if err != nil {
		return err
	}

	return writeAnimatedGLBBinary(w, gltf, buffer)
}

// ─── Build Functions ─────────────────────────────────────────────────────────

// animatedBuildResult extends gltfBuildResult with animation data.
type animatedBuildResult struct {
	*gltfBuildResult
	skinIBMAccessor  int
	animTimeAccessor int
	jointAccessors   map[JointID]jointAccessors
}

// jointAccessors holds accessor indices for one joint's animation data.
type jointAccessors struct {
	translationAccessor int
	rotationAccessor    int
}

func buildAnimatedGLTF(result *AnimatedMeshResult, opts AnimatedGLTFOptions) (*gltfAnimatedRoot, error) {
	gltf, _, err := buildAnimatedGLTFWithBuffer(result, opts)
	return gltf, err
}

func buildAnimatedGLTFWithBuffer(result *AnimatedMeshResult, opts AnimatedGLTFOptions) (*gltfAnimatedRoot, []byte, error) {
	opts.IncludeSkinning = true // Always include skinning for animated export

	// Build mesh data buffer
	meshBr := buildGLTFBuffers(result.Mesh, opts.GLTFExportOptions)

	// Build skeleton and animation data
	abr := &animatedBuildResult{
		gltfBuildResult: meshBr,
		jointAccessors:  make(map[JointID]jointAccessors),
	}

	// Add inverse bind matrices
	abr.appendInverseBindMatrices(result.Skeleton)

	// Add animation data if present
	if opts.IncludeAnimation && result.Animation != nil {
		abr.appendAnimationData(result.Animation)
	}

	// Build glTF structure
	gltf := buildAnimatedGLTFStructure(abr, result, opts)

	if opts.EmbedBuffers {
		encoded := base64.StdEncoding.EncodeToString(abr.buffer)
		gltf.Buffers[0].URI = fmt.Sprintf("data:application/octet-stream;base64,%s", encoded)
	}

	return gltf, abr.buffer, nil
}

func (r *animatedBuildResult) appendInverseBindMatrices(skel *Skeleton) {
	r.alignTo4Bytes()
	offset := len(r.buffer)

	// Write inverse bind matrices for all joints
	for i := range skel.Joints {
		ibm := skel.Joints[i].InverseBindMatrix
		for j := 0; j < 16; j++ {
			data := make([]byte, 4)
			binary.LittleEndian.PutUint32(data, math.Float32bits(ibm[j]))
			r.buffer = append(r.buffer, data...)
		}
	}

	r.addBufferView(offset, len(r.buffer)-offset, 0) // No target for IBM
	r.accessors = append(r.accessors, gltfAccessor{
		BufferView:    len(r.bufferViews) - 1,
		ComponentType: 5126, // FLOAT
		Count:         len(skel.Joints),
		Type:          "MAT4",
	})
	r.skinIBMAccessor = len(r.accessors) - 1
}

func (r *animatedBuildResult) appendAnimationData(anim *Animation) {
	// Add time samples accessor
	r.alignTo4Bytes()
	timeOffset := len(r.buffer)
	for f := 0; f < anim.FrameCount; f++ {
		time := float32(f) * anim.FrameTime
		data := make([]byte, 4)
		binary.LittleEndian.PutUint32(data, math.Float32bits(time))
		r.buffer = append(r.buffer, data...)
	}
	r.addBufferView(timeOffset, len(r.buffer)-timeOffset, 0)
	r.accessors = append(r.accessors, gltfAccessor{
		BufferView:    len(r.bufferViews) - 1,
		ComponentType: 5126,
		Count:         anim.FrameCount,
		Type:          "SCALAR",
		Min:           []float64{0},
		Max:           []float64{float64(anim.FrameCount-1) * float64(anim.FrameTime)},
	})
	r.animTimeAccessor = len(r.accessors) - 1

	// Add animation data for each joint
	for _, jf := range anim.JointFrames {
		ja := jointAccessors{translationAccessor: -1, rotationAccessor: -1}

		// Add translations if present
		if len(jf.Translations) > 0 {
			r.alignTo4Bytes()
			transOffset := len(r.buffer)
			for _, t := range jf.Translations {
				r.buffer = appendVec3(r.buffer, [3]float32(t))
			}
			r.addBufferView(transOffset, len(r.buffer)-transOffset, 0)
			r.accessors = append(r.accessors, gltfAccessor{
				BufferView:    len(r.bufferViews) - 1,
				ComponentType: 5126,
				Count:         len(jf.Translations),
				Type:          "VEC3",
			})
			ja.translationAccessor = len(r.accessors) - 1
		}

		// Add rotations if present
		if len(jf.Rotations) > 0 {
			r.alignTo4Bytes()
			rotOffset := len(r.buffer)
			for _, rot := range jf.Rotations {
				r.buffer = appendVec4(r.buffer, [4]float32(rot))
			}
			r.addBufferView(rotOffset, len(r.buffer)-rotOffset, 0)
			r.accessors = append(r.accessors, gltfAccessor{
				BufferView:    len(r.bufferViews) - 1,
				ComponentType: 5126,
				Count:         len(jf.Rotations),
				Type:          "VEC4",
			})
			ja.rotationAccessor = len(r.accessors) - 1
		}

		r.jointAccessors[jf.JointID] = ja
	}
}

// buildMeshNode creates the glTF node for the mesh with skin binding.
func buildMeshNode(assetName string) gltfAnimNode {
	skinIdx := 0
	meshIdx := 0
	return gltfAnimNode{
		Name: assetName,
		Mesh: &meshIdx,
		Skin: &skinIdx,
	}
}

// buildJointNode creates a glTF node for a skeleton joint.
func buildJointNode(joint *Joint, firstJointNode int, allJoints []Joint) gltfAnimNode {
	node := gltfAnimNode{Name: joint.Name}

	pos := [3]float32(joint.Position)
	node.Translation = &pos
	rot := [4]float32(joint.Rotation)
	node.Rotation = &rot

	// Build children list
	for ci, cj := range allJoints {
		if cj.ParentID == joint.ID {
			node.Children = append(node.Children, firstJointNode+ci)
		}
	}
	return node
}

// buildSkeletonNodes creates all glTF nodes for the skeleton hierarchy.
func buildSkeletonNodes(skel *Skeleton, firstJointNode int) []gltfAnimNode {
	nodes := make([]gltfAnimNode, 1+len(skel.Joints))
	for i := range skel.Joints {
		nodes[firstJointNode+i] = buildJointNode(&skel.Joints[i], firstJointNode, skel.Joints)
	}
	return nodes
}

// findRootJointNodes returns indices of joints with no parent.
func findRootJointNodes(skel *Skeleton, firstJointNode int) []int {
	var rootNodes []int
	for i, j := range skel.Joints {
		if j.ParentID == -1 {
			rootNodes = append(rootNodes, firstJointNode+i)
		}
	}
	return rootNodes
}

// buildSkinJointList creates the ordered list of joint node indices.
func buildSkinJointList(jointCount, firstJointNode int) []int {
	joints := make([]int, jointCount)
	for i := range joints {
		joints[i] = firstJointNode + i
	}
	return joints
}

// buildGLTFSkin creates the skin definition for the skeleton.
func buildGLTFSkin(jointCount, firstJointNode, ibmAccessor int) gltfSkin {
	rootSkelNode := firstJointNode
	return gltfSkin{
		Name:                "skeleton",
		Joints:              buildSkinJointList(jointCount, firstJointNode),
		InverseBindMatrices: ibmAccessor,
		Skeleton:            &rootSkelNode,
	}
}

func buildAnimatedGLTFStructure(br *animatedBuildResult, result *AnimatedMeshResult, opts AnimatedGLTFOptions) *gltfAnimatedRoot {
	skel := result.Skeleton
	meshNodeIdx := 0
	firstJointNode := 1

	// Build nodes
	nodes := buildSkeletonNodes(skel, firstJointNode)
	nodes[meshNodeIdx] = buildMeshNode(opts.AssetName)

	rootJointNodes := findRootJointNodes(skel, firstJointNode)

	// Build animations
	var animations []gltfAnimation
	if opts.IncludeAnimation && result.Animation != nil {
		animations = buildAnimations(br, result.Animation, firstJointNode, skel)
	}

	sceneNodes := append([]int{meshNodeIdx}, rootJointNodes...)
	materialIdx := 0

	return &gltfAnimatedRoot{
		Asset:  gltfAsset{Version: "2.0", Generator: "unpeople"},
		Scene:  0,
		Scenes: []gltfScene{{Name: "Scene", Nodes: sceneNodes}},
		Nodes:  nodes,
		Meshes: []gltfMesh{{
			Name: opts.AssetName,
			Primitives: []gltfPrimitive{{
				Attributes: br.attributes,
				Indices:    br.indicesIdx,
				Mode:       4,
				Material:   &materialIdx,
			}},
		}},
		Accessors:   br.accessors,
		BufferViews: br.bufferViews,
		Buffers:     []gltfBuffer{{ByteLength: len(br.buffer)}},
		Materials:   []gltfMaterial{buildDefaultSkinMaterial()},
		Skins:       []gltfSkin{buildGLTFSkin(len(skel.Joints), firstJointNode, br.skinIBMAccessor)},
		Animations:  animations,
	}
}

// findJointNodeIndex locates the glTF node index for a joint ID.
func findJointNodeIndex(jointID JointID, skel *Skeleton, firstJointNode int) int {
	for i, j := range skel.Joints {
		if j.ID == jointID {
			return firstJointNode + i
		}
	}
	return -1
}

// addAnimationChannel appends a sampler and channel to the animation.
func addAnimationChannel(gltfAnim *gltfAnimation, timeAccessor, outputAccessor, nodeIdx int, path string) {
	samplerIdx := len(gltfAnim.Samplers)
	gltfAnim.Samplers = append(gltfAnim.Samplers, gltfSampler{
		Input:         timeAccessor,
		Output:        outputAccessor,
		Interpolation: "LINEAR",
	})
	gltfAnim.Channels = append(gltfAnim.Channels, gltfChannel{
		Sampler: samplerIdx,
		Target:  gltfChannelTarget{Node: nodeIdx, Path: path},
	})
}

func buildAnimations(br *animatedBuildResult, anim *Animation, firstJointNode int, skel *Skeleton) []gltfAnimation {
	gltfAnim := gltfAnimation{Name: anim.Name}

	for jointID, ja := range br.jointAccessors {
		nodeIdx := findJointNodeIndex(jointID, skel, firstJointNode)
		if nodeIdx == -1 {
			continue
		}

		if ja.translationAccessor >= 0 {
			addAnimationChannel(&gltfAnim, br.animTimeAccessor, ja.translationAccessor, nodeIdx, "translation")
		}
		if ja.rotationAccessor >= 0 {
			addAnimationChannel(&gltfAnim, br.animTimeAccessor, ja.rotationAccessor, nodeIdx, "rotation")
		}
	}

	if len(gltfAnim.Samplers) == 0 {
		return nil
	}
	return []gltfAnimation{gltfAnim}
}

// writeAnimatedGLBBinary writes the GLB file format for animated glTF.
func writeAnimatedGLBBinary(w io.Writer, gltf *gltfAnimatedRoot, binBuf []byte) error {
	return writeGLBBinary(w, gltf, binBuf)
}
