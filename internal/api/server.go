package api

import (
	"net/http"
	"time"

	"blockyserver/internal/config"
	"blockyserver/internal/service"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// NewServer creates a new HTTP server with all routes configured
func NewServer(svc *service.MergeService) http.Handler {
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// Load endpoint config
	cfg := config.LoadEndpointConfig()
	guards := NewEndpointGuards(cfg)

	// Create handlers
	h := NewHandlers(svc)

	// Routes
	r.Get("/health", h.HandleHealth)
	r.Get("/openapi.json", h.HandleOpenAPISpec)
	r.Get("/docs", h.HandleSwaggerUI)
	r.With(guards["glb"]).Post("/render/glb", h.HandleGLB)
	r.With(guards["png"]).Post("/render/png", h.HandlePNG)
	r.With(guards["gif"]).Post("/render/gif", h.HandleGIF)
	r.With(guards["mp4"]).Post("/render/mp4", h.HandleMP4)

	return r
}
