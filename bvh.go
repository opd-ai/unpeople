// Package unpeople – BVH motion capture file import
//
// This file implements parsing of Biovision Hierarchy (BVH) files, the standard
// format for motion capture data. BVH files contain a skeleton hierarchy
// description followed by frame-by-frame animation data.
//
// The parser maps BVH joint names to unpeople skeleton joints, enabling
// motion capture data to be applied to generated characters.
package unpeople

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"
)

// ─── BVH Types ───────────────────────────────────────────────────────────────

// BVHChannelType represents the type of channel in a BVH file.
type BVHChannelType int

const (
	// Position channels
	BVHChannelXPosition BVHChannelType = iota
	BVHChannelYPosition
	BVHChannelZPosition
	// Rotation channels
	BVHChannelXRotation
	BVHChannelYRotation
	BVHChannelZRotation
)

// BVHJoint represents a joint in the BVH skeleton hierarchy.
type BVHJoint struct {
	Name      string
	Offset    Vec3
	Channels  []BVHChannelType
	Children  []*BVHJoint
	Parent    *BVHJoint
	IsEndSite bool
}

// BVHFile represents a parsed BVH motion capture file.
type BVHFile struct {
	Root       *BVHJoint
	FrameCount int
	FrameTime  float32
	Frames     []BVHFrame

	// Computed during parsing
	channelCount int
	jointList    []*BVHJoint // Flattened in parse order (for channel mapping)
}

// BVHFrame represents a single frame of animation data.
type BVHFrame struct {
	// ChannelValues contains all channel values in order matching the hierarchy.
	ChannelValues []float32
}

// ─── BVH Parsing ─────────────────────────────────────────────────────────────

// ParseBVH parses a BVH file from the given reader.
func ParseBVH(r io.Reader) (*BVHFile, error) {
	scanner := bufio.NewScanner(r)
	bvh := &BVHFile{}

	if err := parseHierarchy(scanner, bvh); err != nil {
		return nil, fmt.Errorf("BVH hierarchy parse: %w", err)
	}

	if err := parseMotion(scanner, bvh); err != nil {
		return nil, fmt.Errorf("BVH motion parse: %w", err)
	}

	return bvh, nil
}

// scanToHierarchyKeyword scans until the HIERARCHY keyword is found.
func scanToHierarchyKeyword(scanner *bufio.Scanner) error {
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		if line == "HIERARCHY" {
			return nil
		}
		return fmt.Errorf("expected HIERARCHY, got %q", line)
	}
	return scanner.Err()
}

// parseRootJoint parses the ROOT joint and initializes the BVH structure.
func parseRootJoint(scanner *bufio.Scanner, bvh *BVHFile) error {
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) >= 2 && fields[0] == "ROOT" {
			root, err := parseJoint(scanner, fields[1], nil)
			if err != nil {
				return err
			}
			bvh.Root = root
			bvh.channelCount = countChannels(root)
			flattenJoints(root, &bvh.jointList)
			return nil
		}
		return fmt.Errorf("expected ROOT, got %q", line)
	}
	return errors.New("unexpected end of file in hierarchy")
}

// parseHierarchy parses the HIERARCHY section of a BVH file.
func parseHierarchy(scanner *bufio.Scanner, bvh *BVHFile) error {
	if err := scanToHierarchyKeyword(scanner); err != nil {
		return err
	}
	return parseRootJoint(scanner, bvh)
}

// expectOpenBrace scans for and validates an opening brace.
func expectOpenBrace(scanner *bufio.Scanner, context string) error {
	if !scanNonEmpty(scanner) {
		return fmt.Errorf("unexpected end of file after %s", context)
	}
	if strings.TrimSpace(scanner.Text()) != "{" {
		return fmt.Errorf("expected '{', got %q", scanner.Text())
	}
	return nil
}

// parseJointOffset parses an OFFSET line into a Vec3.
func parseJointOffset(fields []string) (Vec3, error) {
	if len(fields) < 4 {
		return Vec3{}, fmt.Errorf("OFFSET requires 3 values, got %d", len(fields)-1)
	}
	offset, err := parseVec3(fields[1:4])
	if err != nil {
		return Vec3{}, fmt.Errorf("OFFSET parse: %w", err)
	}
	return offset, nil
}

// handleJointKeyword processes a single keyword within a joint block.
// Returns (done, error) where done=true signals the block is complete.
func handleJointKeyword(scanner *bufio.Scanner, joint *BVHJoint, fields []string) (bool, error) {
	switch fields[0] {
	case "OFFSET":
		offset, err := parseJointOffset(fields)
		if err != nil {
			return false, err
		}
		joint.Offset = offset

	case "CHANNELS":
		channels, err := parseChannels(fields)
		if err != nil {
			return false, err
		}
		joint.Channels = channels

	case "JOINT":
		if len(fields) < 2 {
			return false, errors.New("JOINT requires a name")
		}
		child, err := parseJoint(scanner, fields[1], joint)
		if err != nil {
			return false, err
		}
		joint.Children = append(joint.Children, child)

	case "End":
		if len(fields) >= 2 && fields[1] == "Site" {
			endSite, err := parseEndSite(scanner, joint)
			if err != nil {
				return false, err
			}
			joint.Children = append(joint.Children, endSite)
		}

	case "}":
		return true, nil
	}
	return false, nil
}

// parseJoint recursively parses a joint and its children.
func parseJoint(scanner *bufio.Scanner, name string, parent *BVHJoint) (*BVHJoint, error) {
	joint := &BVHJoint{
		Name:   name,
		Parent: parent,
	}

	if err := expectOpenBrace(scanner, "joint name"); err != nil {
		return nil, err
	}

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}

		done, err := handleJointKeyword(scanner, joint, fields)
		if err != nil {
			return nil, err
		}
		if done {
			return joint, nil
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return nil, errors.New("unexpected end of file in joint")
}

// parseEndSite parses an End Site block.
func parseEndSite(scanner *bufio.Scanner, parent *BVHJoint) (*BVHJoint, error) {
	endSite := &BVHJoint{
		Name:      parent.Name + "_End",
		Parent:    parent,
		IsEndSite: true,
	}

	if err := expectOpenBrace(scanner, "End Site"); err != nil {
		return nil, err
	}

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}

		switch fields[0] {
		case "OFFSET":
			offset, err := parseJointOffset(fields)
			if err != nil {
				return nil, err
			}
			endSite.Offset = offset
		case "}":
			return endSite, nil
		}
	}

	return nil, errors.New("unexpected end of file in End Site")
}

// parseChannels parses the CHANNELS line.
func parseChannels(fields []string) ([]BVHChannelType, error) {
	if len(fields) < 2 {
		return nil, errors.New("CHANNELS requires count")
	}
	count, err := strconv.Atoi(fields[1])
	if err != nil {
		return nil, fmt.Errorf("invalid channel count: %w", err)
	}
	if len(fields) < 2+count {
		return nil, fmt.Errorf("CHANNELS declared %d but only %d provided", count, len(fields)-2)
	}

	channels := make([]BVHChannelType, count)
	for i := 0; i < count; i++ {
		ch, err := parseChannelType(fields[2+i])
		if err != nil {
			return nil, err
		}
		channels[i] = ch
	}
	return channels, nil
}

// parseChannelType converts a channel name to BVHChannelType.
func parseChannelType(name string) (BVHChannelType, error) {
	switch name {
	case "Xposition":
		return BVHChannelXPosition, nil
	case "Yposition":
		return BVHChannelYPosition, nil
	case "Zposition":
		return BVHChannelZPosition, nil
	case "Xrotation":
		return BVHChannelXRotation, nil
	case "Yrotation":
		return BVHChannelYRotation, nil
	case "Zrotation":
		return BVHChannelZRotation, nil
	default:
		return 0, fmt.Errorf("unknown channel type: %s", name)
	}
}

// scanToMotionKeyword scans until the MOTION keyword is found.
func scanToMotionKeyword(scanner *bufio.Scanner) error {
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		if line == "MOTION" {
			return nil
		}
	}
	return scanner.Err()
}

// parseFrameCount parses the "Frames:" line and returns the count.
func parseFrameCount(scanner *bufio.Scanner) (int, error) {
	if !scanNonEmpty(scanner) {
		return 0, errors.New("missing Frames line")
	}
	framesLine := strings.TrimSpace(scanner.Text())
	if !strings.HasPrefix(framesLine, "Frames:") {
		return 0, fmt.Errorf("expected 'Frames:', got %q", framesLine)
	}
	count, err := strconv.Atoi(strings.TrimSpace(strings.TrimPrefix(framesLine, "Frames:")))
	if err != nil {
		return 0, fmt.Errorf("invalid frame count: %w", err)
	}
	return count, nil
}

// parseFrameTime parses the "Frame Time:" line and returns the value.
func parseFrameTime(scanner *bufio.Scanner) (float32, error) {
	if !scanNonEmpty(scanner) {
		return 0, errors.New("missing Frame Time line")
	}
	frameTimeLine := strings.TrimSpace(scanner.Text())
	if !strings.HasPrefix(frameTimeLine, "Frame Time:") {
		return 0, fmt.Errorf("expected 'Frame Time:', got %q", frameTimeLine)
	}
	frameTime, err := strconv.ParseFloat(strings.TrimSpace(strings.TrimPrefix(frameTimeLine, "Frame Time:")), 32)
	if err != nil {
		return 0, fmt.Errorf("invalid frame time: %w", err)
	}
	return float32(frameTime), nil
}

// parseAllFrames parses all frame data lines from the BVH file.
func parseAllFrames(scanner *bufio.Scanner, frameCount, channelCount int) ([]BVHFrame, error) {
	frames := make([]BVHFrame, 0, frameCount)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		frame, err := parseFrameData(line, channelCount)
		if err != nil {
			return nil, fmt.Errorf("frame %d: %w", len(frames)+1, err)
		}
		frames = append(frames, frame)
		if len(frames) >= frameCount {
			break
		}
	}
	if len(frames) < frameCount {
		return nil, fmt.Errorf("expected %d frames, got %d", frameCount, len(frames))
	}
	return frames, nil
}

// parseMotion parses the MOTION section of a BVH file.
func parseMotion(scanner *bufio.Scanner, bvh *BVHFile) error {
	if err := scanToMotionKeyword(scanner); err != nil {
		return err
	}

	frameCount, err := parseFrameCount(scanner)
	if err != nil {
		return err
	}
	bvh.FrameCount = frameCount

	frameTime, err := parseFrameTime(scanner)
	if err != nil {
		return err
	}
	bvh.FrameTime = frameTime

	frames, err := parseAllFrames(scanner, frameCount, bvh.channelCount)
	if err != nil {
		return err
	}
	bvh.Frames = frames

	return scanner.Err()
}

// parseFrameData parses a single line of frame data.
func parseFrameData(line string, expectedCount int) (BVHFrame, error) {
	fields := strings.Fields(line)
	if len(fields) != expectedCount {
		return BVHFrame{}, fmt.Errorf("expected %d values, got %d", expectedCount, len(fields))
	}

	values := make([]float32, expectedCount)
	for i, field := range fields {
		val, err := strconv.ParseFloat(field, 32)
		if err != nil {
			return BVHFrame{}, fmt.Errorf("value %d: %w", i, err)
		}
		values[i] = float32(val)
	}

	return BVHFrame{ChannelValues: values}, nil
}

// ─── Helper Functions ────────────────────────────────────────────────────────

// scanNonEmpty scans until a non-empty line is found.
func scanNonEmpty(scanner *bufio.Scanner) bool {
	for scanner.Scan() {
		if strings.TrimSpace(scanner.Text()) != "" {
			return true
		}
	}
	return false
}

// parseVec3 parses three float values into a Vec3.
func parseVec3(fields []string) (Vec3, error) {
	var v Vec3
	for i := 0; i < 3; i++ {
		val, err := strconv.ParseFloat(fields[i], 32)
		if err != nil {
			return Vec3{}, err
		}
		v[i] = float32(val)
	}
	return v, nil
}

// countChannels recursively counts total channels in the hierarchy.
func countChannels(joint *BVHJoint) int {
	count := len(joint.Channels)
	for _, child := range joint.Children {
		count += countChannels(child)
	}
	return count
}

// flattenJoints creates a flat list of joints in parse order (for channel mapping).
func flattenJoints(joint *BVHJoint, list *[]*BVHJoint) {
	if !joint.IsEndSite {
		*list = append(*list, joint)
	}
	for _, child := range joint.Children {
		flattenJoints(child, list)
	}
}

// ─── Joint Name Mapping ──────────────────────────────────────────────────────

// BVHJointMapping maps BVH joint names to unpeople JointIDs.
// This handles common naming conventions in BVH files from various sources.
var BVHJointMapping = map[string]JointID{
	// Root / Hips variations
	"Hips":   JointHips,
	"hip":    JointHips,
	"HIP":    JointHips,
	"Pelvis": JointHips,
	"pelvis": JointHips,
	"Root":   JointRoot,
	"root":   JointRoot,

	// Spine variations
	"Spine":      JointSpine,
	"spine":      JointSpine,
	"Spine1":     JointSpine1,
	"spine1":     JointSpine1,
	"Spine2":     JointSpine2,
	"spine2":     JointSpine2,
	"Spine3":     JointSpine2,
	"Chest":      JointSpine2,
	"chest":      JointSpine2,
	"UpperChest": JointSpine2,

	// Neck / Head variations
	"Neck":  JointNeck,
	"neck":  JointNeck,
	"Neck1": JointNeck,
	"Head":  JointHead,
	"head":  JointHead,

	// Left arm variations
	"LeftShoulder": JointLeftShoulder,
	"leftShoulder": JointLeftShoulder,
	"LeftCollar":   JointLeftShoulder,
	"lShoulder":    JointLeftShoulder,
	"L_Shoulder":   JointLeftShoulder,
	"LeftUpArm":    JointLeftUpperArm,
	"LeftUpperArm": JointLeftUpperArm,
	"lUpperArm":    JointLeftUpperArm,
	"L_UpperArm":   JointLeftUpperArm,
	"LeftArm":      JointLeftUpperArm,
	"LeftForeArm":  JointLeftForearm,
	"LeftLowArm":   JointLeftForearm,
	"lForeArm":     JointLeftForearm,
	"L_ForeArm":    JointLeftForearm,
	"LeftForearm":  JointLeftForearm,
	"LeftHand":     JointLeftHand,
	"leftHand":     JointLeftHand,
	"lHand":        JointLeftHand,
	"L_Hand":       JointLeftHand,

	// Right arm variations
	"RightShoulder": JointRightShoulder,
	"rightShoulder": JointRightShoulder,
	"RightCollar":   JointRightShoulder,
	"rShoulder":     JointRightShoulder,
	"R_Shoulder":    JointRightShoulder,
	"RightUpArm":    JointRightUpperArm,
	"RightUpperArm": JointRightUpperArm,
	"rUpperArm":     JointRightUpperArm,
	"R_UpperArm":    JointRightUpperArm,
	"RightArm":      JointRightUpperArm,
	"RightForeArm":  JointRightForearm,
	"RightLowArm":   JointRightForearm,
	"rForeArm":      JointRightForearm,
	"R_ForeArm":     JointRightForearm,
	"RightForearm":  JointRightForearm,
	"RightHand":     JointRightHand,
	"rightHand":     JointRightHand,
	"rHand":         JointRightHand,
	"R_Hand":        JointRightHand,

	// Left leg variations
	"LeftUpLeg":    JointLeftUpperLeg,
	"LeftUpperLeg": JointLeftUpperLeg,
	"LeftHip":      JointLeftUpperLeg,
	"lThigh":       JointLeftUpperLeg,
	"L_UpperLeg":   JointLeftUpperLeg,
	"LeftLeg":      JointLeftUpperLeg,
	"LeftLowLeg":   JointLeftLowerLeg,
	"LeftLowerLeg": JointLeftLowerLeg,
	"lShin":        JointLeftLowerLeg,
	"L_LowerLeg":   JointLeftLowerLeg,
	"LeftKnee":     JointLeftLowerLeg,
	"LeftFoot":     JointLeftFoot,
	"leftFoot":     JointLeftFoot,
	"lFoot":        JointLeftFoot,
	"L_Foot":       JointLeftFoot,
	"LeftToe":      JointLeftToe,
	"LeftToeBase":  JointLeftToe,
	"lToe":         JointLeftToe,
	"L_Toe":        JointLeftToe,

	// Right leg variations
	"RightUpLeg":    JointRightUpperLeg,
	"RightUpperLeg": JointRightUpperLeg,
	"RightHip":      JointRightUpperLeg,
	"rThigh":        JointRightUpperLeg,
	"R_UpperLeg":    JointRightUpperLeg,
	"RightLeg":      JointRightUpperLeg,
	"RightLowLeg":   JointRightLowerLeg,
	"RightLowerLeg": JointRightLowerLeg,
	"rShin":         JointRightLowerLeg,
	"R_LowerLeg":    JointRightLowerLeg,
	"RightKnee":     JointRightLowerLeg,
	"RightFoot":     JointRightFoot,
	"rightFoot":     JointRightFoot,
	"rFoot":         JointRightFoot,
	"R_Foot":        JointRightFoot,
	"RightToe":      JointRightToe,
	"RightToeBase":  JointRightToe,
	"rToe":          JointRightToe,
	"R_Toe":         JointRightToe,

	// Finger names (common variations)
	"LeftHandThumb1":  JointLeftThumb1,
	"LeftThumb1":      JointLeftThumb1,
	"lThumb1":         JointLeftThumb1,
	"LeftHandThumb2":  JointLeftThumb2,
	"LeftThumb2":      JointLeftThumb2,
	"lThumb2":         JointLeftThumb2,
	"LeftHandIndex1":  JointLeftIndex1,
	"LeftIndex1":      JointLeftIndex1,
	"lIndex1":         JointLeftIndex1,
	"LeftHandIndex2":  JointLeftIndex2,
	"LeftIndex2":      JointLeftIndex2,
	"lIndex2":         JointLeftIndex2,
	"LeftHandIndex3":  JointLeftIndex3,
	"LeftIndex3":      JointLeftIndex3,
	"lIndex3":         JointLeftIndex3,
	"LeftHandMiddle1": JointLeftMiddle1,
	"LeftMiddle1":     JointLeftMiddle1,
	"lMiddle1":        JointLeftMiddle1,
	"LeftHandMiddle2": JointLeftMiddle2,
	"LeftMiddle2":     JointLeftMiddle2,
	"lMiddle2":        JointLeftMiddle2,
	"LeftHandMiddle3": JointLeftMiddle3,
	"LeftMiddle3":     JointLeftMiddle3,
	"lMiddle3":        JointLeftMiddle3,
	"LeftHandRing1":   JointLeftRing1,
	"LeftRing1":       JointLeftRing1,
	"lRing1":          JointLeftRing1,
	"LeftHandRing2":   JointLeftRing2,
	"LeftRing2":       JointLeftRing2,
	"lRing2":          JointLeftRing2,
	"LeftHandRing3":   JointLeftRing3,
	"LeftRing3":       JointLeftRing3,
	"lRing3":          JointLeftRing3,
	"LeftHandPinky1":  JointLeftPinky1,
	"LeftPinky1":      JointLeftPinky1,
	"lPinky1":         JointLeftPinky1,
	"LeftHandPinky2":  JointLeftPinky2,
	"LeftPinky2":      JointLeftPinky2,
	"lPinky2":         JointLeftPinky2,
	"LeftHandPinky3":  JointLeftPinky3,
	"LeftPinky3":      JointLeftPinky3,
	"lPinky3":         JointLeftPinky3,

	// Right hand fingers
	"RightHandThumb1":  JointRightThumb1,
	"RightThumb1":      JointRightThumb1,
	"rThumb1":          JointRightThumb1,
	"RightHandThumb2":  JointRightThumb2,
	"RightThumb2":      JointRightThumb2,
	"rThumb2":          JointRightThumb2,
	"RightHandIndex1":  JointRightIndex1,
	"RightIndex1":      JointRightIndex1,
	"rIndex1":          JointRightIndex1,
	"RightHandIndex2":  JointRightIndex2,
	"RightIndex2":      JointRightIndex2,
	"rIndex2":          JointRightIndex2,
	"RightHandIndex3":  JointRightIndex3,
	"RightIndex3":      JointRightIndex3,
	"rIndex3":          JointRightIndex3,
	"RightHandMiddle1": JointRightMiddle1,
	"RightMiddle1":     JointRightMiddle1,
	"rMiddle1":         JointRightMiddle1,
	"RightHandMiddle2": JointRightMiddle2,
	"RightMiddle2":     JointRightMiddle2,
	"rMiddle2":         JointRightMiddle2,
	"RightHandMiddle3": JointRightMiddle3,
	"RightMiddle3":     JointRightMiddle3,
	"rMiddle3":         JointRightMiddle3,
	"RightHandRing1":   JointRightRing1,
	"RightRing1":       JointRightRing1,
	"rRing1":           JointRightRing1,
	"RightHandRing2":   JointRightRing2,
	"RightRing2":       JointRightRing2,
	"rRing2":           JointRightRing2,
	"RightHandRing3":   JointRightRing3,
	"RightRing3":       JointRightRing3,
	"rRing3":           JointRightRing3,
	"RightHandPinky1":  JointRightPinky1,
	"RightPinky1":      JointRightPinky1,
	"rPinky1":          JointRightPinky1,
	"RightHandPinky2":  JointRightPinky2,
	"RightPinky2":      JointRightPinky2,
	"rPinky2":          JointRightPinky2,
	"RightHandPinky3":  JointRightPinky3,
	"RightPinky3":      JointRightPinky3,
	"rPinky3":          JointRightPinky3,
}

// MapBVHJoint maps a BVH joint name to an unpeople JointID.
// Returns JointID(-1) if no mapping is found.
func MapBVHJoint(name string) JointID {
	if id, ok := BVHJointMapping[name]; ok {
		return id
	}
	return JointID(-1)
}

// ─── Animation Types ─────────────────────────────────────────────────────────

// Animation represents a sequence of animation frames for a skeleton.
type Animation struct {
	Name        string
	FrameCount  int
	FrameTime   float32 // Seconds per frame
	JointFrames []JointAnimationData
}

// JointAnimationData contains per-frame transforms for a single joint.
type JointAnimationData struct {
	JointID      JointID
	Translations []Vec3 // Position per frame (root only typically)
	Rotations    []Vec4 // Quaternion per frame
}

// Duration returns the total animation duration in seconds.
func (a *Animation) Duration() float32 {
	return float32(a.FrameCount) * a.FrameTime
}

// ─── BVH to Animation Conversion ─────────────────────────────────────────────

// BVHToAnimation converts parsed BVH data to an Animation for the unpeople skeleton.
func BVHToAnimation(bvh *BVHFile, name string) *Animation {
	anim := &Animation{
		Name:       name,
		FrameCount: bvh.FrameCount,
		FrameTime:  bvh.FrameTime,
	}

	// Build mapping from BVH joints to unpeople joints
	mappings := buildJointMappings(bvh)

	// Convert each mapped joint's channels to animation data
	for _, m := range mappings {
		jad := extractJointAnimationData(bvh, m)
		if jad != nil {
			anim.JointFrames = append(anim.JointFrames, *jad)
		}
	}

	return anim
}

// jointMapping holds the mapping from BVH joint to unpeople joint.
type jointMapping struct {
	bvhJoint   *BVHJoint
	unpeopleID JointID
	channelIdx int // Starting index in frame channel data
}

// buildJointMappings creates mappings from BVH to unpeople joints.
func buildJointMappings(bvh *BVHFile) []jointMapping {
	var mappings []jointMapping
	channelIdx := 0

	for _, joint := range bvh.jointList {
		unpeopleID := MapBVHJoint(joint.Name)
		if unpeopleID != JointID(-1) {
			mappings = append(mappings, jointMapping{
				bvhJoint:   joint,
				unpeopleID: unpeopleID,
				channelIdx: channelIdx,
			})
		}
		channelIdx += len(joint.Channels)
	}

	return mappings
}

// channelLayout holds the parsed channel information for a joint.
type channelLayout struct {
	posChans    [3]int // Index of X,Y,Z position channels
	rotChans    [3]int // Index of X,Y,Z rotation channels
	hasPosition bool
	hasRotation bool
}

// parseChannelLayout determines which channels are present and their positions.
func parseChannelLayout(channels []BVHChannelType) channelLayout {
	var layout channelLayout
	for i, ch := range channels {
		switch ch {
		case BVHChannelXPosition:
			layout.posChans[0] = i
			layout.hasPosition = true
		case BVHChannelYPosition:
			layout.posChans[1] = i
			layout.hasPosition = true
		case BVHChannelZPosition:
			layout.posChans[2] = i
			layout.hasPosition = true
		case BVHChannelXRotation:
			layout.rotChans[0] = i
			layout.hasRotation = true
		case BVHChannelYRotation:
			layout.rotChans[1] = i
			layout.hasRotation = true
		case BVHChannelZRotation:
			layout.rotChans[2] = i
			layout.hasRotation = true
		}
	}
	return layout
}

// extractTranslations extracts position data from frames for a joint.
func extractTranslations(bvh *BVHFile, channelIdx int, posChans [3]int) []Vec3 {
	translations := make([]Vec3, bvh.FrameCount)
	for f := 0; f < bvh.FrameCount; f++ {
		translations[f] = Vec3{
			bvh.Frames[f].ChannelValues[channelIdx+posChans[0]],
			bvh.Frames[f].ChannelValues[channelIdx+posChans[1]],
			bvh.Frames[f].ChannelValues[channelIdx+posChans[2]],
		}
	}
	return translations
}

// extractRotations extracts rotation data from frames for a joint.
func extractRotations(bvh *BVHFile, channelIdx int, rotChans [3]int) []Vec4 {
	rotations := make([]Vec4, bvh.FrameCount)
	for f := 0; f < bvh.FrameCount; f++ {
		angles := Vec3{
			bvh.Frames[f].ChannelValues[channelIdx+rotChans[0]],
			bvh.Frames[f].ChannelValues[channelIdx+rotChans[1]],
			bvh.Frames[f].ChannelValues[channelIdx+rotChans[2]],
		}
		rotations[f] = eulerToQuaternionZXY(angles)
	}
	return rotations
}

// extractJointAnimationData extracts animation data for a single joint.
func extractJointAnimationData(bvh *BVHFile, m jointMapping) *JointAnimationData {
	layout := parseChannelLayout(m.bvhJoint.Channels)

	if !layout.hasPosition && !layout.hasRotation {
		return nil
	}

	jad := &JointAnimationData{JointID: m.unpeopleID}

	if layout.hasPosition {
		jad.Translations = extractTranslations(bvh, m.channelIdx, layout.posChans)
	}
	if layout.hasRotation {
		jad.Rotations = extractRotations(bvh, m.channelIdx, layout.rotChans)
	}

	return jad
}

// eulerToQuaternionZXY converts Euler angles (degrees, ZXY order) to a quaternion.
// ZXY is the most common rotation order in BVH files.
func eulerToQuaternionZXY(angles Vec3) Vec4 {
	// Convert degrees to radians
	rx := float64(angles[0]) * math.Pi / 180.0
	ry := float64(angles[1]) * math.Pi / 180.0
	rz := float64(angles[2]) * math.Pi / 180.0

	// Half angles
	cx := math.Cos(rx / 2)
	sx := math.Sin(rx / 2)
	cy := math.Cos(ry / 2)
	sy := math.Sin(ry / 2)
	cz := math.Cos(rz / 2)
	sz := math.Sin(rz / 2)

	// ZXY rotation order: first Z, then X, then Y
	// Combined quaternion multiplication
	return Vec4{
		float32(sx*cy*cz - cx*sy*sz),
		float32(cx*sy*cz + sx*cy*sz),
		float32(cx*cy*sz - sx*sy*cz),
		float32(cx*cy*cz + sx*sy*sz),
	}
}

// ─── Accessors ───────────────────────────────────────────────────────────────

// JointCount returns the number of joints in the BVH skeleton.
func (b *BVHFile) JointCount() int {
	return len(b.jointList)
}

// TotalChannels returns the total number of channels across all joints.
func (b *BVHFile) TotalChannels() int {
	return b.channelCount
}

// Duration returns the total animation duration in seconds.
func (b *BVHFile) Duration() float32 {
	return float32(b.FrameCount) * b.FrameTime
}

// GetJointNames returns the names of all joints in the BVH hierarchy.
func (b *BVHFile) GetJointNames() []string {
	names := make([]string, len(b.jointList))
	for i, j := range b.jointList {
		names[i] = j.Name
	}
	return names
}

// GetMappedJoints returns only the BVH joints that map to unpeople joints.
func (b *BVHFile) GetMappedJoints() []*BVHJoint {
	var mapped []*BVHJoint
	for _, j := range b.jointList {
		if MapBVHJoint(j.Name) != JointID(-1) {
			mapped = append(mapped, j)
		}
	}
	return mapped
}
