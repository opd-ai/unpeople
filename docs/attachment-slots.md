# Attachment Slots

The attachment slot system provides predefined points on the character mesh where external meshes (clothing, weapons, accessories) can be attached. Each slot provides position, rotation, scale transforms along with the associated skeleton joint for animation.

## Overview

Attachment slots are generated alongside the mesh using `GenerateWithSlots()`:

```go
g := unpeople.NewGenerator()
p := unpeople.DefaultParams()

result, err := g.GenerateWithSlots(p)
if err != nil {
    // handle error
}

// Access mesh and slots
mesh := result.Mesh
slots := result.Slots
```

## Available Slots

The system defines 13 attachment slots covering common attachment points:

| Slot ID | Name | Description | Typical Use Cases |
|---------|------|-------------|-------------------|
| `SlotHead` | Head | Top of head | Hats, helmets, crowns |
| `SlotNeck` | Neck | Base of neck | Necklaces, collars, scarves |
| `SlotLeftShoulder` | LeftShoulder | Left shoulder | Pauldrons, shoulder pads |
| `SlotRightShoulder` | RightShoulder | Right shoulder | Pauldrons, shoulder pads |
| `SlotChest` | Chest | Front of chest | Armor plates, badges, brooches |
| `SlotBack` | Back | Upper back | Backpacks, capes, wings |
| `SlotHips` | Hips | Hip level | Belts, holsters, pouches |
| `SlotLeftWrist` | LeftWrist | Left wrist | Bracelets, watches, gauntlets |
| `SlotRightWrist` | RightWrist | Right wrist | Bracelets, watches, gauntlets |
| `SlotLeftAnkle` | LeftAnkle | Left ankle | Ankle bracelets, leg armor |
| `SlotRightAnkle` | RightAnkle | Right ankle | Ankle bracelets, leg armor |
| `SlotLeftHand` | LeftHand | Left hand center | Weapons, tools, shields |
| `SlotRightHand` | RightHand | Right hand center | Weapons, tools, shields |

## Slot Data Structure

Each slot provides:

```go
type AttachmentSlot struct {
    ID       SlotID          // Unique slot identifier
    Name     string          // Human-readable name
    Position Vec3            // World position (x, y, z)
    Rotation Vec4            // Quaternion rotation (x, y, z, w)
    Scale    Vec3            // Scale factors (usually 1, 1, 1)
    JointID  JointID         // Associated skeleton joint
    Matrix   [16]float32     // Combined 4x4 transform matrix
}
```

## Accessing Slots

### By ID

```go
headSlot := slots.GetSlot(unpeople.SlotHead)
if headSlot != nil {
    fmt.Printf("Head position: %v\n", headSlot.Position)
}
```

### By Name

```go
slot := slots.GetSlotByName("LeftShoulder")
if slot != nil {
    fmt.Printf("Left shoulder at: %v\n", slot.Position)
}
```

### Iterating All Slots

```go
for _, slot := range slots.Slots {
    fmt.Printf("%s: position=%v, joint=%d\n", 
        slot.Name, slot.Position, slot.JointID)
}
```

## glTF Export

Slots can be exported as glTF nodes for use in 3D applications:

```go
opts := unpeople.DefaultGLTFOptions()
opts.IncludeSlots = true
opts.Slots = result.Slots

var buf bytes.Buffer
err := unpeople.ExportGLTF(&buf, result.Mesh, opts)
```

In the exported glTF:
- Each slot becomes a separate node named `Slot_<SlotName>` (e.g., `Slot_LeftShoulder`)
- Nodes include translation, rotation, and scale transforms
- Slots are siblings of the mesh node in the scene hierarchy

## Animation Integration

Each slot has an associated `JointID` that links to the skeleton. When animating the character:

1. Apply animation transforms to skeleton joints
2. Compute the animated world transform for each slot's joint
3. Use the slot's local transform relative to the joint

For skeletal animation, multiply the slot's transform matrix by the animated joint's world matrix.

## Coordinate System

- Y-up, right-handed coordinate system
- Units are in meters
- Character faces +Z direction
- Left is -X, right is +X

## Example: Attaching a Weapon

```go
// Generate character with slots
g := unpeople.NewGenerator()
p := unpeople.DefaultParams()
result, _ := g.GenerateWithSlots(p)

// Get right hand slot
handSlot := result.Slots.GetSlot(unpeople.SlotRightHand)

// Use the slot's transform matrix to position a weapon mesh
// In your rendering engine:
weaponTransform := handSlot.Matrix  // 4x4 matrix in column-major order
```

## Slot Positioning

Slots are positioned relative to the body layout computed from character parameters:

- **Head**: Top of head ellipsoid
- **Shoulders**: Above upper arm origin
- **Chest/Back**: Center of chest, offset forward/backward by chest depth
- **Hips**: Center of hip region
- **Wrists**: End of forearm (start of hand)
- **Ankles**: Above foot center
- **Hands**: Palm center

Slot positions automatically scale with character height, build, and other parameters.
