package render

import (
	"image"
	"image/color"
	"sort"
)

// MedianCutQuantize generates an optimal palette for the given images using median-cut algorithm
func MedianCutQuantize(images []image.Image, maxColors int) color.Palette {
	// Collect all unique colors from all images
	colorMap := make(map[uint32]struct{})
	for _, img := range images {
		bounds := img.Bounds()
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				r, g, b, a := img.At(x, y).RGBA()
				if a < 128<<8 {
					continue // skip transparent pixels
				}
				// Pack RGB into uint32 (ignore alpha for palette)
				key := (r>>8)<<16 | (g>>8)<<8 | (b >> 8)
				colorMap[key] = struct{}{}
			}
		}
	}

	// Convert to slice of colors
	colors := make([]rgbColor, 0, len(colorMap))
	for key := range colorMap {
		colors = append(colors, rgbColor{
			r: uint8(key >> 16),
			g: uint8(key >> 8),
			b: uint8(key),
		})
	}

	// If fewer colors than max, just return them all
	if len(colors) <= maxColors {
		palette := make(color.Palette, len(colors))
		for i, c := range colors {
			palette[i] = color.RGBA{c.r, c.g, c.b, 255}
		}
		return palette
	}

	// Perform median-cut
	buckets := medianCut(colors, maxColors)

	// Convert buckets to palette (average color of each bucket)
	palette := make(color.Palette, len(buckets))
	for i, bucket := range buckets {
		palette[i] = bucket.average()
	}

	return palette
}

type rgbColor struct {
	r, g, b uint8
}

type colorBucket []rgbColor

func (b colorBucket) average() color.RGBA {
	if len(b) == 0 {
		return color.RGBA{0, 0, 0, 255}
	}
	var rSum, gSum, bSum int
	for _, c := range b {
		rSum += int(c.r)
		gSum += int(c.g)
		bSum += int(c.b)
	}
	n := len(b)
	return color.RGBA{uint8(rSum / n), uint8(gSum / n), uint8(bSum / n), 255}
}

func (b colorBucket) rangeOfChannel(ch int) int {
	if len(b) == 0 {
		return 0
	}
	min, max := 255, 0
	for _, c := range b {
		var v int
		switch ch {
		case 0:
			v = int(c.r)
		case 1:
			v = int(c.g)
		case 2:
			v = int(c.b)
		}
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}
	return max - min
}

func medianCut(colors []rgbColor, maxBuckets int) []colorBucket {
	if len(colors) == 0 {
		return nil
	}

	buckets := []colorBucket{colors}

	for len(buckets) < maxBuckets {
		// Find bucket with largest range
		maxRange := 0
		maxIdx := 0
		maxCh := 0

		for i, bucket := range buckets {
			if len(bucket) < 2 {
				continue
			}
			for ch := 0; ch < 3; ch++ {
				r := bucket.rangeOfChannel(ch)
				if r > maxRange {
					maxRange = r
					maxIdx = i
					maxCh = ch
				}
			}
		}

		if maxRange == 0 {
			break // can't split further
		}

		// Split the bucket with largest range
		bucket := buckets[maxIdx]
		sort.Slice(bucket, func(i, j int) bool {
			switch maxCh {
			case 0:
				return bucket[i].r < bucket[j].r
			case 1:
				return bucket[i].g < bucket[j].g
			default:
				return bucket[i].b < bucket[j].b
			}
		})

		mid := len(bucket) / 2
		buckets[maxIdx] = bucket[:mid]
		buckets = append(buckets, bucket[mid:])
	}

	return buckets
}
