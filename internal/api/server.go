package api

import (
	"net/http"
	"time"

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

	// Create handlers
	h := NewHandlers(svc)

	// Routes
	r.Get("/health", h.HandleHealth)
	r.Get("/openapi.json", h.HandleOpenAPISpec)
	r.Get("/docs", h.HandleSwaggerUI)
	r.Post("/render/glb", h.HandleGLB)
	r.Post("/render/png", h.HandlePNG)
	r.Post("/render/gif", h.HandleGIF)

	return r
}
