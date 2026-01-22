package config

import (
	"os"
	"strings"
)

// EndpointConfig holds enable/disable flags for render endpoints
type EndpointConfig struct {
	GLBEnabled bool
	PNGEnabled bool
	GIFEnabled bool
	MP4Enabled bool
}

// LoadEndpointConfig reads endpoint configuration from environment variables.
// All endpoints are enabled by default.
// Set BLOCKY_DISABLE_GLB=true, BLOCKY_DISABLE_PNG=true, etc. to disable.
func LoadEndpointConfig() *EndpointConfig {
	return &EndpointConfig{
		GLBEnabled: !isDisabled("BLOCKY_DISABLE_GLB"),
		PNGEnabled: !isDisabled("BLOCKY_DISABLE_PNG"),
		GIFEnabled: !isDisabled("BLOCKY_DISABLE_GIF"),
		MP4Enabled: !isDisabled("BLOCKY_DISABLE_MP4"),
	}
}

func isDisabled(envVar string) bool {
	val := strings.ToLower(os.Getenv(envVar))
	return val == "true" || val == "1" || val == "yes"
}
