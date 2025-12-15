package main

import (
	"flag"
	"fmt"

	"cbf2go/internal/qdrant"
)

func main() {
	file := flag.String("file", "", "CBF file path")
	qurl := flag.String("url", "http://localhost:6334", "Qdrant URL")
	qcol := flag.String("collection", "cbf_images", "CBF collection name")
	size := flag.Int("image-size", 224, "image size for embedding")
	verbose := flag.Int("verbose", 0, "verbosity level")
	nworkers := flag.Int("nworkers", 10, "number of workers for batch submission")
	flag.Parse()

	client, err := qdrant.NewQdrantClient(*qurl, *qcol, *verbose)
	if err != nil {
		panic(err)
	}
	err = client.BatchIngest(*file, *nworkers, *size)
	if err != nil {
		fmt.Println("ERROR: batch ingestion error", err)
	}
}
