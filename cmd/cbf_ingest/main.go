package main

import (
	"flag"
	"fmt"

	"cbf2go/internal/qdrant"
)

func main() {
	var file, qurl, qcol string
	var size, verbose, nworkers int
	flag.StringVar(&file, "file", "", "CBF file path")
	flag.StringVar(&qurl, "url", "http://localhost:6334", "Qdrant URL")
	flag.StringVar(&qcol, "collection", "cbf_images", "CBF collection name")
	flag.IntVar(&size, "image-size", 224, "image size for embedding")
	flag.IntVar(&verbose, "verbose", 0, "verbosity level")
	flag.IntVar(&nworkers, "nworkers", 10, "number of workers for batch submission")
	flag.Parse()

	client, err := qdrant.NewQdrantClient(qurl, qcol, verbose)
	if err != nil {
		panic(err)
	}
	err = client.BatchIngest(file, nworkers, size)
	if err != nil {
		fmt.Println("ERROR: batch ingestion error", err)
	}
}
