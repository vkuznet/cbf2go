package main

import (
	"flag"
	"fmt"

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

	//w, h = cbf.ReconcileDimensions(pixels, w, h)

	vec := embed.ImageToEmbedding(pixels, w, h, *size)

	client := &qdrant.Client{
		URL:        *qurl,
		Collection: *qcol,
	}

	err = client.Upsert(uuid.New().String(), vec, map[string]any{
		"path":   *file,
		"width":  w,
		"height": h,
	})
	if err != nil {
		panic(err)
	}

	fmt.Println("Injected:", *file)
}
