// Package unpeople – vertex skinning weights
//
// This file implements vertex skinning weight calculation for animation support.
// Each vertex is assigned up to 4 joint influences with normalized weights.
// Weights are calculated based on vertex proximity to joints.
package unpeople

import "math"

// ─── Skinning Configuration ──────────────────────────────────────────────────

// maxInfluences is the maximum number of joints that can influence a vertex.
const maxInfluences = 4

// SkinningParams configures the skinning weight calculation.
type SkinningParams struct {
	// FalloffRadius controls how far joint influence extends (in metres).
	// Smaller values = sharper transitions, larger = smoother blending.
	FalloffRadius float32
}

// DefaultSkinningParams returns reasonable default skinning parameters.
func DefaultSkinningParams() SkinningParams {
	return SkinningParams{
		FalloffRadius: 0.15, // 15cm falloff
	}
}

// ─── Skinning Weight Calculation ─────────────────────────────────────────────

// ComputeSkinningWeights calculates skinning weights for all vertices in a mesh
// based on proximity to skeleton joints. Weights are written directly into the
// mesh's Vertex.JointIds and Vertex.JointWeights fields.
func ComputeSkinningWeights(mesh *Mesh, skel *Skeleton, params SkinningParams) {
	for i := range mesh.Vertices {
		v := &mesh.Vertices[i]
		computeVertexWeights(v, skel, params)
	}
}

// computeVertexWeights calculates the skinning weights for a single vertex.
func computeVertexWeights(v *Vertex, skel *Skeleton, params SkinningParams) {
	// Find the closest joints and their influence weights
	influences := make([]jointInfluence, 0, len(skel.Joints))

	for i := range skel.Joints {
		joint := &skel.Joints[i]

		// Calculate distance from vertex to joint
		dist := vec3Len(vec3Sub(v.Position, joint.Position))

		// Calculate influence based on distance (exponential falloff)
		if dist < params.FalloffRadius*3 { // Only consider nearby joints
			weight := computeInfluenceWeight(dist, params.FalloffRadius, joint.ID, skel)
			if weight > 0.001 { // Threshold small weights
				influences = append(influences, jointInfluence{
					jointID: joint.ID,
					weight:  weight,
				})
			}
		}
	}

	// If no influences found, assign to nearest joint
	if len(influences) == 0 {
		nearest := findNearestJoint(v.Position, skel)
		influences = append(influences, jointInfluence{jointID: nearest, weight: 1.0})
	}

	// Sort by weight descending and keep top maxInfluences
	sortInfluencesByWeight(influences)
	if len(influences) > maxInfluences {
		influences = influences[:maxInfluences]
	}

	// Normalize weights
	normalizeInfluences(influences)

	// Write to vertex
	v.JointIds = Vec4i{0, 0, 0, 0}
	v.JointWeights = Vec4{0, 0, 0, 0}
	for i := 0; i < len(influences) && i < maxInfluences; i++ {
		v.JointIds[i] = int32(influences[i].jointID)
		v.JointWeights[i] = influences[i].weight
	}
}

// jointInfluence pairs a joint with its influence weight on a vertex.
type jointInfluence struct {
	jointID JointID
	weight  float32
}

// computeInfluenceWeight calculates the influence weight based on distance.
func computeInfluenceWeight(dist, falloffRadius float32, jointID JointID, skel *Skeleton) float32 {
	// Base weight from distance (Gaussian falloff)
	t := dist / falloffRadius
	baseWeight := float32(math.Exp(float64(-t * t * 2)))

	// Adjust weight based on joint type
	// Some joints (like spine) should have broader influence
	jointWeight := getJointInfluenceMultiplier(jointID)

	return baseWeight * jointWeight
}

// getJointInfluenceMultiplier returns a multiplier for joint influence.
// Some joints naturally have broader or narrower influence areas.
func getJointInfluenceMultiplier(id JointID) float32 {
	switch id {
	// Spine joints have broad influence for smooth torso deformation
	case JointSpine, JointSpine1, JointSpine2:
		return 1.5
	// Hips influence the entire pelvis area
	case JointHips:
		return 1.4
	// Shoulder provides transition between arm and torso
	case JointLeftShoulder, JointRightShoulder:
		return 1.3
	// Root joint has minimal direct vertex influence
	case JointRoot:
		return 0.1
	// Head top is mainly for hair attachment
	case JointHeadTop:
		return 0.3
	// Finger joints have narrow influence
	case JointLeftThumb1, JointLeftThumb2,
		JointLeftIndex1, JointLeftIndex2, JointLeftIndex3,
		JointLeftMiddle1, JointLeftMiddle2, JointLeftMiddle3,
		JointLeftRing1, JointLeftRing2, JointLeftRing3,
		JointLeftPinky1, JointLeftPinky2, JointLeftPinky3,
		JointRightThumb1, JointRightThumb2,
		JointRightIndex1, JointRightIndex2, JointRightIndex3,
		JointRightMiddle1, JointRightMiddle2, JointRightMiddle3,
		JointRightRing1, JointRightRing2, JointRightRing3,
		JointRightPinky1, JointRightPinky2, JointRightPinky3:
		return 0.8
	// Toe joints have narrow influence
	case JointLeftToe, JointRightToe:
		return 0.6
	default:
		return 1.0
	}
}

// findNearestJoint returns the ID of the joint closest to the given position.
func findNearestJoint(pos Vec3, skel *Skeleton) JointID {
	nearestID := JointID(0)
	minDist := float32(math.MaxFloat32)

	for i := range skel.Joints {
		joint := &skel.Joints[i]
		// Skip root joint (it shouldn't directly influence vertices)
		if joint.ID == JointRoot {
			continue
		}
		dist := vec3Len(vec3Sub(pos, joint.Position))
		if dist < minDist {
			minDist = dist
			nearestID = joint.ID
		}
	}
	return nearestID
}

// sortInfluencesByWeight sorts influences by weight in descending order.
func sortInfluencesByWeight(influences []jointInfluence) {
	// Simple insertion sort (small array)
	for i := 1; i < len(influences); i++ {
		j := i
		for j > 0 && influences[j].weight > influences[j-1].weight {
			influences[j], influences[j-1] = influences[j-1], influences[j]
			j--
		}
	}
}

// normalizeInfluences ensures all weights sum to 1.0.
func normalizeInfluences(influences []jointInfluence) {
	var sum float32
	for _, inf := range influences {
		sum += inf.weight
	}
	if sum > 0.001 {
		for i := range influences {
			influences[i].weight /= sum
		}
	}
}

// ─── Body Part to Joint Mapping ──────────────────────────────────────────────

// BodyPartJoints returns the primary joints that should influence a body part.
// This is used for more accurate weight assignment based on body region.
type BodyPartJoints struct {
	Primary    JointID
	Secondary  []JointID
	BlendZones []blendZone
}

// blendZone defines a region where influence transitions between joints.
type blendZone struct {
	FromJoint JointID
	ToJoint   JointID
	StartY    float32 // Y coordinate where blend starts
	EndY      float32 // Y coordinate where blend ends
}

// GetBodyPartJoints returns the joint assignments for common body regions.
// The Y coordinate of a vertex helps determine which region it belongs to.
func GetBodyPartJoints(vertexY, vertexX float32) BodyPartJoints {
	// Determine body region based on Y height
	switch {
	case vertexY > 1.5: // Head and neck region
		return BodyPartJoints{
			Primary:   JointHead,
			Secondary: []JointID{JointNeck, JointHeadTop},
		}
	case vertexY > 1.3: // Upper chest/shoulder region
		return BodyPartJoints{
			Primary:   JointSpine2,
			Secondary: []JointID{JointNeck, JointLeftShoulder, JointRightShoulder},
		}
	case vertexY > 1.0: // Chest region
		return BodyPartJoints{
			Primary:   JointSpine1,
			Secondary: []JointID{JointSpine2, JointSpine},
		}
	case vertexY > 0.8: // Abdomen/hips region
		return BodyPartJoints{
			Primary:   JointHips,
			Secondary: []JointID{JointSpine, JointLeftUpperLeg, JointRightUpperLeg},
		}
	case vertexY > 0.4: // Upper leg region
		if vertexX < 0 {
			return BodyPartJoints{
				Primary:   JointLeftUpperLeg,
				Secondary: []JointID{JointHips, JointLeftLowerLeg},
			}
		}
		return BodyPartJoints{
			Primary:   JointRightUpperLeg,
			Secondary: []JointID{JointHips, JointRightLowerLeg},
		}
	default: // Lower leg/foot region
		if vertexX < 0 {
			return BodyPartJoints{
				Primary:   JointLeftLowerLeg,
				Secondary: []JointID{JointLeftFoot, JointLeftUpperLeg},
			}
		}
		return BodyPartJoints{
			Primary:   JointRightLowerLeg,
			Secondary: []JointID{JointRightFoot, JointRightUpperLeg},
		}
	}
}

// ─── Integration with Generator ──────────────────────────────────────────────

// MeshWithRig bundles a mesh with skeleton and computed skinning weights.
type MeshWithRig struct {
	Mesh     *Mesh
	Skeleton *Skeleton
}

// GenerateWithRig produces a mesh with skeleton and skinning weights applied.
// This is the complete rigged output ready for animation.
func (g *Generator) GenerateWithRig(p Params) (*MeshWithRig, error) {
	// Generate mesh and skeleton
	result, err := g.GenerateWithSkeleton(p)
	if err != nil {
		return nil, err
	}

	// Compute skinning weights
	params := DefaultSkinningParams()
	ComputeSkinningWeights(result.Mesh, result.Skeleton, params)

	return &MeshWithRig{
		Mesh:     result.Mesh,
		Skeleton: result.Skeleton,
	}, nil
}

// ─── Skinning Validation ─────────────────────────────────────────────────────

// ValidateSkinning checks that all vertices have valid skinning data.
func ValidateSkinning(mesh *Mesh) error {
	for i, v := range mesh.Vertices {
		// Check weights sum to approximately 1.0
		sum := v.JointWeights[0] + v.JointWeights[1] + v.JointWeights[2] + v.JointWeights[3]
		if sum < 0.99 || sum > 1.01 {
			return &SkinningError{
				VertexIndex: i,
				Message:     "weights do not sum to 1.0",
			}
		}

		// Check at least one weight is non-zero
		if v.JointWeights[0] == 0 && v.JointWeights[1] == 0 &&
			v.JointWeights[2] == 0 && v.JointWeights[3] == 0 {
			return &SkinningError{
				VertexIndex: i,
				Message:     "vertex has no joint influences",
			}
		}
	}
	return nil
}

// SkinningError indicates a problem with vertex skinning data.
type SkinningError struct {
	VertexIndex int
	Message     string
}

func (e *SkinningError) Error() string {
	return "skinning error at vertex " + string(rune('0'+e.VertexIndex)) + ": " + e.Message
}

// ─── Weight Smoothing ────────────────────────────────────────────────────────

// SmoothSkinningWeights performs a smoothing pass on skinning weights.
// This helps reduce harsh transitions at joint boundaries.
func SmoothSkinningWeights(mesh *Mesh, iterations int) {
	// Build adjacency information (simplified: just use spatial proximity)
	for iter := 0; iter < iterations; iter++ {
		// Store original weights
		originalWeights := make([]Vec4, len(mesh.Vertices))
		originalIds := make([]Vec4i, len(mesh.Vertices))
		for i := range mesh.Vertices {
			originalWeights[i] = mesh.Vertices[i].JointWeights
			originalIds[i] = mesh.Vertices[i].JointIds
		}

		// Average with nearby vertices (simplified nearest-neighbor smoothing)
		for i := range mesh.Vertices {
			v := &mesh.Vertices[i]
			smoothVertex(v, mesh.Vertices, originalWeights, originalIds, i)
		}
	}
}

// smoothVertex averages weights with nearby vertices.
func smoothVertex(v *Vertex, allVerts []Vertex, origWeights []Vec4, origIds []Vec4i, selfIdx int) {
	const searchRadius = float32(0.05) // 5cm radius for neighbor search
	const blendFactor = float32(0.3)   // How much to blend with neighbors

	var neighborCount int
	var avgWeights [maxInfluences]float32
	jointPresent := make(map[int32]int) // joint ID -> slot in avgWeights

	// Map current joints to slots
	for slot := 0; slot < maxInfluences; slot++ {
		jointPresent[origIds[selfIdx][slot]] = slot
	}

	// Find nearby vertices and accumulate their weights
	for j := range allVerts {
		if j == selfIdx {
			continue
		}
		dist := vec3Len(vec3Sub(v.Position, allVerts[j].Position))
		if dist < searchRadius {
			neighborCount++
			// Add neighbor's weight contributions
			for k := 0; k < maxInfluences; k++ {
				jointID := origIds[j][k]
				weight := origWeights[j][k]
				if slot, ok := jointPresent[jointID]; ok {
					avgWeights[slot] += weight
				}
			}
		}
	}

	// Blend with self weights
	if neighborCount > 0 {
		neighborFactor := blendFactor / float32(neighborCount)
		selfFactor := 1.0 - blendFactor

		for slot := 0; slot < maxInfluences; slot++ {
			v.JointWeights[slot] = selfFactor*origWeights[selfIdx][slot] +
				neighborFactor*avgWeights[slot]
		}

		// Renormalize
		sum := v.JointWeights[0] + v.JointWeights[1] + v.JointWeights[2] + v.JointWeights[3]
		if sum > 0.001 {
			for slot := 0; slot < maxInfluences; slot++ {
				v.JointWeights[slot] /= sum
			}
		}
	}
}
