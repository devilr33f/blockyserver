package render

import (
	"fmt"
	"image/png"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	"github.com/hytale-tools/blockymodel-merger/pkg/texture"
)

// RenderMP4 renders a GLB model to an MP4 video rotating 360 degrees
func RenderMP4(glbBytes []byte, atlas *texture.Atlas, background string, frames, width, height, fps int, autoZoom bool) ([]byte, error) {
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

	// Create temp directory for frames
	tempDir, err := os.MkdirTemp("", "blockyserver-mp4-*")
	if err != nil {
		return nil, fmt.Errorf("creating temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Calculate rotation per frame
	rotationPerFrame := 360.0 / float64(frames)

	// Render all frames in parallel
	var wg sync.WaitGroup
	errChan := make(chan error, frames)

	for i := 0; i < frames; i++ {
		wg.Add(1)
		go func(frameIdx int) {
			defer wg.Done()
			rotation := float64(frameIdx) * rotationPerFrame
			img := RenderScene(mesh, atlasImage, rotation, width, height, bgColor, autoZoom)

			// Write frame to temp file
			framePath := filepath.Join(tempDir, fmt.Sprintf("frame_%04d.png", frameIdx))
			f, err := os.Create(framePath)
			if err != nil {
				errChan <- fmt.Errorf("creating frame file: %w", err)
				return
			}
			defer f.Close()

			if err := png.Encode(f, img); err != nil {
				errChan <- fmt.Errorf("encoding frame PNG: %w", err)
				return
			}
		}(i)
	}
	wg.Wait()
	close(errChan)

	// Check for errors
	for err := range errChan {
		if err != nil {
			return nil, err
		}
	}

	// Run FFmpeg to encode MP4
	outputPath := filepath.Join(tempDir, "output.mp4")
	inputPattern := filepath.Join(tempDir, "frame_%04d.png")

	cmd := exec.Command("ffmpeg",
		"-y",
		"-framerate", fmt.Sprintf("%d", fps),
		"-i", inputPattern,
		"-c:v", "libx264",
		"-pix_fmt", "yuv420p",
		"-movflags", "+faststart",
		outputPath,
	)

	if output, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("ffmpeg encoding failed: %w\nOutput: %s", err, string(output))
	}

	// Read output file
	mp4Bytes, err := os.ReadFile(outputPath)
	if err != nil {
		return nil, fmt.Errorf("reading output MP4: %w", err)
	}

	return mp4Bytes, nil
}
