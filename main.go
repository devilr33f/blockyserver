package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"blockyserver/internal/api"
	"blockyserver/internal/service"
)

func main() {
	port := flag.Int("port", 8080, "Port to listen on")
	flag.Parse()

	log.Println("Loading merge service...")
	svc, err := service.NewMergeService()
	if err != nil {
		log.Fatalf("Failed to create merge service: %v", err)
	}

	srv := api.NewServer(svc)

	addr := fmt.Sprintf(":%d", *port)
	log.Printf("Starting server on %s", addr)
	log.Printf("Endpoints:")
	log.Printf("  GET  /docs         - Swagger UI")
	log.Printf("  GET  /openapi.json - OpenAPI spec")
	log.Printf("  POST /render/glb   - Returns GLB binary")
	log.Printf("  POST /render/png   - Returns PNG image")
	log.Printf("  POST /render/gif   - Returns animated GIF")
	log.Printf("  GET  /health       - Health check")

	if err := http.ListenAndServe(addr, srv); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
