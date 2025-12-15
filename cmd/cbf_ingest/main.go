package main

import (
	"context"
	"flag"
	"fmt"
	"path/filepath"
	"time"

	"github.com/google/uuid"

	"cbf2go/internal/cbf"
	"cbf2go/internal/embed"
	"cbf2go/internal/qdrant"
)

func main() {
	file := flag.String("file", "", "CBF file path")
	qurl := flag.String("url", "http://localhost:6333", "Qdrant URL")
	qcol := flag.String("collection", "cbf_images", "CBF collection name")
	size := flag.Int("image-size", 224, "image size for embedding")
	flag.Parse()

	if *file == "" {
		panic("missing --file")
	}

	pixels, w, h, err := cbf.ReadCBF(*file)
	if err != nil {
		panic(err)
	}

	vec := embed.ImageToEmbedding(pixels, w, h, *size)

	client, err := qdrant.NewQdrantClient(*qurl, *qcol)
	if err != nil {
		panic(err)
	}

	absPath, err := filepath.Abs(*file)
	if err != nil {
		panic(err)
	}

	payload := map[string]any{
		"filename": filepath.Base(absPath),
		"path":     absPath,
		"width":    int(w),
		"height":   int(h),
		"method":   "pixel",
		"engine":   "cbf2go",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = client.Upsert(ctx, uuid.New().String(), vec, payload)
	if err != nil {
		panic(err)
	}

	fmt.Println("Injected:", absPath)
}
