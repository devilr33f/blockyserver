package api

import (
	"encoding/json"
	"io"
	"net/http"

	"blockyserver/internal/render"
	"blockyserver/internal/service"
)

// Handlers contains HTTP handlers for the API
type Handlers struct {
	svc *service.MergeService
}

// NewHandlers creates a new Handlers instance
func NewHandlers(svc *service.MergeService) *Handlers {
	return &Handlers{svc: svc}
}

// HandleGLB handles POST /render/glb
func (h *Handlers) HandleGLB(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, "failed to read request body")
		return
	}
	defer r.Body.Close()

	result, err := h.svc.MergeFromJSON(body)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "model/gltf-binary")
	w.Header().Set("Content-Disposition", "attachment; filename=character.glb")
	w.Write(result.GLBBytes)
}

// HandlePNG handles POST /render/png
func (h *Handlers) HandlePNG(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, "failed to read request body")
		return
	}
	defer r.Body.Close()

	var req PNGRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	req.ApplyDefaults()

	if req.Character == nil {
		writeError(w, http.StatusBadRequest, "character field is required")
		return
	}

	result, err := h.svc.MergeFromJSON(req.Character)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "merge failed: "+err.Error())
		return
	}

	pngBytes, err := render.RenderPNG(result.GLBBytes, result.Atlas, req.Rotation, req.Background, req.Width, req.Height)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "render failed: "+err.Error())
		return
	}

	w.Header().Set("Content-Type", "image/png")
	w.Write(pngBytes)
}

// HandleGIF handles POST /render/gif
func (h *Handlers) HandleGIF(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, "failed to read request body")
		return
	}
	defer r.Body.Close()

	var req GIFRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	req.ApplyDefaults()

	if req.Character == nil {
		writeError(w, http.StatusBadRequest, "character field is required")
		return
	}

	result, err := h.svc.MergeFromJSON(req.Character)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "merge failed: "+err.Error())
		return
	}

	gifBytes, err := render.RenderGIF(result.GLBBytes, result.Atlas, req.Background, req.Frames, req.Width, req.Height, req.Delay, *req.Dithering)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "render failed: "+err.Error())
		return
	}

	w.Header().Set("Content-Type", "image/gif")
	w.Write(gifBytes)
}

// HandleHealth handles GET /health
func (h *Handlers) HandleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"ok"}`))
}

// HandleOpenAPISpec handles GET /openapi.json
func (h *Handlers) HandleOpenAPISpec(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(OpenAPISpec))
}

// HandleSwaggerUI handles GET /docs
func (h *Handlers) HandleSwaggerUI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(SwaggerUIHTML))
}

func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(ErrorResponse{Error: message})
}
