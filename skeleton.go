// Package unpeople – skeleton and joint hierarchy
//
// This file implements the bind-pose skeleton for humanoid meshes. The skeleton
// consists of a joint hierarchy (root → spine → shoulders/hips → limb chains)
// that matches the generated mesh geometry. Joint positions are derived from
// the bodyLayout to ensure correspondence between skeleton and mesh.
package unpeople

import (
	"fmt"
	"math"
)

// ─── Joint Types ─────────────────────────────────────────────────────────────

// JointID uniquely identifies a joint within a skeleton.
type JointID int

// Joint IDs for the humanoid skeleton.
// These correspond to standard rigging conventions used in game engines.
const (
	JointRoot JointID = iota
	JointHips
	JointSpine
	JointSpine1
	JointSpine2
	JointNeck
	JointHead
	JointHeadTop // For hair attachment

	// Left arm chain
	JointLeftShoulder
	JointLeftUpperArm
	JointLeftForearm
	JointLeftHand
	JointLeftThumb1
	JointLeftThumb2
	JointLeftIndex1
	JointLeftIndex2
	JointLeftIndex3
	JointLeftMiddle1
	JointLeftMiddle2
	JointLeftMiddle3
	JointLeftRing1
	JointLeftRing2
	JointLeftRing3
	JointLeftPinky1
	JointLeftPinky2
	JointLeftPinky3

	// Right arm chain
	JointRightShoulder
	JointRightUpperArm
	JointRightForearm
	JointRightHand
	JointRightThumb1
	JointRightThumb2
	JointRightIndex1
	JointRightIndex2
	JointRightIndex3
	JointRightMiddle1
	JointRightMiddle2
	JointRightMiddle3
	JointRightRing1
	JointRightRing2
	JointRightRing3
	JointRightPinky1
	JointRightPinky2
	JointRightPinky3

	// Left leg chain
	JointLeftUpperLeg
	JointLeftLowerLeg
	JointLeftFoot
	JointLeftToe

	// Right leg chain
	JointRightUpperLeg
	JointRightLowerLeg
	JointRightFoot
	JointRightToe

	// Total joint count
	JointCount
)

// Joint represents a single bone/joint in the skeleton hierarchy.
type Joint struct {
	ID       JointID
	Name     string
	ParentID JointID // -1 for root

	// Bind pose in world space
	Position Vec3 // World position of joint
	Rotation Vec4 // Quaternion rotation (x, y, z, w)
	Scale    Vec3 // Usually (1, 1, 1)

	// Inverse bind matrix (computed from position/rotation)
	InverseBindMatrix [16]float32
}

// Skeleton represents the complete joint hierarchy for a humanoid mesh.
type Skeleton struct {
	Joints []Joint
}

// ─── Skeleton Generation ─────────────────────────────────────────────────────

// aPoseAngle is the rotation angle for A-pose (45 degrees in radians).
// A-pose rotates shoulders ~45° downward from horizontal T-pose.
const aPoseAngle = math.Pi / 4.0 // 45 degrees

// GenerateSkeleton creates a bind-pose skeleton matching the given body layout.
// The skeleton joints are positioned to match the mesh geometry exactly.
// This function generates a T-pose skeleton; use GenerateSkeletonWithPose for A-pose.
func GenerateSkeleton(layout bodyLayout) *Skeleton {
	return GenerateSkeletonWithPose(layout, SkeletonPoseTPose)
}

// GenerateSkeletonWithPose creates a skeleton with the specified bind pose.
// For A-pose, shoulder joints are rotated ~45° downward and arm positions adjusted.
func GenerateSkeletonWithPose(layout bodyLayout, pose SkeletonPose) *Skeleton {
	skel := &Skeleton{
		Joints: make([]Joint, JointCount),
	}

	// Initialize all joints with identity transforms
	for i := range skel.Joints {
		skel.Joints[i] = Joint{
			ID:       JointID(i),
			Name:     jointNames[i],
			ParentID: jointParents[i],
			Rotation: Vec4{0, 0, 0, 1}, // Identity quaternion
			Scale:    Vec3{1, 1, 1},
		}
	}

	// Calculate joint positions from body layout
	computeJointPositions(skel, layout)

	// Apply A-pose transformation if requested
	if pose == SkeletonPoseAPose {
		applyAPoseTransform(skel, layout)
	}

	// Compute inverse bind matrices
	computeInverseBindMatrices(skel)

	return skel
}

// applyAPoseTransform rotates shoulder joints and repositions arm chains for A-pose.
// A-pose has arms at ~45° downward from horizontal, providing better shoulder deformation.
func applyAPoseTransform(skel *Skeleton, l bodyLayout) {
	// Create Z-axis rotation quaternion for ~45° downward angle
	// Left arm rotates clockwise (-45°), right arm rotates counter-clockwise (+45°)
	leftRot := quatFromAxisAngle(Vec3{0, 0, 1}, aPoseAngle)   // +45° around Z (arms down)
	rightRot := quatFromAxisAngle(Vec3{0, 0, 1}, -aPoseAngle) // -45° around Z (arms down)

	// Apply rotation to shoulder joints
	skel.Joints[JointLeftShoulder].Rotation = leftRot
	skel.Joints[JointRightShoulder].Rotation = rightRot

	// Get shoulder pivot points
	leftShoulderPos := skel.Joints[JointLeftShoulder].Position
	rightShoulderPos := skel.Joints[JointRightShoulder].Position

	// Rotate all joints in the left arm chain around the left shoulder
	rotateArmChain(skel, leftShoulderPos, leftRot, []JointID{
		JointLeftUpperArm, JointLeftForearm, JointLeftHand,
		JointLeftThumb1, JointLeftThumb2,
		JointLeftIndex1, JointLeftIndex2, JointLeftIndex3,
		JointLeftMiddle1, JointLeftMiddle2, JointLeftMiddle3,
		JointLeftRing1, JointLeftRing2, JointLeftRing3,
		JointLeftPinky1, JointLeftPinky2, JointLeftPinky3,
	})

	// Rotate all joints in the right arm chain around the right shoulder
	rotateArmChain(skel, rightShoulderPos, rightRot, []JointID{
		JointRightUpperArm, JointRightForearm, JointRightHand,
		JointRightThumb1, JointRightThumb2,
		JointRightIndex1, JointRightIndex2, JointRightIndex3,
		JointRightMiddle1, JointRightMiddle2, JointRightMiddle3,
		JointRightRing1, JointRightRing2, JointRightRing3,
		JointRightPinky1, JointRightPinky2, JointRightPinky3,
	})
}

// rotateArmChain rotates a list of joint positions around a pivot point.
func rotateArmChain(skel *Skeleton, pivot Vec3, rot Vec4, joints []JointID) {
	for _, jid := range joints {
		j := &skel.Joints[jid]
		// Translate to origin relative to pivot
		rel := vec3Sub(j.Position, pivot)
		// Rotate the position
		rotated := quatRotateVec3(rot, rel)
		// Translate back
		j.Position = vec3Add(pivot, rotated)
	}
}

// quatFromAxisAngle creates a quaternion from an axis and angle (radians).
func quatFromAxisAngle(axis Vec3, angle float32) Vec4 {
	halfAngle := angle / 2.0
	s := float32(math.Sin(float64(halfAngle)))
	c := float32(math.Cos(float64(halfAngle)))
	return Vec4{axis[0] * s, axis[1] * s, axis[2] * s, c}
}

// quatRotateVec3 rotates a vector by a quaternion.
func quatRotateVec3(q Vec4, v Vec3) Vec3 {
	// q * v * q^-1, using the optimized formula:
	// t = 2 * cross(q.xyz, v)
	// result = v + q.w * t + cross(q.xyz, t)
	qv := Vec3{q[0], q[1], q[2]}
	t := vec3Scale(vec3Cross(qv, v), 2.0)
	return vec3Add(vec3Add(v, vec3Scale(t, q[3])), vec3Cross(qv, t))
}

// computeJointPositions sets joint positions based on body layout.
func computeJointPositions(skel *Skeleton, l bodyLayout) {
	// Root at ground level
	skel.Joints[JointRoot].Position = Vec3{0, 0, 0}

	// Hips at hip center (midpoint between hipsTop and hipsBottom)
	hipsCenterY := (l.hipsTop[1] + l.hipsBottom[1]) / 2.0
	skel.Joints[JointHips].Position = Vec3{0, hipsCenterY, 0}

	// Spine chain
	spineBaseY := l.abdomenBottom[1]
	chestCenterY := (l.chestTop[1] + l.chestBottom[1]) / 2.0
	spineStep := (chestCenterY - spineBaseY) / 3.0
	skel.Joints[JointSpine].Position = Vec3{0, spineBaseY + spineStep, 0}
	skel.Joints[JointSpine1].Position = Vec3{0, spineBaseY + spineStep*2, 0}
	skel.Joints[JointSpine2].Position = Vec3{0, chestCenterY, 0}

	// Neck center
	neckCenterY := (l.neckTop[1] + l.neckBottom[1]) / 2.0
	skel.Joints[JointNeck].Position = Vec3{0, neckCenterY, 0}
	skel.Joints[JointHead].Position = l.headCenter
	skel.Joints[JointHeadTop].Position = Vec3{
		l.headCenter[0],
		l.headCenter[1] + l.headRY,
		l.headCenter[2],
	}

	// Left arm chain - shoulders attach at chest top
	skel.Joints[JointLeftShoulder].Position = Vec3{-l.chestRX, l.chestTop[1], 0}
	skel.Joints[JointLeftUpperArm].Position = l.upperArmTopL
	skel.Joints[JointLeftForearm].Position = l.forearmTopL
	skel.Joints[JointLeftHand].Position = l.handCenterL
	computeFingerJoints(skel, l, true)

	// Right arm chain
	skel.Joints[JointRightShoulder].Position = Vec3{l.chestRX, l.chestTop[1], 0}
	skel.Joints[JointRightUpperArm].Position = l.upperArmTopR
	skel.Joints[JointRightForearm].Position = l.forearmTopR
	skel.Joints[JointRightHand].Position = l.handCenterR
	computeFingerJoints(skel, l, false)

	// Left leg chain
	skel.Joints[JointLeftUpperLeg].Position = l.upperLegTopL
	skel.Joints[JointLeftLowerLeg].Position = l.lowerLegTopL
	skel.Joints[JointLeftFoot].Position = l.footCenterL
	skel.Joints[JointLeftToe].Position = Vec3{
		l.footCenterL[0],
		l.footCenterL[1],
		l.footCenterL[2] + l.footHD*0.8,
	}

	// Right leg chain
	skel.Joints[JointRightUpperLeg].Position = l.upperLegTopR
	skel.Joints[JointRightLowerLeg].Position = l.lowerLegTopR
	skel.Joints[JointRightFoot].Position = l.footCenterR
	skel.Joints[JointRightToe].Position = Vec3{
		l.footCenterR[0],
		l.footCenterR[1],
		l.footCenterR[2] + l.footHD*0.8,
	}
}

// computeFingerJoints calculates finger joint positions.
// fingerConfig holds the computed configuration for one hand's fingers.
type fingerConfig struct {
	handCenter Vec3
	baseID     JointID
	sign       float32
}

// computeFingerConfig returns the configuration for left or right hand fingers.
func computeFingerConfig(l bodyLayout, isLeft bool) fingerConfig {
	if isLeft {
		return fingerConfig{handCenter: l.handCenterL, baseID: JointLeftThumb1, sign: -1.0}
	}
	return fingerConfig{handCenter: l.handCenterR, baseID: JointRightThumb1, sign: 1.0}
}

// computeFingerOffsets returns X-offsets for each finger from palm center.
func computeFingerOffsets(sign, handHW float32) [5]float32 {
	return [5]float32{
		-sign * handHW * 0.8, // Thumb (offset laterally)
		-sign * handHW * 0.5, // Index
		-sign * handHW * 0.2, // Middle
		sign * handHW * 0.2,  // Ring
		sign * handHW * 0.5,  // Pinky
	}
}

// setFingerJointPositions positions joints for a single finger.
func setFingerJointPositions(skel *Skeleton, cfg fingerConfig, offsets [5]float32, palmEdge, phalanxLen float32, fingerIdx int) {
	numJoints := 3
	if fingerIdx == 0 { // Thumb has 2 joints
		numJoints = 2
	}

	for j := 0; j < numJoints; j++ {
		jointIdx := cfg.baseID + JointID(fingerIdx*3+j)
		if jointIdx >= JointCount {
			break
		}
		yOffset := palmEdge - float32(j+1)*phalanxLen
		skel.Joints[jointIdx].Position = Vec3{
			cfg.handCenter[0] + offsets[fingerIdx],
			yOffset,
			cfg.handCenter[2],
		}
	}
}

func computeFingerJoints(skel *Skeleton, l bodyLayout, isLeft bool) {
	cfg := computeFingerConfig(l, isLeft)
	offsets := computeFingerOffsets(cfg.sign, l.handHW)

	fingerLength := (l.proximalLength + l.middleLength + l.distalLength) * l.fingerLengthMult
	phalanxLen := fingerLength / 3.0
	palmEdge := cfg.handCenter[1] - l.handHH

	for f := 0; f < 5; f++ {
		setFingerJointPositions(skel, cfg, offsets, palmEdge, phalanxLen, f)
	}
}

// ─── Inverse Bind Matrices ───────────────────────────────────────────────────

// computeInverseBindMatrices calculates the inverse bind matrix for each joint.
func computeInverseBindMatrices(skel *Skeleton) {
	for i := range skel.Joints {
		j := &skel.Joints[i]
		// Inverse bind matrix = inverse of (translation * rotation * scale)
		// For bind pose with identity rotation and unit scale, this is just
		// the inverse translation matrix.
		j.InverseBindMatrix = inverseTranslationMatrix(j.Position)
	}
}

// inverseTranslationMatrix creates the inverse of a translation matrix.
func inverseTranslationMatrix(pos Vec3) [16]float32 {
	return [16]float32{
		1, 0, 0, 0,
		0, 1, 0, 0,
		0, 0, 1, 0,
		-pos[0], -pos[1], -pos[2], 1,
	}
}

// ─── Joint Hierarchy ─────────────────────────────────────────────────────────

// jointNames maps joint IDs to human-readable names.
var jointNames = [JointCount]string{
	"Root",
	"Hips",
	"Spine",
	"Spine1",
	"Spine2",
	"Neck",
	"Head",
	"HeadTop",
	"LeftShoulder",
	"LeftUpperArm",
	"LeftForearm",
	"LeftHand",
	"LeftThumb1",
	"LeftThumb2",
	"LeftIndex1",
	"LeftIndex2",
	"LeftIndex3",
	"LeftMiddle1",
	"LeftMiddle2",
	"LeftMiddle3",
	"LeftRing1",
	"LeftRing2",
	"LeftRing3",
	"LeftPinky1",
	"LeftPinky2",
	"LeftPinky3",
	"RightShoulder",
	"RightUpperArm",
	"RightForearm",
	"RightHand",
	"RightThumb1",
	"RightThumb2",
	"RightIndex1",
	"RightIndex2",
	"RightIndex3",
	"RightMiddle1",
	"RightMiddle2",
	"RightMiddle3",
	"RightRing1",
	"RightRing2",
	"RightRing3",
	"RightPinky1",
	"RightPinky2",
	"RightPinky3",
	"LeftUpperLeg",
	"LeftLowerLeg",
	"LeftFoot",
	"LeftToe",
	"RightUpperLeg",
	"RightLowerLeg",
	"RightFoot",
	"RightToe",
}

// jointParents defines the parent-child relationship of joints.
// -1 indicates root joint (no parent).
var jointParents = [JointCount]JointID{
	-1,                 // Root
	JointRoot,          // Hips
	JointHips,          // Spine
	JointSpine,         // Spine1
	JointSpine1,        // Spine2
	JointSpine2,        // Neck
	JointNeck,          // Head
	JointHead,          // HeadTop
	JointSpine2,        // LeftShoulder
	JointLeftShoulder,  // LeftUpperArm
	JointLeftUpperArm,  // LeftForearm
	JointLeftForearm,   // LeftHand
	JointLeftHand,      // LeftThumb1
	JointLeftThumb1,    // LeftThumb2
	JointLeftHand,      // LeftIndex1
	JointLeftIndex1,    // LeftIndex2
	JointLeftIndex2,    // LeftIndex3
	JointLeftHand,      // LeftMiddle1
	JointLeftMiddle1,   // LeftMiddle2
	JointLeftMiddle2,   // LeftMiddle3
	JointLeftHand,      // LeftRing1
	JointLeftRing1,     // LeftRing2
	JointLeftRing2,     // LeftRing3
	JointLeftHand,      // LeftPinky1
	JointLeftPinky1,    // LeftPinky2
	JointLeftPinky2,    // LeftPinky3
	JointSpine2,        // RightShoulder
	JointRightShoulder, // RightUpperArm
	JointRightUpperArm, // RightForearm
	JointRightForearm,  // RightHand
	JointRightHand,     // RightThumb1
	JointRightThumb1,   // RightThumb2
	JointRightHand,     // RightIndex1
	JointRightIndex1,   // RightIndex2
	JointRightIndex2,   // RightIndex3
	JointRightHand,     // RightMiddle1
	JointRightMiddle1,  // RightMiddle2
	JointRightMiddle2,  // RightMiddle3
	JointRightHand,     // RightRing1
	JointRightRing1,    // RightRing2
	JointRightRing2,    // RightRing3
	JointRightHand,     // RightPinky1
	JointRightPinky1,   // RightPinky2
	JointRightPinky2,   // RightPinky3
	JointHips,          // LeftUpperLeg
	JointLeftUpperLeg,  // LeftLowerLeg
	JointLeftLowerLeg,  // LeftFoot
	JointLeftFoot,      // LeftToe
	JointHips,          // RightUpperLeg
	JointRightUpperLeg, // RightLowerLeg
	JointRightLowerLeg, // RightFoot
	JointRightFoot,     // RightToe
}

// ─── Skeleton Accessors ──────────────────────────────────────────────────────

// Joint returns the joint at the given ID.
func (s *Skeleton) Joint(id JointID) *Joint {
	if id < 0 || int(id) >= len(s.Joints) {
		return nil
	}
	return &s.Joints[id]
}

// JointByName finds a joint by name.
func (s *Skeleton) JointByName(name string) *Joint {
	for i := range s.Joints {
		if s.Joints[i].Name == name {
			return &s.Joints[i]
		}
	}
	return nil
}

// Parent returns the parent joint of the given joint.
func (s *Skeleton) Parent(j *Joint) *Joint {
	if j.ParentID < 0 {
		return nil
	}
	return s.Joint(j.ParentID)
}

// Children returns all joints that have the given joint as parent.
func (s *Skeleton) Children(j *Joint) []*Joint {
	var children []*Joint
	for i := range s.Joints {
		if s.Joints[i].ParentID == j.ID {
			children = append(children, &s.Joints[i])
		}
	}
	return children
}

// ─── MeshWithSkeleton ────────────────────────────────────────────────────────

// MeshWithSkeleton bundles a mesh with its corresponding skeleton.
type MeshWithSkeleton struct {
	Mesh     *Mesh
	Skeleton *Skeleton
}

// GenerateWithSkeleton produces a mesh with its bind-pose skeleton.
func (g *Generator) GenerateWithSkeleton(p Params) (*MeshWithSkeleton, error) {
	if err := p.Validate(); err != nil {
		return nil, fmt.Errorf("unpeople: invalid params: %w", err)
	}

	rng := newSplitmix64(p.Seed)
	layout := computeBodyLayout(&p, rng)

	mesh, err := g.Generate(p)
	if err != nil {
		return nil, err
	}

	skeleton := GenerateSkeletonWithPose(layout, p.SkeletonPose)

	return &MeshWithSkeleton{
		Mesh:     mesh,
		Skeleton: skeleton,
	}, nil
}

// ─── Skeleton Export ─────────────────────────────────────────────────────────

// BoneLength returns the length of the bone from this joint to its first child.
// Returns 0 if the joint has no children.
func (s *Skeleton) BoneLength(j *Joint) float32 {
	children := s.Children(j)
	if len(children) == 0 {
		return 0
	}
	// Return distance to first child
	return vec3Len(vec3Sub(children[0].Position, j.Position))
}

// TotalBoneCount returns the number of bones (joint connections) in the skeleton.
func (s *Skeleton) TotalBoneCount() int {
	count := 0
	for i := range s.Joints {
		if s.Joints[i].ParentID >= 0 {
			count++
		}
	}
	return count
}

// Validate checks that the skeleton hierarchy is consistent.
func (s *Skeleton) Validate() error {
	for i := range s.Joints {
		j := &s.Joints[i]

		// Check parent reference is valid
		if j.ParentID >= 0 {
			if int(j.ParentID) >= len(s.Joints) {
				return fmt.Errorf("joint %s has invalid parent ID %d", j.Name, j.ParentID)
			}
		}

		// Check that joint ID matches index
		if int(j.ID) != i {
			return fmt.Errorf("joint at index %d has mismatched ID %d", i, j.ID)
		}
	}
	return nil
}

// ─── T-Pose / A-Pose Validation ───────────────────────────────────────────────
//
// The bind pose conforms to these industry-standard conventions for A-pose
// (which is preferred by many modern game engines over T-pose):
//
// 1. **Root at Origin**: JointRoot is at (0, 0, 0)
// 2. **Y-Up Orientation**: Spine chain extends upward along positive Y
// 3. **Arms at Sides**: Arms extend downward from shoulders along the body,
//    with slight lateral offset (A-pose rather than T-pose)
// 4. **Palms Inward**: Hands are oriented with palms facing the body
// 5. **Legs Pointing Down**: Legs extend downward along negative Y
// 6. **Feet Flat**: Feet are positioned near Y=0 with toes pointing forward (+Z)
// 7. **Identity Rotation**: All joints have identity quaternion rotation
// 8. **Unit Scale**: All joints have (1, 1, 1) scale
//
// A-pose advantages:
// - More natural shoulder deformation during animation
// - Better weight distribution at shoulder joint
// - Reduced texture stretching on deltoids
// - Compatible with Unity, Unreal, and most motion capture sources
//
// Note: The mesh can be converted to T-pose at runtime by rotating shoulder
// joints if required for specific animation retargeting pipelines.

// TPoseError describes a deviation from standard bind-pose conventions.
type TPoseError struct {
	JointID JointID
	Issue   string
	Details string
}

// Error implements the error interface for TPoseError.
func (e *TPoseError) Error() string {
	return fmt.Sprintf("bind-pose error at %s: %s (%s)", jointNames[e.JointID], e.Issue, e.Details)
}

// validateRootAtOrigin checks that the root joint is at the origin.
func (s *Skeleton) validateRootAtOrigin() *TPoseError {
	root := s.Joint(JointRoot)
	if vec3Len(root.Position) > 0.001 {
		return &TPoseError{
			JointID: JointRoot,
			Issue:   "not at origin",
			Details: fmt.Sprintf("position (%.3f, %.3f, %.3f)", root.Position[0], root.Position[1], root.Position[2]),
		}
	}
	return nil
}

// validateSpineUpward checks that the head is above the hips.
func (s *Skeleton) validateSpineUpward() *TPoseError {
	hips := s.Joint(JointHips)
	head := s.Joint(JointHead)
	if head.Position[1] <= hips.Position[1] {
		return &TPoseError{
			JointID: JointHead,
			Issue:   "not above hips",
			Details: fmt.Sprintf("head Y=%.3f, hips Y=%.3f", head.Position[1], hips.Position[1]),
		}
	}
	return nil
}

// validateArmPosition checks that a hand is positioned correctly for A-pose.
func validateArmPosition(hand, shoulder *Joint, isLeft bool) []error {
	var errs []error
	jointID := JointRightHand
	if isLeft {
		jointID = JointLeftHand
	}

	// Arms should be below shoulders (A-pose)
	if hand.Position[1] >= shoulder.Position[1] {
		errs = append(errs, &TPoseError{
			JointID: jointID,
			Issue:   "hand not below shoulder",
			Details: fmt.Sprintf("hand Y=%.3f, shoulder Y=%.3f", hand.Position[1], shoulder.Position[1]),
		})
	}

	// Check lateral position
	if isLeft && hand.Position[0] >= 0 {
		errs = append(errs, &TPoseError{
			JointID: jointID,
			Issue:   "left hand not on left side",
			Details: fmt.Sprintf("hand X=%.3f", hand.Position[0]),
		})
	} else if !isLeft && hand.Position[0] <= 0 {
		errs = append(errs, &TPoseError{
			JointID: jointID,
			Issue:   "right hand not on right side",
			Details: fmt.Sprintf("hand X=%.3f", hand.Position[0]),
		})
	}
	return errs
}

// validateFootNearGround checks that a foot is near ground level.
func validateFootNearGround(foot *Joint, jointID JointID, tolerance float32) *TPoseError {
	if foot.Position[1] > tolerance {
		return &TPoseError{
			JointID: jointID,
			Issue:   "foot not near ground",
			Details: fmt.Sprintf("foot Y=%.3f", foot.Position[1]),
		}
	}
	return nil
}

// validateJointRotations checks that all joints have identity rotations.
func (s *Skeleton) validateJointRotations() []error {
	var errs []error
	identityQuat := Vec4{0, 0, 0, 1}
	for i := range s.Joints {
		j := &s.Joints[i]
		if j.Rotation != identityQuat {
			errs = append(errs, &TPoseError{
				JointID: j.ID,
				Issue:   "non-identity rotation",
				Details: fmt.Sprintf("rotation (%.3f, %.3f, %.3f, %.3f)",
					j.Rotation[0], j.Rotation[1], j.Rotation[2], j.Rotation[3]),
			})
		}
	}
	return errs
}

// validateJointScales checks that all joints have unit scales.
func (s *Skeleton) validateJointScales() []error {
	var errs []error
	unitScale := Vec3{1, 1, 1}
	for i := range s.Joints {
		j := &s.Joints[i]
		if j.Scale != unitScale {
			errs = append(errs, &TPoseError{
				JointID: j.ID,
				Issue:   "non-unit scale",
				Details: fmt.Sprintf("scale (%.3f, %.3f, %.3f)",
					j.Scale[0], j.Scale[1], j.Scale[2]),
			})
		}
	}
	return errs
}

// ValidateTPose checks that the skeleton conforms to industry-standard bind-pose
// conventions (A-pose variant) for animation compatibility. Returns nil if the
// pose is valid, or a list of issues otherwise.
func (s *Skeleton) ValidateTPose() []error {
	var errors []error

	if err := s.validateRootAtOrigin(); err != nil {
		errors = append(errors, err)
	}

	if err := s.validateSpineUpward(); err != nil {
		errors = append(errors, err)
	}

	// Validate arm positions (A-pose: arms extend downward)
	leftShoulder := s.Joint(JointLeftShoulder)
	leftHand := s.Joint(JointLeftHand)
	rightShoulder := s.Joint(JointRightShoulder)
	rightHand := s.Joint(JointRightHand)
	errors = append(errors, validateArmPosition(leftHand, leftShoulder, true)...)
	errors = append(errors, validateArmPosition(rightHand, rightShoulder, false)...)

	// Validate feet near ground
	groundTolerance := float32(0.15)
	if err := validateFootNearGround(s.Joint(JointLeftFoot), JointLeftFoot, groundTolerance); err != nil {
		errors = append(errors, err)
	}
	if err := validateFootNearGround(s.Joint(JointRightFoot), JointRightFoot, groundTolerance); err != nil {
		errors = append(errors, err)
	}

	errors = append(errors, s.validateJointRotations()...)
	errors = append(errors, s.validateJointScales()...)

	return errors
}

// IsTPoseValid returns true if the skeleton passes all T-pose validation checks.
func (s *Skeleton) IsTPoseValid() bool {
	return len(s.ValidateTPose()) == 0
}

// ─── Skeleton Export Data ────────────────────────────────────────────────────

// SkeletonExportData contains the skeleton in a format suitable for export
// to standard animation formats like glTF or FBX.
type SkeletonExportData struct {
	// JointNames in hierarchy order (parent before child)
	JointNames []string
	// ParentIndices maps each joint to its parent (-1 for root)
	ParentIndices []int
	// InverseBindMatrices for skinning (column-major, 4x4)
	InverseBindMatrices [][16]float32
	// BindPosePositions for each joint in world space
	BindPosePositions []Vec3
	// BindPoseRotations for each joint (quaternion)
	BindPoseRotations []Vec4
}

// ExportData returns the skeleton data in a format suitable for standard exports.
func (s *Skeleton) ExportData() *SkeletonExportData {
	data := &SkeletonExportData{
		JointNames:          make([]string, len(s.Joints)),
		ParentIndices:       make([]int, len(s.Joints)),
		InverseBindMatrices: make([][16]float32, len(s.Joints)),
		BindPosePositions:   make([]Vec3, len(s.Joints)),
		BindPoseRotations:   make([]Vec4, len(s.Joints)),
	}

	for i, j := range s.Joints {
		data.JointNames[i] = j.Name
		data.ParentIndices[i] = int(j.ParentID)
		data.InverseBindMatrices[i] = j.InverseBindMatrix
		data.BindPosePositions[i] = j.Position
		data.BindPoseRotations[i] = j.Rotation
	}

	return data
}

// ─── Animation-Ready Export ──────────────────────────────────────────────────

// AnimationReadyExport bundles mesh, skeleton, skinning, and morph data
// in a complete format suitable for animation in external engines.
type AnimationReadyExport struct {
	Mesh           *Mesh
	Skeleton       *Skeleton
	SkeletonData   *SkeletonExportData
	MorphTargets   *MorphTargetSet
	TPoseValid     bool
	ValidationErrs []error
}

// GenerateAnimationReady produces a complete animation-ready character export.
func (g *Generator) GenerateAnimationReady(p Params) (*AnimationReadyExport, error) {
	result, err := g.GenerateWithMorphs(p)
	if err != nil {
		return nil, err
	}

	// Validate T-pose
	tposeErrs := result.Skeleton.ValidateTPose()

	return &AnimationReadyExport{
		Mesh:           result.Mesh,
		Skeleton:       result.Skeleton,
		SkeletonData:   result.Skeleton.ExportData(),
		MorphTargets:   result.Morphs,
		TPoseValid:     len(tposeErrs) == 0,
		ValidationErrs: tposeErrs,
	}, nil
}
