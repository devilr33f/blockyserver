# BlockyServer

> [!WARNING]
> Not affiliated with Hypixel Studios.
> All trademarks and assets are property of their respective owners.

HTTP API for rendering Hytale character models as GLB, PNG, or animated GIF.

Built on top of [blockymodel-merger](https://github.com/hytale-tools/blockymodel-merger) by [JackGamesFTW](https://github.com/JackGamesFTW), special thanks to him!

## Features

- Merge character accessories into a single model
- Export as GLB (glTF binary)
- Render to PNG with configurable rotation and background
- Render to animated rotating GIF
- Swagger UI documentation

## Requirements

- Go 1.21+
- `assets/` directory with character models and textures
- `data/` directory with JSON registry files

## Obtaining Assets

> [!NOTE]
> You must purchase [Hytale](https://store.hytale.com) to obtain assets. Use code `jack` in the Hytale Store to support Jack's projects.

Download `assets.zip` from the [Hytale Server Manual](https://support.hytale.com/hc/en-us/articles/45326769420827-Hytale-Server-Manual#server-setup).

**Important:** Use the server version, not the client version. The server package includes the `data/` directory with registry JSON files.

### Using extract-assets tool

Clone and build the extraction tool from blockymodel-merger:

```bash
git clone https://github.com/hytale-tools/blockymodel-merger.git
cd blockymodel-merger
go build -o extract-assets ./cmd/extract-assets
./extract-assets /path/to/assets.zip
```

This extracts files into the required structure:
- `Common/Characters` → `assets/Characters/`
- `Common/Cosmetics` → `assets/Cosmetics/`
- `Common/TintGradients` → `assets/TintGradients/`
- `Cosmetics/CharacterCreator` → `data/`

Copy the resulting `assets/` and `data/` directories to your blockyserver folder.

## Installation

```bash
git clone https://github.com/yourusername/blockyserver.git
cd blockyserver
go mod tidy
go build -o blockyserver.exe .
```

## Usage

```bash
# Start server on default port 8080
./blockyserver.exe

# Start on custom port
./blockyserver.exe -port 3000
```

## Docker

### Using Docker Compose (recommended)

```bash
docker compose up -d
```

This mounts `assets/` and `data/` directories as read-only volumes.

### Using Docker directly

```bash
# Build image
docker build -t blockyserver .

# Run container
docker run -d -p 8080:8080 \
  -v $(pwd)/assets:/app/assets:ro \
  -v $(pwd)/data:/app/data:ro \
  blockyserver
```

## API Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/render/glb` | POST | Returns GLB binary |
| `/render/png` | POST | Returns PNG image |
| `/render/gif` | POST | Returns animated GIF |
| `/docs` | GET | Swagger UI |
| `/openapi.json` | GET | OpenAPI specification |
| `/health` | GET | Health check |

## Example Request

### Render PNG

```bash
curl -X POST http://localhost:8080/render/png \
  -H "Content-Type: application/json" \
  -d '{
    "character": {
      "bodyCharacteristic": "Default.02",
      "haircut": "Scavenger_Hair.PitchBlack",
      "eyes": "Large_Eyes.Pink",
      "pants": "Pants_A.Blue",
      "undertop": "Shirt_A.White"
    },
    "rotation": 45,
    "background": "transparent",
    "width": 512,
    "height": 512
  }' --output character.png
```
