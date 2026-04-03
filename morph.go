// Package unpeople – morph target (blend shape) support
//
// This file implements morph targets for blend-shape animation. Morph targets
// store per-vertex position offsets that can be blended at runtime to create
// facial expressions, breathing effects, and other deformations.
package unpeople

// ─── Morph Target Types ──────────────────────────────────────────────────────

// MorphTargetType identifies different morph target categories.
type MorphTargetType int

const (
	// Facial expressions
	MorphSmile MorphTargetType = iota
	MorphFrown
	MorphEyebrowsRaised
	MorphEyebrowsFurrowed
	MorphEyesClosed
	MorphMouthOpen
	MorphMouthWide
	MorphMouthNarrow
	MorphJawOpen
	MorphNoseFlare
	MorphCheekPuff

	// Body deformations
	MorphBreathIn
	MorphBreathOut
	MorphFlex
	MorphRelax

	// Corrective shapes
	MorphShoulderCorrectL
	MorphShoulderCorrectR
	MorphElbowCorrectL
	MorphElbowCorrectR

	MorphTargetCount
)

// MorphTarget represents a single blend shape target.
type MorphTarget struct {
	Type     MorphTargetType
	Name     string
	Offsets  []Vec3  // Per-vertex position offsets (same length as mesh vertices)
	Strength float32 // Default blend weight (0-1)
}

// MorphTargetSet contains all morph targets for a mesh.
type MorphTargetSet struct {
	Targets []MorphTarget
}

// ─── Morph Target Names ──────────────────────────────────────────────────────

// morphTargetNames maps morph target types to human-readable names.
var morphTargetNames = [MorphTargetCount]string{
	"smile",
	"frown",
	"eyebrows_raised",
	"eyebrows_furrowed",
	"eyes_closed",
	"mouth_open",
	"mouth_wide",
	"mouth_narrow",
	"jaw_open",
	"nose_flare",
	"cheek_puff",
	"breath_in",
	"breath_out",
	"flex",
	"relax",
	"shoulder_correct_l",
	"shoulder_correct_r",
	"elbow_correct_l",
	"elbow_correct_r",
}

// Name returns the string name for a morph target type.
func (m MorphTargetType) Name() string {
	if m >= 0 && m < MorphTargetCount {
		return morphTargetNames[m]
	}
	return "unknown"
}

// ─── Morph Target Generation ─────────────────────────────────────────────────

// GenerateMorphTargets creates a complete set of morph targets for a mesh.
// The offsets are computed based on vertex positions relative to the skeleton.
func GenerateMorphTargets(mesh *Mesh, skel *Skeleton, layout bodyLayout) *MorphTargetSet {
	set := &MorphTargetSet{
		Targets: make([]MorphTarget, 0, MorphTargetCount),
	}

	// Generate facial expression morphs
	set.Targets = append(set.Targets, generateSmileMorph(mesh, skel, layout))
	set.Targets = append(set.Targets, generateFrownMorph(mesh, skel, layout))
	set.Targets = append(set.Targets, generateEyebrowsRaisedMorph(mesh, skel, layout))
	set.Targets = append(set.Targets, generateEyesFurrowedMorph(mesh, skel, layout))
	set.Targets = append(set.Targets, generateEyesClosedMorph(mesh, skel, layout))
	set.Targets = append(set.Targets, generateMouthOpenMorph(mesh, skel, layout))
	set.Targets = append(set.Targets, generateJawOpenMorph(mesh, skel, layout))

	// Generate body morphs
	set.Targets = append(set.Targets, generateBreathInMorph(mesh, skel, layout))
	set.Targets = append(set.Targets, generateBreathOutMorph(mesh, skel, layout))
	set.Targets = append(set.Targets, generateFlexMorph(mesh, skel, layout))

	return set
}

// ─── Facial Morph Generators ─────────────────────────────────────────────────

// generateSmileMorph creates the smile expression morph target.
func generateSmileMorph(mesh *Mesh, skel *Skeleton, layout bodyLayout) MorphTarget {
	offsets := make([]Vec3, len(mesh.Vertices))
	headJoint := skel.Joint(JointHead)

	for i, v := range mesh.Vertices {
		// Only affect head region vertices
		if !isInHeadRegion(v.Position, headJoint.Position, layout) {
			continue
		}

		// Calculate relative position within head
		localPos := vec3Sub(v.Position, headJoint.Position)

		// Smile affects corners of mouth (lower front of head)
		// Pull corners up and back
		if localPos[1] < 0 && localPos[2] > 0 { // Below center, in front
			mouthFactor := gaussianInfluence(localPos[1], -layout.headRY*0.6, 0.03)
			sideFactor := absFloat32(localPos[0]) / layout.headRX

			if sideFactor > 0.3 { // Only affect sides (mouth corners)
				intensity := mouthFactor * sideFactor * 0.015
				offsets[i] = Vec3{
					0,                // No X movement
					intensity,        // Pull up
					-intensity * 0.5, // Pull back slightly
				}
			}
		}
	}

	return MorphTarget{
		Type:     MorphSmile,
		Name:     MorphSmile.Name(),
		Offsets:  offsets,
		Strength: 1.0,
	}
}

// generateFrownMorph creates the frown expression morph target.
func generateFrownMorph(mesh *Mesh, skel *Skeleton, layout bodyLayout) MorphTarget {
	offsets := make([]Vec3, len(mesh.Vertices))
	headJoint := skel.Joint(JointHead)

	for i, v := range mesh.Vertices {
		if !isInHeadRegion(v.Position, headJoint.Position, layout) {
			continue
		}

		localPos := vec3Sub(v.Position, headJoint.Position)

		// Frown pulls mouth corners down and furrows brow
		if localPos[1] < 0 && localPos[2] > 0 { // Mouth region
			mouthFactor := gaussianInfluence(localPos[1], -layout.headRY*0.6, 0.03)
			sideFactor := absFloat32(localPos[0]) / layout.headRX

			if sideFactor > 0.3 {
				intensity := mouthFactor * sideFactor * 0.012
				offsets[i] = Vec3{0, -intensity, 0}
			}
		}

		// Furrow brow region
		if localPos[1] > layout.headRY*0.3 && localPos[2] > layout.headRZ*0.5 {
			browFactor := gaussianInfluence(localPos[1], layout.headRY*0.5, 0.02)
			offsets[i] = vec3Add(offsets[i], Vec3{0, -browFactor * 0.008, 0})
		}
	}

	return MorphTarget{
		Type:     MorphFrown,
		Name:     MorphFrown.Name(),
		Offsets:  offsets,
		Strength: 1.0,
	}
}

// generateEyebrowsRaisedMorph creates the raised eyebrows morph target.
func generateEyebrowsRaisedMorph(mesh *Mesh, skel *Skeleton, layout bodyLayout) MorphTarget {
	offsets := make([]Vec3, len(mesh.Vertices))
	headJoint := skel.Joint(JointHead)

	for i, v := range mesh.Vertices {
		if !isInHeadRegion(v.Position, headJoint.Position, layout) {
			continue
		}

		localPos := vec3Sub(v.Position, headJoint.Position)

		// Brow region (upper front of head)
		if localPos[1] > layout.headRY*0.2 && localPos[2] > layout.headRZ*0.4 {
			browFactor := gaussianInfluence(localPos[1], layout.headRY*0.5, 0.025)
			// Raise eyebrows
			offsets[i] = Vec3{0, browFactor * 0.015, browFactor * 0.005}
		}
	}

	return MorphTarget{
		Type:     MorphEyebrowsRaised,
		Name:     MorphEyebrowsRaised.Name(),
		Offsets:  offsets,
		Strength: 1.0,
	}
}

// generateEyesFurrowedMorph creates the furrowed eyebrows morph target.
func generateEyesFurrowedMorph(mesh *Mesh, skel *Skeleton, layout bodyLayout) MorphTarget {
	offsets := make([]Vec3, len(mesh.Vertices))
	headJoint := skel.Joint(JointHead)

	for i, v := range mesh.Vertices {
		if !isInHeadRegion(v.Position, headJoint.Position, layout) {
			continue
		}

		localPos := vec3Sub(v.Position, headJoint.Position)

		// Brow region - pull inward and down
		if localPos[1] > layout.headRY*0.2 && localPos[2] > layout.headRZ*0.4 {
			browFactor := gaussianInfluence(localPos[1], layout.headRY*0.5, 0.025)
			centerDist := absFloat32(localPos[0]) / layout.headRX

			// Inner brow pulls down and in, outer brow less affected
			inwardPull := (1.0 - centerDist) * browFactor * 0.008
			offsets[i] = Vec3{
				-sign(localPos[0]) * inwardPull, // Pull toward center
				-browFactor * 0.01,              // Pull down
				0,
			}
		}
	}

	return MorphTarget{
		Type:     MorphEyebrowsFurrowed,
		Name:     MorphEyebrowsFurrowed.Name(),
		Offsets:  offsets,
		Strength: 1.0,
	}
}

// generateEyesClosedMorph creates the closed eyes morph target.
func generateEyesClosedMorph(mesh *Mesh, skel *Skeleton, layout bodyLayout) MorphTarget {
	offsets := make([]Vec3, len(mesh.Vertices))
	headJoint := skel.Joint(JointHead)

	for i, v := range mesh.Vertices {
		if !isInHeadRegion(v.Position, headJoint.Position, layout) {
			continue
		}

		localPos := vec3Sub(v.Position, headJoint.Position)

		// Eye region (upper portion of face, on sides)
		eyeY := layout.headRY * 0.2
		if localPos[1] > 0 && localPos[1] < layout.headRY*0.5 && localPos[2] > layout.headRZ*0.5 {
			eyeFactor := gaussianInfluence(localPos[1], eyeY, 0.02)
			sideFactor := gaussianInfluence(absFloat32(localPos[0]), layout.headRX*0.4, 0.03)

			// Close eyes by pulling down upper eyelid region
			intensity := eyeFactor * sideFactor * 0.008
			offsets[i] = Vec3{0, -intensity, 0}
		}
	}

	return MorphTarget{
		Type:     MorphEyesClosed,
		Name:     MorphEyesClosed.Name(),
		Offsets:  offsets,
		Strength: 1.0,
	}
}

// generateMouthOpenMorph creates the open mouth morph target.
func generateMouthOpenMorph(mesh *Mesh, skel *Skeleton, layout bodyLayout) MorphTarget {
	offsets := make([]Vec3, len(mesh.Vertices))
	headJoint := skel.Joint(JointHead)

	for i, v := range mesh.Vertices {
		if !isInHeadRegion(v.Position, headJoint.Position, layout) {
			continue
		}

		localPos := vec3Sub(v.Position, headJoint.Position)

		// Lower jaw region
		if localPos[1] < -layout.headRY*0.3 && localPos[2] > layout.headRZ*0.3 {
			jawFactor := gaussianInfluence(localPos[1], -layout.headRY*0.7, 0.04)
			frontFactor := (localPos[2] - layout.headRZ*0.3) / (layout.headRZ * 0.7)
			if frontFactor < 0 {
				frontFactor = 0
			}

			// Pull lower jaw down
			intensity := jawFactor * frontFactor * 0.02
			offsets[i] = Vec3{0, -intensity, 0}
		}
	}

	return MorphTarget{
		Type:     MorphMouthOpen,
		Name:     MorphMouthOpen.Name(),
		Offsets:  offsets,
		Strength: 1.0,
	}
}

// generateJawOpenMorph creates the jaw open morph target with rotation.
func generateJawOpenMorph(mesh *Mesh, skel *Skeleton, layout bodyLayout) MorphTarget {
	offsets := make([]Vec3, len(mesh.Vertices))
	headJoint := skel.Joint(JointHead)

	// Jaw pivot point (just below ear level)
	pivotY := headJoint.Position[1] - layout.headRY*0.5
	pivotZ := headJoint.Position[2] - layout.headRZ*0.3

	for i, v := range mesh.Vertices {
		if !isInHeadRegion(v.Position, headJoint.Position, layout) {
			continue
		}

		localPos := vec3Sub(v.Position, headJoint.Position)

		// Only affect lower jaw (below mouth line)
		if localPos[1] < -layout.headRY*0.4 {
			// Distance from pivot point
			relY := v.Position[1] - pivotY
			relZ := v.Position[2] - pivotZ

			// Rotate around X axis (jaw opens by rotating down/back)
			jawOpenAngle := float32(0.3) // ~17 degrees max
			influence := clampFloat32(-localPos[1]/layout.headRY, 0, 1)
			angle := jawOpenAngle * influence

			// Rotation offset (approximation)
			offsets[i] = Vec3{
				0,
				-relZ*angle - relY*(1-cosApprox(angle)),
				relY * angle,
			}
		}
	}

	return MorphTarget{
		Type:     MorphJawOpen,
		Name:     MorphJawOpen.Name(),
		Offsets:  offsets,
		Strength: 1.0,
	}
}

// ─── Body Morph Generators ───────────────────────────────────────────────────

// generateBreathInMorph creates the inhale breathing morph target.
func generateBreathInMorph(mesh *Mesh, skel *Skeleton, layout bodyLayout) MorphTarget {
	offsets := make([]Vec3, len(mesh.Vertices))

	chestY := (layout.chestTop[1] + layout.chestBottom[1]) / 2

	for i, v := range mesh.Vertices {
		// Only affect chest region
		if v.Position[1] < layout.abdomenBottom[1] || v.Position[1] > layout.neckBottom[1] {
			continue
		}

		// Chest expansion during inhale
		chestFactor := gaussianInfluence(v.Position[1], chestY, 0.15)

		// Expand outward from centerline
		xSign := sign(v.Position[0])
		if xSign == 0 {
			xSign = 1
		}

		expansion := chestFactor * 0.02
		offsets[i] = Vec3{
			xSign * expansion * 0.5, // Lateral expansion
			expansion * 0.3,         // Slight rise
			expansion,               // Forward expansion
		}
	}

	return MorphTarget{
		Type:     MorphBreathIn,
		Name:     MorphBreathIn.Name(),
		Offsets:  offsets,
		Strength: 1.0,
	}
}

// generateBreathOutMorph creates the exhale breathing morph target.
func generateBreathOutMorph(mesh *Mesh, skel *Skeleton, layout bodyLayout) MorphTarget {
	offsets := make([]Vec3, len(mesh.Vertices))

	chestY := (layout.chestTop[1] + layout.chestBottom[1]) / 2

	for i, v := range mesh.Vertices {
		if v.Position[1] < layout.abdomenBottom[1] || v.Position[1] > layout.neckBottom[1] {
			continue
		}

		chestFactor := gaussianInfluence(v.Position[1], chestY, 0.15)
		xSign := sign(v.Position[0])
		if xSign == 0 {
			xSign = 1
		}

		// Contract (negative of breath in, but smaller magnitude)
		contraction := chestFactor * 0.01
		offsets[i] = Vec3{
			-xSign * contraction * 0.5,
			-contraction * 0.2,
			-contraction,
		}
	}

	return MorphTarget{
		Type:     MorphBreathOut,
		Name:     MorphBreathOut.Name(),
		Offsets:  offsets,
		Strength: 1.0,
	}
}

// generateFlexMorph creates the muscle flex morph target.
func generateFlexMorph(mesh *Mesh, skel *Skeleton, layout bodyLayout) MorphTarget {
	offsets := make([]Vec3, len(mesh.Vertices))

	for i, v := range mesh.Vertices {
		// Upper arm region (bicep flex)
		if isInArmRegion(v.Position, layout) {
			armFactor := computeArmFlexFactor(v.Position, layout)
			// Expand bicep region
			offsets[i] = Vec3{
				sign(v.Position[0]) * armFactor * 0.01,
				0,
				armFactor * 0.015,
			}
		}
	}

	return MorphTarget{
		Type:     MorphFlex,
		Name:     MorphFlex.Name(),
		Offsets:  offsets,
		Strength: 1.0,
	}
}

// ─── Helper Functions ────────────────────────────────────────────────────────

// isInHeadRegion checks if a position is within the head region.
func isInHeadRegion(pos, headCenter Vec3, layout bodyLayout) bool {
	localPos := vec3Sub(pos, headCenter)
	// Check if within head bounding ellipsoid
	nx := localPos[0] / layout.headRX
	ny := localPos[1] / layout.headRY
	nz := localPos[2] / layout.headRZ
	return (nx*nx + ny*ny + nz*nz) < 2.0 // Slightly larger than 1 for tolerance
}

// isInArmRegion checks if a position is in the arm region.
func isInArmRegion(pos Vec3, layout bodyLayout) bool {
	absX := absFloat32(pos[0])
	return absX > layout.chestRX && pos[1] > layout.forearmBottomL[1] && pos[1] < layout.upperArmTopL[1]
}

// computeArmFlexFactor calculates flex influence for arm region.
func computeArmFlexFactor(pos Vec3, layout bodyLayout) float32 {
	// Bicep region (upper arm, front)
	upperArmY := (layout.upperArmTopL[1] + layout.upperArmBottomL[1]) / 2
	return gaussianInfluence(pos[1], upperArmY, 0.08)
}

// gaussianInfluence returns a Gaussian falloff based on distance from center.
func gaussianInfluence(value, center, sigma float32) float32 {
	diff := value - center
	return float32(expApprox(-diff * diff / (2 * sigma * sigma)))
}

// expApprox is a fast approximation of e^x for small x.
func expApprox(x float32) float32 {
	// Taylor series approximation for small x
	if x < -4 {
		return 0
	}
	if x > 4 {
		return 50 // Cap large values
	}
	// 1 + x + x²/2 + x³/6
	return 1 + x + x*x*0.5 + x*x*x/6
}

// cosApprox is a fast approximation of cos(x) for small angles.
func cosApprox(x float32) float32 {
	// Taylor series: 1 - x²/2 + x⁴/24
	x2 := x * x
	return 1 - x2*0.5 + x2*x2/24
}

// sign returns -1, 0, or 1 based on the sign of x.
func sign(x float32) float32 {
	if x < 0 {
		return -1
	}
	if x > 0 {
		return 1
	}
	return 0
}

// ─── Apply Morph to Mesh ─────────────────────────────────────────────────────

// ApplyMorphTargetToVertex applies morph target offsets to mesh vertex MorphTarget fields.
// The primary morph target is stored in Vertex.MorphTarget for runtime blending.
func ApplyMorphTargetToVertex(mesh *Mesh, morphSet *MorphTargetSet, targetType MorphTargetType) {
	var target *MorphTarget
	for i := range morphSet.Targets {
		if morphSet.Targets[i].Type == targetType {
			target = &morphSet.Targets[i]
			break
		}
	}
	if target == nil {
		return
	}

	for i := range mesh.Vertices {
		if i < len(target.Offsets) {
			mesh.Vertices[i].MorphTarget = target.Offsets[i]
		}
	}
}

// ─── MorphTargetSet Accessors ────────────────────────────────────────────────

// GetTarget retrieves a specific morph target by type.
func (m *MorphTargetSet) GetTarget(t MorphTargetType) *MorphTarget {
	for i := range m.Targets {
		if m.Targets[i].Type == t {
			return &m.Targets[i]
		}
	}
	return nil
}

// TargetNames returns the names of all morph targets in the set.
func (m *MorphTargetSet) TargetNames() []string {
	names := make([]string, len(m.Targets))
	for i, t := range m.Targets {
		names[i] = t.Name
	}
	return names
}

// ─── Integration with Generator ──────────────────────────────────────────────

// MeshWithMorphs bundles a mesh with its morph targets for blend-shape animation.
type MeshWithMorphs struct {
	Mesh     *Mesh
	Skeleton *Skeleton
	Morphs   *MorphTargetSet
}

// GenerateWithMorphs produces a complete rigged mesh with morph targets.
func (g *Generator) GenerateWithMorphs(p Params) (*MeshWithMorphs, error) {
	if err := p.Validate(); err != nil {
		return nil, err
	}

	rng := newSplitmix64(p.Seed)
	layout := computeBodyLayout(&p, rng)

	// Generate mesh and skeleton
	result, err := g.GenerateWithSkeleton(p)
	if err != nil {
		return nil, err
	}

	// Compute skinning weights
	params := DefaultSkinningParams()
	ComputeSkinningWeights(result.Mesh, result.Skeleton, params)

	// Generate morph targets
	morphs := GenerateMorphTargets(result.Mesh, result.Skeleton, layout)

	return &MeshWithMorphs{
		Mesh:     result.Mesh,
		Skeleton: result.Skeleton,
		Morphs:   morphs,
	}, nil
}
