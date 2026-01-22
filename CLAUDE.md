# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Run Commands

```bash
# Build
go build -o blockyserver.exe .

# Run (default port 8080)
./blockyserver.exe
./blockyserver.exe -port 3000

# Dependencies
go mod tidy
```

## Architecture

BlockyServer is an HTTP API for rendering Hytale character models. It wraps the `github.com/hytale-tools/blockymodel-merger` library to provide web endpoints.

### Request Flow

```
HTTP Request → api.Handlers → service.MergeService → blockymodel-merger pkg
                                    ↓
                              render.RenderPNG/GIF (for image endpoints)
                                    ↓
                              HTTP Response (GLB/PNG/GIF)
```

### Package Structure

- **main.go** - Entry point, flag parsing, server startup
- **internal/api/** - HTTP layer (chi router, handlers, OpenAPI spec)
- **internal/service/** - Business logic wrapping blockymodel-merger
- **internal/render/** - Software 3D rendering using fauxgl (GLB→PNG/GIF)

### Key Dependencies

- `github.com/hytale-tools/blockymodel-merger` - Core model merging, texture atlas, GLB export
- `github.com/fogleman/fauxgl` - Software 3D renderer for PNG/GIF output
- `github.com/go-chi/chi/v5` - HTTP router

### Runtime Data

Server requires these directories at runtime (relative to working directory):
- `assets/` - Character models (.blockymodel), textures
- `data/` - JSON registry files (accessories, colors, gradients)

### API Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/render/glb` | POST | Character JSON → GLB binary |
| `/render/png` | POST | Character JSON + options → PNG image |
| `/render/gif` | POST | Character JSON + options → Animated GIF |
| `/docs` | GET | Swagger UI |
| `/openapi.json` | GET | OpenAPI spec |
| `/health` | GET | Health check |

### Character JSON Format

Accessory format: `"AccessoryId.Color.Variant"` (e.g., `"Scavenger_Hair.PitchBlack"`)

Key fields: `bodyCharacteristic`, `haircut`, `eyes`, `pants`, `undertop`, `overtop`, `shoes`
