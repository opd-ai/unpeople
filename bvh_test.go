package unpeople_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/opd-ai/unpeople"
)

// ─── BVH Parser Tests ────────────────────────────────────────────────────────

// Sample BVH file content for testing
// Total channels: 6 (Hips) + 3*9 (Spine, Spine1, Neck, Head, LeftUpLeg, LeftLowLeg, LeftFoot, RightUpLeg, RightLowLeg, RightFoot) = 6 + 30 = 36
const testBVHContent = `HIERARCHY
ROOT Hips
{
	OFFSET 0.00 0.00 0.00
	CHANNELS 6 Xposition Yposition Zposition Zrotation Xrotation Yrotation
	JOINT Spine
	{
		OFFSET 0.00 10.00 0.00
		CHANNELS 3 Zrotation Xrotation Yrotation
		JOINT Spine1
		{
			OFFSET 0.00 10.00 0.00
			CHANNELS 3 Zrotation Xrotation Yrotation
			JOINT Neck
			{
				OFFSET 0.00 10.00 0.00
				CHANNELS 3 Zrotation Xrotation Yrotation
				JOINT Head
				{
					OFFSET 0.00 5.00 0.00
					CHANNELS 3 Zrotation Xrotation Yrotation
					End Site
					{
						OFFSET 0.00 5.00 0.00
					}
				}
			}
		}
	}
	JOINT LeftUpLeg
	{
		OFFSET 5.00 0.00 0.00
		CHANNELS 3 Zrotation Xrotation Yrotation
		JOINT LeftLowLeg
		{
			OFFSET 0.00 -20.00 0.00
			CHANNELS 3 Zrotation Xrotation Yrotation
			JOINT LeftFoot
			{
				OFFSET 0.00 -20.00 0.00
				CHANNELS 3 Zrotation Xrotation Yrotation
				End Site
				{
					OFFSET 0.00 -5.00 5.00
				}
			}
		}
	}
	JOINT RightUpLeg
	{
		OFFSET -5.00 0.00 0.00
		CHANNELS 3 Zrotation Xrotation Yrotation
		JOINT RightLowLeg
		{
			OFFSET 0.00 -20.00 0.00
			CHANNELS 3 Zrotation Xrotation Yrotation
			JOINT RightFoot
			{
				OFFSET 0.00 -20.00 0.00
				CHANNELS 3 Zrotation Xrotation Yrotation
				End Site
				{
					OFFSET 0.00 -5.00 5.00
				}
			}
		}
	}
}
MOTION
Frames: 3
Frame Time: 0.033333
0.0 0.0 0.0 0.0 0.0 0.0 0.0 0.0 0.0 0.0 0.0 0.0 0.0 0.0 0.0 0.0 0.0 0.0 0.0 0.0 0.0 0.0 0.0 0.0 0.0 0.0 0.0 0.0 0.0 0.0 0.0 0.0 0.0 0.0 0.0 0.0
0.0 5.0 0.0 5.0 0.0 0.0 0.0 0.0 0.0 0.0 0.0 0.0 0.0 0.0 0.0 0.0 0.0 0.0 10.0 0.0 0.0 20.0 0.0 0.0 0.0 0.0 0.0 -10.0 0.0 0.0 -20.0 0.0 0.0 0.0 0.0 0.0
0.0 10.0 0.0 10.0 0.0 0.0 0.0 0.0 0.0 0.0 0.0 0.0 0.0 0.0 0.0 0.0 0.0 0.0 20.0 0.0 0.0 40.0 0.0 0.0 0.0 0.0 0.0 -20.0 0.0 0.0 -40.0 0.0 0.0 0.0 0.0 0.0
`

func TestParseBVH(t *testing.T) {
	bvh, err := unpeople.ParseBVH(strings.NewReader(testBVHContent))
	if err != nil {
		t.Fatalf("ParseBVH failed: %v", err)
	}

	// Verify hierarchy
	if bvh.Root == nil {
		t.Fatal("Root joint is nil")
	}
	if bvh.Root.Name != "Hips" {
		t.Errorf("Root name = %q, want Hips", bvh.Root.Name)
	}

	// Verify channel count (6 root + 3*10 other joints = 36)
	totalChannels := bvh.TotalChannels()
	if totalChannels != 36 {
		t.Errorf("TotalChannels = %d, want 36", totalChannels)
	}

	// Verify frame data
	if bvh.FrameCount != 3 {
		t.Errorf("FrameCount = %d, want 3", bvh.FrameCount)
	}
	if bvh.FrameTime < 0.033 || bvh.FrameTime > 0.034 {
		t.Errorf("FrameTime = %f, want ~0.033333", bvh.FrameTime)
	}
	if len(bvh.Frames) != 3 {
		t.Errorf("len(Frames) = %d, want 3", len(bvh.Frames))
	}

	// Verify frame channel values match expected count
	for i, frame := range bvh.Frames {
		if len(frame.ChannelValues) != 36 {
			t.Errorf("Frame %d: len(ChannelValues) = %d, want 36", i, len(frame.ChannelValues))
		}
	}
}

func TestBVHJointMapping(t *testing.T) {
	tests := []struct {
		bvhName string
		wantID  unpeople.JointID
	}{
		{"Hips", unpeople.JointHips},
		{"hip", unpeople.JointHips},
		{"Spine", unpeople.JointSpine},
		{"Neck", unpeople.JointNeck},
		{"Head", unpeople.JointHead},
		{"LeftUpLeg", unpeople.JointLeftUpperLeg},
		{"LeftLowLeg", unpeople.JointLeftLowerLeg},
		{"LeftFoot", unpeople.JointLeftFoot},
		{"RightUpperArm", unpeople.JointRightUpperArm},
		{"RightForeArm", unpeople.JointRightForearm},
		{"UnknownJoint", unpeople.JointID(-1)},
	}

	for _, tt := range tests {
		t.Run(tt.bvhName, func(t *testing.T) {
			got := unpeople.MapBVHJoint(tt.bvhName)
			if got != tt.wantID {
				t.Errorf("MapBVHJoint(%q) = %d, want %d", tt.bvhName, got, tt.wantID)
			}
		})
	}
}

func TestBVHToAnimation(t *testing.T) {
	bvh, err := unpeople.ParseBVH(strings.NewReader(testBVHContent))
	if err != nil {
		t.Fatalf("ParseBVH failed: %v", err)
	}

	anim := unpeople.BVHToAnimation(bvh, "TestWalk")

	if anim.Name != "TestWalk" {
		t.Errorf("Animation name = %q, want TestWalk", anim.Name)
	}
	if anim.FrameCount != 3 {
		t.Errorf("Animation FrameCount = %d, want 3", anim.FrameCount)
	}
	if len(anim.JointFrames) == 0 {
		t.Fatal("Animation has no joint frames")
	}

	// Check that mapped joints have animation data
	hasHips := false
	for _, jf := range anim.JointFrames {
		if jf.JointID == unpeople.JointHips {
			hasHips = true
			// Hips should have translations (root joint)
			if len(jf.Translations) != 3 {
				t.Errorf("Hips translations = %d frames, want 3", len(jf.Translations))
			}
			// Hips should have rotations
			if len(jf.Rotations) != 3 {
				t.Errorf("Hips rotations = %d frames, want 3", len(jf.Rotations))
			}
		}
	}
	if !hasHips {
		t.Error("Hips joint not found in animation")
	}
}

func TestBVHDuration(t *testing.T) {
	bvh, err := unpeople.ParseBVH(strings.NewReader(testBVHContent))
	if err != nil {
		t.Fatalf("ParseBVH failed: %v", err)
	}

	duration := bvh.Duration()
	expectedDuration := float32(3) * float32(0.033333)
	if duration < expectedDuration-0.001 || duration > expectedDuration+0.001 {
		t.Errorf("Duration = %f, want ~%f", duration, expectedDuration)
	}
}

func TestBVHGetJointNames(t *testing.T) {
	bvh, err := unpeople.ParseBVH(strings.NewReader(testBVHContent))
	if err != nil {
		t.Fatalf("ParseBVH failed: %v", err)
	}

	names := bvh.GetJointNames()
	if len(names) == 0 {
		t.Fatal("GetJointNames returned empty list")
	}

	// First should be Hips (root)
	if names[0] != "Hips" {
		t.Errorf("First joint name = %q, want Hips", names[0])
	}

	// Should contain key joints
	expectedJoints := []string{"Hips", "Spine", "Spine1", "Neck", "Head", "LeftUpLeg", "RightUpLeg"}
	for _, expected := range expectedJoints {
		found := false
		for _, name := range names {
			if name == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected joint %q not found in names", expected)
		}
	}
}

func TestBVHGetMappedJoints(t *testing.T) {
	bvh, err := unpeople.ParseBVH(strings.NewReader(testBVHContent))
	if err != nil {
		t.Fatalf("ParseBVH failed: %v", err)
	}

	mapped := bvh.GetMappedJoints()
	if len(mapped) == 0 {
		t.Fatal("GetMappedJoints returned empty list")
	}

	// All returned joints should have valid mappings
	for _, j := range mapped {
		id := unpeople.MapBVHJoint(j.Name)
		if id == unpeople.JointID(-1) {
			t.Errorf("Joint %q is in mapped list but has no mapping", j.Name)
		}
	}
}

func TestParseBVHInvalidHierarchy(t *testing.T) {
	invalidContent := `NOT_HIERARCHY
ROOT Hips
{
}
`
	_, err := unpeople.ParseBVH(strings.NewReader(invalidContent))
	if err == nil {
		t.Error("Expected error for invalid hierarchy keyword")
	}
}

func TestParseBVHMissingMotion(t *testing.T) {
	noMotionContent := `HIERARCHY
ROOT Hips
{
	OFFSET 0.00 0.00 0.00
	CHANNELS 6 Xposition Yposition Zposition Zrotation Xrotation Yrotation
}
`
	_, err := unpeople.ParseBVH(strings.NewReader(noMotionContent))
	if err == nil {
		t.Error("Expected error for missing motion section")
	}
}

func TestAnimationDuration(t *testing.T) {
	bvh, err := unpeople.ParseBVH(strings.NewReader(testBVHContent))
	if err != nil {
		t.Fatalf("ParseBVH failed: %v", err)
	}

	anim := unpeople.BVHToAnimation(bvh, "Test")
	duration := anim.Duration()
	expectedDuration := float32(3) * float32(0.033333)
	if duration < expectedDuration-0.001 || duration > expectedDuration+0.001 {
		t.Errorf("Animation.Duration = %f, want ~%f", duration, expectedDuration)
	}
}

// Minimal BVH for edge case testing
const minimalBVH = `HIERARCHY
ROOT Hips
{
	OFFSET 0 0 0
	CHANNELS 3 Zrotation Xrotation Yrotation
	End Site
	{
		OFFSET 0 10 0
	}
}
MOTION
Frames: 1
Frame Time: 0.1
0 0 0
`

func TestParseBVHMinimal(t *testing.T) {
	bvh, err := unpeople.ParseBVH(strings.NewReader(minimalBVH))
	if err != nil {
		t.Fatalf("ParseBVH failed: %v", err)
	}

	if bvh.Root == nil {
		t.Fatal("Root is nil")
	}
	if bvh.Root.Name != "Hips" {
		t.Errorf("Root name = %q, want Hips", bvh.Root.Name)
	}
	if bvh.FrameCount != 1 {
		t.Errorf("FrameCount = %d, want 1", bvh.FrameCount)
	}
	if bvh.TotalChannels() != 3 {
		t.Errorf("TotalChannels = %d, want 3", bvh.TotalChannels())
	}
}

func TestParseBVHEndSite(t *testing.T) {
	bvh, err := unpeople.ParseBVH(strings.NewReader(minimalBVH))
	if err != nil {
		t.Fatalf("ParseBVH failed: %v", err)
	}

	// Root should have one child (the End Site)
	if len(bvh.Root.Children) != 1 {
		t.Errorf("Root children = %d, want 1", len(bvh.Root.Children))
	}
	endSite := bvh.Root.Children[0]
	if !endSite.IsEndSite {
		t.Error("Child should be an End Site")
	}
}

// Test channel parsing variations
func TestParseBVHChannelTypes(t *testing.T) {
	channelBVH := `HIERARCHY
ROOT Test
{
	OFFSET 0 0 0
	CHANNELS 6 Xposition Yposition Zposition Xrotation Yrotation Zrotation
}
MOTION
Frames: 1
Frame Time: 0.1
1.0 2.0 3.0 45.0 90.0 180.0
`
	bvh, err := unpeople.ParseBVH(strings.NewReader(channelBVH))
	if err != nil {
		t.Fatalf("ParseBVH failed: %v", err)
	}

	if len(bvh.Root.Channels) != 6 {
		t.Errorf("Channel count = %d, want 6", len(bvh.Root.Channels))
	}

	// Verify frame values are correct
	if bvh.Frames[0].ChannelValues[0] != 1.0 {
		t.Errorf("Channel[0] = %f, want 1.0", bvh.Frames[0].ChannelValues[0])
	}
	if bvh.Frames[0].ChannelValues[3] != 45.0 {
		t.Errorf("Channel[3] = %f, want 45.0", bvh.Frames[0].ChannelValues[3])
	}
}

// ─── Animated Export Tests ───────────────────────────────────────────────────

func TestGenerateAnimated(t *testing.T) {
	gen := unpeople.NewGenerator()
	params := unpeople.DefaultParams()
	params.Seed = 42

	result, err := gen.GenerateAnimated(params, nil)
	if err != nil {
		t.Fatalf("GenerateAnimated failed: %v", err)
	}

	if result.Mesh == nil {
		t.Fatal("Mesh is nil")
	}
	if result.Skeleton == nil {
		t.Fatal("Skeleton is nil")
	}
	if result.Animation != nil {
		t.Error("Animation should be nil when no BVH provided")
	}

	// Verify mesh has skinning data
	foundNonZeroWeights := false
	for _, v := range result.Mesh.Vertices {
		for i := 0; i < 4; i++ {
			if v.JointWeights[i] > 0 {
				foundNonZeroWeights = true
				break
			}
		}
		if foundNonZeroWeights {
			break
		}
	}
	if !foundNonZeroWeights {
		t.Error("Expected non-zero skinning weights")
	}
}

func TestGenerateAnimatedWithBVH(t *testing.T) {
	gen := unpeople.NewGenerator()
	params := unpeople.DefaultParams()
	params.Seed = 42

	bvh, err := unpeople.ParseBVH(strings.NewReader(testBVHContent))
	if err != nil {
		t.Fatalf("ParseBVH failed: %v", err)
	}

	result, err := gen.GenerateAnimated(params, bvh)
	if err != nil {
		t.Fatalf("GenerateAnimated failed: %v", err)
	}

	if result.Animation == nil {
		t.Fatal("Animation should not be nil when BVH provided")
	}
	if result.Animation.FrameCount != 3 {
		t.Errorf("Animation FrameCount = %d, want 3", result.Animation.FrameCount)
	}
}

func TestExportAnimatedGLTF(t *testing.T) {
	gen := unpeople.NewGenerator()
	params := unpeople.DefaultParams()
	params.Seed = 42

	bvh, err := unpeople.ParseBVH(strings.NewReader(testBVHContent))
	if err != nil {
		t.Fatalf("ParseBVH failed: %v", err)
	}

	result, err := gen.GenerateAnimated(params, bvh)
	if err != nil {
		t.Fatalf("GenerateAnimated failed: %v", err)
	}

	var buf bytes.Buffer
	opts := unpeople.DefaultAnimatedGLTFOptions()
	err = unpeople.ExportAnimatedGLTF(&buf, result, opts)
	if err != nil {
		t.Fatalf("ExportAnimatedGLTF failed: %v", err)
	}

	// Verify output contains expected glTF elements
	output := buf.String()
	if !strings.Contains(output, `"asset"`) {
		t.Error("Output missing asset field")
	}
	if !strings.Contains(output, `"skins"`) {
		t.Error("Output missing skins field")
	}
	if !strings.Contains(output, `"animations"`) {
		t.Error("Output missing animations field")
	}
}

func TestExportAnimatedGLB(t *testing.T) {
	gen := unpeople.NewGenerator()
	params := unpeople.DefaultParams()
	params.Seed = 42

	bvh, err := unpeople.ParseBVH(strings.NewReader(testBVHContent))
	if err != nil {
		t.Fatalf("ParseBVH failed: %v", err)
	}

	result, err := gen.GenerateAnimated(params, bvh)
	if err != nil {
		t.Fatalf("GenerateAnimated failed: %v", err)
	}

	var buf bytes.Buffer
	opts := unpeople.DefaultAnimatedGLTFOptions()
	err = unpeople.ExportAnimatedGLB(&buf, result, opts)
	if err != nil {
		t.Fatalf("ExportAnimatedGLB failed: %v", err)
	}

	// Verify GLB magic number
	data := buf.Bytes()
	if len(data) < 12 {
		t.Fatal("GLB output too small")
	}
	magic := string(data[0:4])
	if magic != "glTF" {
		t.Errorf("GLB magic = %q, want glTF", magic)
	}
}
