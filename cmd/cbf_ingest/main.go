package main

import (
	"flag"
	"fmt"
	"time"

	"cbf2go/internal/qdrant"
)

func main() {
	var file, qurl, qcol, fext string
	var size, verbose, nworkers int
	flag.StringVar(&file, "file", "", "CBF file path")
	flag.StringVar(&qurl, "url", "localhost:6334", "Qdrant URL")
	flag.StringVar(&qcol, "collection", "cbf_images", "CBF collection name")
	flag.StringVar(&fext, "file-extension", "cbf", "CBF file extension to use")
	flag.IntVar(&size, "image-size", 224, "image size for embedding")
	flag.IntVar(&verbose, "verbose", 0, "verbosity level")
	flag.IntVar(&nworkers, "nworkers", 10, "number of workers for batch submission")
	flag.Parse()

	client, err := qdrant.NewQdrantClient(qurl, qcol, fext, verbose)
	if err != nil {
		panic(err)
	}
	t0 := time.Now()
	err = client.BatchIngest(file, nworkers, size)
	if err != nil {
		fmt.Println("ERROR: batch ingestion error", err)
	}
	fmt.Printf("Batch insgestion completed in %v", time.Since(t0))
}
