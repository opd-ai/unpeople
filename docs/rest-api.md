# REST API Documentation

The unpeople-server provides an HTTP API for generating humanoid character meshes.
This enables integration with non-Go game engines like Unity, Unreal, Godot, and
web applications.

## Starting the Server

```bash
# Build and run on default port 8080
go build ./cmd/unpeople-server
./unpeople-server

# Run on a custom port
./unpeople-server -addr :9000

# Bind to all interfaces
./unpeople-server -addr 0.0.0.0:8080
```

## Endpoints

### Health Check

```
GET /health
```

Returns the server status.

**Response:**
```json
{
  "status": "ok",
  "time": "2024-01-15T10:30:00Z"
}
```

### Generate Mesh

```
POST /generate
```

Generates a humanoid mesh from the given parameters.

**Request Headers:**
- `Content-Type: application/json` (required)
- `Accept: model/gltf+json | model/gltf-binary | text/plain` (optional)

**Request Body:**
```json
{
  "seed": 42,
  "species": 0,
  "height": 2,
  "build": 2,
  "proportions": 1,
  "phenotype": 0,
  "age": 3,
  "posture": 0
}
```

**Query Parameters:**
- `format=obj|gltf|glb` — Override Accept header format selection

**Response Formats:**

| Format | MIME Type | Description |
|--------|-----------|-------------|
| gltf (default) | `model/gltf+json` | glTF 2.0 JSON with embedded buffers |
| glb | `model/gltf-binary` | glTF 2.0 Binary (single file) |
| obj | `text/plain` | Wavefront OBJ text format |

## Parameter Reference

### Species Values
| Value | Name |
|-------|------|
| 0 | Human |
| 1 | Elf |
| 2 | Dwarf |
| 3 | Gnome |
| 4 | Halfling |
| 5 | Goblin |
| 6 | Kobold |
| 7 | Orc |
| 8 | Troll |
| 9 | Ogre |

### Height Values
| Value | Name |
|-------|------|
| 0 | Giant |
| 1 | Tall |
| 2 | Medium |
| 3 | Short |
| 4 | Tiny |

### Build Values
| Value | Name |
|-------|------|
| 0 | Muscular |
| 1 | Athletic |
| 2 | Average |
| 3 | Lean |
| 4 | Stocky |
| 5 | Fragile |

### Proportions Values
| Value | Name |
|-------|------|
| 0 | Heroic |
| 1 | Realistic |
| 2 | Stylized |
| 3 | Caricature |

### Phenotype Values
| Value | Name |
|-------|------|
| 0 | Masculine |
| 1 | Androgynous |
| 2 | Feminine |

### Age Values
| Value | Name |
|-------|------|
| 0 | Decrepit |
| 1 | Elderly |
| 2 | Old |
| 3 | Adult |
| 4 | Youth |
| 5 | Teen |
| 6 | Child |
| 7 | Toddler |

### Posture Values
| Value | Name |
|-------|------|
| 0 | Upright |
| 1 | Slouched |
| 2 | Hunched |
| 3 | Rigid |

## Examples

### Generate a default human (cURL)

```bash
curl -X POST http://localhost:8080/generate \
  -H "Content-Type: application/json" \
  -d '{"seed": 42}' \
  -o character.gltf
```

### Generate a tall elf as GLB

```bash
curl -X POST "http://localhost:8080/generate?format=glb" \
  -H "Content-Type: application/json" \
  -d '{"seed": 100, "species": 1, "height": 1}' \
  -o elf.glb
```

### Generate OBJ for debugging

```bash
curl -X POST http://localhost:8080/generate \
  -H "Content-Type: application/json" \
  -H "Accept: text/plain" \
  -d '{"seed": 42}' \
  -o debug.obj
```

### JavaScript (fetch)

```javascript
const params = {
  seed: Date.now(),
  species: 0,  // Human
  height: 2,   // Medium
  build: 2,    // Average
  age: 3       // Adult
};

const response = await fetch('http://localhost:8080/generate', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
    'Accept': 'model/gltf-binary'
  },
  body: JSON.stringify(params)
});

const glbData = await response.arrayBuffer();
// Load into Three.js, Babylon.js, etc.
```

### Unity (C#)

```csharp
using UnityEngine;
using UnityEngine.Networking;
using System.Collections;

public class UnpeopleClient : MonoBehaviour
{
    [System.Serializable]
    public class CharacterParams
    {
        public long seed;
        public int species;
        public int height;
        public int build;
        public int age;
    }

    IEnumerator GenerateCharacter()
    {
        var p = new CharacterParams { seed = 42, species = 0, height = 2, build = 2, age = 3 };
        string json = JsonUtility.ToJson(p);
        
        using (var request = new UnityWebRequest("http://localhost:8080/generate?format=glb", "POST"))
        {
            request.uploadHandler = new UploadHandlerRaw(System.Text.Encoding.UTF8.GetBytes(json));
            request.downloadHandler = new DownloadHandlerBuffer();
            request.SetRequestHeader("Content-Type", "application/json");
            
            yield return request.SendWebRequest();
            
            if (request.result == UnityWebRequest.Result.Success)
            {
                byte[] glbData = request.downloadHandler.data;
                // Load GLB into Unity scene
            }
        }
    }
}
```

## Error Responses

| Status | Description |
|--------|-------------|
| 400 | Invalid JSON or parameters |
| 405 | Method not allowed (use POST for /generate) |
| 429 | Rate limit exceeded |
| 500 | Internal server error |

**Error format:**
```json
{"error": "Invalid parameters: seed must be positive"}
```

## Rate Limiting

The server implements token bucket rate limiting:
- Default: 100 requests per second
- When exceeded, returns HTTP 429

## CORS Support

The server includes CORS headers for browser-based clients:
- `Access-Control-Allow-Origin: *`
- `Access-Control-Allow-Methods: GET, POST, OPTIONS`
- `Access-Control-Allow-Headers: Content-Type, Accept`

## Performance

- Typical generation time: <10ms
- Recommended timeout: 5 seconds (allows for complex characters)
- Output size: ~50KB (GLB), ~100KB (glTF with embedded buffers), ~200KB (OBJ)
