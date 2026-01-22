package api

import "encoding/json"

// PNGRequest represents a request to render a character as PNG
type PNGRequest struct {
	Character  json.RawMessage `json:"character"`
	Rotation   float64         `json:"rotation"`   // degrees, default 0
	Background string          `json:"background"` // "transparent" or hex "#RRGGBB"
	Width      int             `json:"width"`      // default 512
	Height     int             `json:"height"`     // default 512
}

// GIFRequest represents a request to render a character as animated GIF
type GIFRequest struct {
	Character  json.RawMessage `json:"character"`
	Background string          `json:"background"` // hex color "#RRGGBB"
	Frames     int             `json:"frames"`     // default 36 (10° per frame)
	Width      int             `json:"width"`      // default 512
	Height     int             `json:"height"`     // default 512
	Delay      int             `json:"delay"`      // centiseconds between frames, default 5
	Dithering  *bool           `json:"dithering"`  // Floyd-Steinberg dithering, default true
}

// ErrorResponse represents an error returned by the API
type ErrorResponse struct {
	Error string `json:"error"`
}

// ApplyDefaults fills in default values for PNGRequest
func (r *PNGRequest) ApplyDefaults() {
	if r.Width == 0 {
		r.Width = 512
	}
	if r.Height == 0 {
		r.Height = 512
	}
	if r.Background == "" {
		r.Background = "transparent"
	}
}

// ApplyDefaults fills in default values for GIFRequest
func (r *GIFRequest) ApplyDefaults() {
	if r.Width == 0 {
		r.Width = 512
	}
	if r.Height == 0 {
		r.Height = 512
	}
	if r.Frames == 0 {
		r.Frames = 36
	}
	if r.Delay == 0 {
		r.Delay = 5
	}
	if r.Background == "" {
		r.Background = "#FFFFFF"
	}
	if r.Dithering == nil {
		defaultDithering := true
		r.Dithering = &defaultDithering
	}
}
