package render

import (
	"bytes"
	"fmt"
	"image"
	"image/color/palette"
	"image/draw"
	"image/gif"

	"github.com/hytale-tools/blockymodel-merger/pkg/texture"
)

// RenderGIF renders a GLB model to an animated GIF rotating 360 degrees
func RenderGIF(glbBytes []byte, atlas *texture.Atlas, background string, frames, width, height, delay int) ([]byte, error) {
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

	// Calculate rotation per frame
	rotationPerFrame := 360.0 / float64(frames)

	// Create GIF structure
	g := &gif.GIF{
		Image:     make([]*image.Paletted, frames),
		Delay:     make([]int, frames),
		LoopCount: 0, // 0 = infinite loop
	}

	// Render each frame
	for i := 0; i < frames; i++ {
		rotation := float64(i) * rotationPerFrame

		// Render frame
		img := RenderScene(mesh, atlasImage, rotation, width, height, bgColor)

		// Quantize to palette
		paletted := image.NewPaletted(img.Bounds(), palette.Plan9)
		draw.FloydSteinberg.Draw(paletted, img.Bounds(), img, image.Point{})

		g.Image[i] = paletted
		g.Delay[i] = delay
	}

	// Encode GIF
	var buf bytes.Buffer
	if err := gif.EncodeAll(&buf, g); err != nil {
		return nil, fmt.Errorf("encoding GIF: %w", err)
	}

	return buf.Bytes(), nil
}
