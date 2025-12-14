package main

import (
	"github.com/gin-gonic/gin"

	"cbf2go/internal/httpapi"
	"cbf2go/internal/qdrant"
)

func main() {
	r := gin.Default()

	server := &httpapi.Server{
		Qdrant: &qdrant.Client{
			URL:        "http://localhost:6333",
			Collection: "cbf_images",
		},
	}

	server.Register(r)
	r.Run(":8080")
}

