// Package unpeople – attachment slot system
//
// This file implements attachment slots for clothing, accessories, and other
// attachable meshes. Slots are positioned relative to the body skeleton and
// provide transform data for attaching external meshes.
package unpeople

// ─── Slot Types ─────────────────────────────────────────────────────────────

// SlotID identifies an attachment slot on the character.
type SlotID int

// Attachment slot IDs for clothing and accessory attachment points.
const (
	SlotHead SlotID = iota
	SlotNeck
	SlotLeftShoulder
	SlotRightShoulder
	SlotChest
	SlotBack
	SlotHips
	SlotLeftWrist
	SlotRightWrist
	SlotLeftAnkle
	SlotRightAnkle
	SlotLeftHand
	SlotRightHand
	SlotCount
)

// SlotNames maps slot IDs to human-readable names.
var SlotNames = map[SlotID]string{
	SlotHead:          "Head",
	SlotNeck:          "Neck",
	SlotLeftShoulder:  "LeftShoulder",
	SlotRightShoulder: "RightShoulder",
	SlotChest:         "Chest",
	SlotBack:          "Back",
	SlotHips:          "Hips",
	SlotLeftWrist:     "LeftWrist",
	SlotRightWrist:    "RightWrist",
	SlotLeftAnkle:     "LeftAnkle",
	SlotRightAnkle:    "RightAnkle",
	SlotLeftHand:      "LeftHand",
	SlotRightHand:     "RightHand",
}

// Name returns the human-readable name of the slot.
func (s SlotID) Name() string {
	if name, ok := SlotNames[s]; ok {
		return name
	}
	return "Unknown"
}

// AttachmentSlot represents a single attachment point on the character mesh.
type AttachmentSlot struct {
	ID       SlotID
	Name     string
	Position Vec3        // World position of attachment point
	Rotation Vec4        // Quaternion rotation (x, y, z, w)
	Scale    Vec3        // Scale factors (usually 1, 1, 1)
	JointID  JointID     // Associated skeleton joint for animation
	Matrix   [16]float32 // Combined transform matrix
}

// AttachmentSlots contains all attachment slots for a character.
type AttachmentSlots struct {
	Slots []AttachmentSlot
}

// ─── Slot Generation ────────────────────────────────────────────────────────

// GenerateAttachmentSlots creates attachment slots based on the body layout.
// The slots are positioned at logical attachment points on the character.
func GenerateAttachmentSlots(layout bodyLayout) *AttachmentSlots {
	slots := &AttachmentSlots{
		Slots: make([]AttachmentSlot, SlotCount),
	}

	// Head slot - top of head for hats, helmets
	slots.Slots[SlotHead] = AttachmentSlot{
		ID:       SlotHead,
		Name:     SlotNames[SlotHead],
		Position: Vec3{layout.headCenter[0], layout.headCenter[1] + layout.headRY, layout.headCenter[2]},
		Rotation: identityQuat(),
		Scale:    Vec3{1, 1, 1},
		JointID:  JointHead,
	}

	// Neck slot - for necklaces, collars
	slots.Slots[SlotNeck] = AttachmentSlot{
		ID:       SlotNeck,
		Name:     SlotNames[SlotNeck],
		Position: layout.neckTop,
		Rotation: identityQuat(),
		Scale:    Vec3{1, 1, 1},
		JointID:  JointNeck,
	}

	// Left shoulder slot - for shoulder pads, pauldrons
	slots.Slots[SlotLeftShoulder] = AttachmentSlot{
		ID:       SlotLeftShoulder,
		Name:     SlotNames[SlotLeftShoulder],
		Position: Vec3{layout.upperArmTopL[0], layout.upperArmTopL[1] + layout.upperArmRadius, layout.upperArmTopL[2]},
		Rotation: identityQuat(),
		Scale:    Vec3{1, 1, 1},
		JointID:  JointLeftShoulder,
	}

	// Right shoulder slot
	slots.Slots[SlotRightShoulder] = AttachmentSlot{
		ID:       SlotRightShoulder,
		Name:     SlotNames[SlotRightShoulder],
		Position: Vec3{layout.upperArmTopR[0], layout.upperArmTopR[1] + layout.upperArmRadius, layout.upperArmTopR[2]},
		Rotation: identityQuat(),
		Scale:    Vec3{1, 1, 1},
		JointID:  JointRightShoulder,
	}

	// Chest slot - for armor, badges
	chestCenterY := (layout.chestTop[1] + layout.chestBottom[1]) / 2
	slots.Slots[SlotChest] = AttachmentSlot{
		ID:       SlotChest,
		Name:     SlotNames[SlotChest],
		Position: Vec3{layout.chestTop[0], chestCenterY, layout.chestTop[2] + layout.chestRZ},
		Rotation: identityQuat(),
		Scale:    Vec3{1, 1, 1},
		JointID:  JointSpine2,
	}

	// Back slot - for backpacks, capes
	slots.Slots[SlotBack] = AttachmentSlot{
		ID:       SlotBack,
		Name:     SlotNames[SlotBack],
		Position: Vec3{layout.chestTop[0], chestCenterY, layout.chestTop[2] - layout.chestRZ},
		Rotation: identityQuat(),
		Scale:    Vec3{1, 1, 1},
		JointID:  JointSpine1,
	}

	// Hips slot - for belts, holsters
	hipsCenterY := (layout.hipsTop[1] + layout.hipsBottom[1]) / 2
	slots.Slots[SlotHips] = AttachmentSlot{
		ID:       SlotHips,
		Name:     SlotNames[SlotHips],
		Position: Vec3{layout.hipsTop[0], hipsCenterY, layout.hipsTop[2]},
		Rotation: identityQuat(),
		Scale:    Vec3{1, 1, 1},
		JointID:  JointHips,
	}

	// Left wrist slot - for bracelets, watches
	slots.Slots[SlotLeftWrist] = AttachmentSlot{
		ID:       SlotLeftWrist,
		Name:     SlotNames[SlotLeftWrist],
		Position: layout.forearmBottomL,
		Rotation: identityQuat(),
		Scale:    Vec3{1, 1, 1},
		JointID:  JointLeftHand,
	}

	// Right wrist slot
	slots.Slots[SlotRightWrist] = AttachmentSlot{
		ID:       SlotRightWrist,
		Name:     SlotNames[SlotRightWrist],
		Position: layout.forearmBottomR,
		Rotation: identityQuat(),
		Scale:    Vec3{1, 1, 1},
		JointID:  JointRightHand,
	}

	// Left ankle slot - for ankle bracelets
	slots.Slots[SlotLeftAnkle] = AttachmentSlot{
		ID:       SlotLeftAnkle,
		Name:     SlotNames[SlotLeftAnkle],
		Position: Vec3{layout.footCenterL[0], layout.footCenterL[1] + layout.footHH, layout.footCenterL[2]},
		Rotation: identityQuat(),
		Scale:    Vec3{1, 1, 1},
		JointID:  JointLeftFoot,
	}

	// Right ankle slot
	slots.Slots[SlotRightAnkle] = AttachmentSlot{
		ID:       SlotRightAnkle,
		Name:     SlotNames[SlotRightAnkle],
		Position: Vec3{layout.footCenterR[0], layout.footCenterR[1] + layout.footHH, layout.footCenterR[2]},
		Rotation: identityQuat(),
		Scale:    Vec3{1, 1, 1},
		JointID:  JointRightFoot,
	}

	// Left hand slot - for weapons, tools
	slots.Slots[SlotLeftHand] = AttachmentSlot{
		ID:       SlotLeftHand,
		Name:     SlotNames[SlotLeftHand],
		Position: layout.handCenterL,
		Rotation: identityQuat(),
		Scale:    Vec3{1, 1, 1},
		JointID:  JointLeftHand,
	}

	// Right hand slot
	slots.Slots[SlotRightHand] = AttachmentSlot{
		ID:       SlotRightHand,
		Name:     SlotNames[SlotRightHand],
		Position: layout.handCenterR,
		Rotation: identityQuat(),
		Scale:    Vec3{1, 1, 1},
		JointID:  JointRightHand,
	}

	// Compute transform matrices for all slots
	for i := range slots.Slots {
		slots.Slots[i].Matrix = computeSlotMatrix(&slots.Slots[i])
	}

	return slots
}

// identityQuat returns an identity quaternion (no rotation).
func identityQuat() Vec4 {
	return Vec4{0, 0, 0, 1}
}

// computeSlotMatrix computes the 4x4 transform matrix for a slot.
func computeSlotMatrix(slot *AttachmentSlot) [16]float32 {
	// Extract quaternion components
	qx, qy, qz, qw := slot.Rotation[0], slot.Rotation[1], slot.Rotation[2], slot.Rotation[3]
	sx, sy, sz := slot.Scale[0], slot.Scale[1], slot.Scale[2]
	tx, ty, tz := slot.Position[0], slot.Position[1], slot.Position[2]

	// Convert quaternion to rotation matrix and combine with scale and translation
	x2, y2, z2 := qx+qx, qy+qy, qz+qz
	xx, xy, xz := qx*x2, qx*y2, qx*z2
	yy, yz, zz := qy*y2, qy*z2, qz*z2
	wx, wy, wz := qw*x2, qw*y2, qw*z2

	return [16]float32{
		(1 - yy - zz) * sx, (xy + wz) * sx, (xz - wy) * sx, 0,
		(xy - wz) * sy, (1 - xx - zz) * sy, (yz + wx) * sy, 0,
		(xz + wy) * sz, (yz - wx) * sz, (1 - xx - yy) * sz, 0,
		tx, ty, tz, 1,
	}
}

// GetSlot returns the attachment slot with the given ID, or nil if not found.
func (slots *AttachmentSlots) GetSlot(id SlotID) *AttachmentSlot {
	if int(id) >= len(slots.Slots) {
		return nil
	}
	return &slots.Slots[id]
}

// GetSlotByName returns the attachment slot with the given name, or nil if not found.
func (slots *AttachmentSlots) GetSlotByName(name string) *AttachmentSlot {
	for i := range slots.Slots {
		if slots.Slots[i].Name == name {
			return &slots.Slots[i]
		}
	}
	return nil
}

// ─── Generator Integration ──────────────────────────────────────────────────

// MeshWithSlots contains a generated mesh along with its attachment slots.
type MeshWithSlots struct {
	Mesh  *Mesh
	Slots *AttachmentSlots
}

// GenerateWithSlots creates a mesh along with attachment slot transforms.
// This is useful for attaching external meshes (clothing, weapons) at predefined points.
func (g *Generator) GenerateWithSlots(p Params) (*MeshWithSlots, error) {
	if err := p.Validate(); err != nil {
		return nil, err
	}

	rng := newSplitmix64(p.Seed)
	layout := computeBodyLayout(&p, rng)

	// Generate mesh using standard path
	mesh, err := g.Generate(p)
	if err != nil {
		return nil, err
	}

	slots := GenerateAttachmentSlots(layout)

	return &MeshWithSlots{
		Mesh:  mesh,
		Slots: slots,
	}, nil
}
