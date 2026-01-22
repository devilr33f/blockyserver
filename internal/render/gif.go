package render

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/color/palette"
	"image/draw"
	"image/gif"
	"sync"

	"github.com/hytale-tools/blockymodel-merger/pkg/texture"
)

// RenderGIF renders a GLB model to an animated GIF rotating 360 degrees
func RenderGIF(glbBytes []byte, atlas *texture.Atlas, background string, frames, width, height, delay int, dithering, autoZoom bool) ([]byte, error) {
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

	// Render all frames first (in parallel)
	renderedFrames := make([]image.Image, frames)
	var wg sync.WaitGroup
	for i := 0; i < frames; i++ {
		wg.Add(1)
		go func(frameIdx int) {
			defer wg.Done()
			rotation := float64(frameIdx) * rotationPerFrame
			renderedFrames[frameIdx] = RenderScene(mesh, atlasImage, rotation, width, height, bgColor, autoZoom)
		}(i)
	}
	wg.Wait()

	// Determine palette
	var pal color.Palette
	if dithering {
		pal = palette.Plan9
	} else {
		pal = MedianCutQuantize(renderedFrames, 256)
	}

	// Create GIF structure
	g := &gif.GIF{
		Image:     make([]*image.Paletted, frames),
		Delay:     make([]int, frames),
		LoopCount: 0, // 0 = infinite loop
	}

	// Quantize frames to palette (in parallel)
	for i := 0; i < frames; i++ {
		wg.Add(1)
		go func(frameIdx int) {
			defer wg.Done()
			img := renderedFrames[frameIdx]
			paletted := image.NewPaletted(img.Bounds(), pal)
			if dithering {
				draw.FloydSteinberg.Draw(paletted, img.Bounds(), img, image.Point{})
			} else {
				draw.Draw(paletted, img.Bounds(), img, image.Point{}, draw.Src)
			}
			g.Image[frameIdx] = paletted
			g.Delay[frameIdx] = delay
		}(i)
	}
	wg.Wait()

	// Encode GIF
	var buf bytes.Buffer
	if err := gif.EncodeAll(&buf, g); err != nil {
		return nil, fmt.Errorf("encoding GIF: %w", err)
	}

	return buf.Bytes(), nil
}
