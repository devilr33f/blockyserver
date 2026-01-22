package api

import (
	"encoding/json"
	"net/http"

	"blockyserver/internal/config"
)

// EndpointGuard creates middleware that returns 403 if endpoint is disabled
func EndpointGuard(enabled bool, endpointName string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !enabled {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				json.NewEncoder(w).Encode(ErrorResponse{
					Error: endpointName + " endpoint is disabled",
				})
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// NewEndpointGuards creates guards for all render endpoints based on config
func NewEndpointGuards(cfg *config.EndpointConfig) map[string]func(http.Handler) http.Handler {
	return map[string]func(http.Handler) http.Handler{
		"glb": EndpointGuard(cfg.GLBEnabled, "/render/glb"),
		"png": EndpointGuard(cfg.PNGEnabled, "/render/png"),
		"gif": EndpointGuard(cfg.GIFEnabled, "/render/gif"),
		"mp4": EndpointGuard(cfg.MP4Enabled, "/render/mp4"),
	}
}
