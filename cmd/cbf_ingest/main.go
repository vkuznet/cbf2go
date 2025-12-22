package main

import (
	"flag"
	"fmt"

	"cbf2go/internal/qdrant"
)

func main() {
	var file, qurl, qcol, fext, eurl string
	var size, verbose, nworkers, timeoutLimit int
	flag.StringVar(&file, "file", "", "CBF file path")
	flag.StringVar(&qurl, "url", "localhost:6334", "Qdrant URL")
	flag.StringVar(&qcol, "collection", "cbf_images", "CBF collection name")
	flag.StringVar(&fext, "file-extension", "cbf", "CBF file extension to use")
	flag.StringVar(&eurl, "embed-url", "image", "URL of embedding service")
	flag.IntVar(&size, "embed-size", 512, "embedding vector size")
	flag.IntVar(&verbose, "verbose", 0, "verbosity level")
	flag.IntVar(&timeoutLimit, "timeout-limit", 60, "timeout limit buffer for batch ingestion")
	flag.IntVar(&nworkers, "nworkers", 10, "number of workers for batch submission")
	flag.Parse()

	client, err := qdrant.NewQdrantClient(qurl, qcol, fext, verbose)
	if err != nil {
		panic(err)
	}
	err = client.BatchIngest(file, nworkers, size, timeoutLimit, eurl)
	if err != nil {
		fmt.Println("ERROR: batch ingestion error", err)
	}
}
