package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"

	"cbf2go/internal/httpapi"
	"cbf2go/internal/qdrant"
)

func main() {
	configPath := "config.yaml" // or "config.json"

	cfg, err := LoadConfig(configPath)
	if err != nil {
		log.Printf("failed to load config %q: %v, using defaults", configPath, err)
		cfg = &Config{}
	}

	// Default values if not set
	if cfg.Server.Host == "" {
		cfg.Server.Host = "0.0.0.0"
	}
	if cfg.Server.Port == 0 {
		cfg.Server.Port = 8111
	}
	if cfg.Qdrant.URL == "" {
		cfg.Qdrant.URL = "http://localhost:6334"
	}
	if cfg.Qdrant.Collection == "" {
		cfg.Qdrant.Collection = "cbf_images"
	}

	// Initialize Qdrant client
	client, err := qdrant.NewQdrantClient(cfg.Qdrant.URL, cfg.Qdrant.Collection, cfg.Qdrant.Verbose)
	if err != nil {
		log.Fatalf("failed to create qdrant client: %v", err)
	}

	// Initialize Gin server
	r := gin.Default()
	server := &httpapi.Server{
		Qdrant: client,
	}

	server.Register(r)
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	r.Run(addr)
}

