package render

import (
	"bytes"
	"fmt"
	"image/png"

	"github.com/hytale-tools/blockymodel-merger/pkg/texture"
)

// RenderPNG renders a GLB model to PNG with the given parameters
func RenderPNG(glbBytes []byte, atlas *texture.Atlas, rotation float64, background string, width, height int, autoZoom bool) ([]byte, error) {
	// Parse background color
	bgColor, err := ParseHexColor(background)
	if err != nil {
		return nil, fmt.Errorf("invalid background color: %w", err)
	}

	// Get atlas image
	var atlasImage = atlas.Image

	// Convert GLB to mesh
	mesh, err := GLBToMesh(glbBytes, atlasImage)
	if err != nil {
		return nil, fmt.Errorf("converting GLB to mesh: %w", err)
	}

	// Render the scene
	img := RenderScene(mesh, atlasImage, rotation, width, height, bgColor, autoZoom)

	// Encode to PNG
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, fmt.Errorf("encoding PNG: %w", err)
	}

	return buf.Bytes(), nil
}
