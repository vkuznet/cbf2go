package main

import (
	"cbf2go/internal/cbf"
	"cbf2go/internal/embed"
	"cbf2go/internal/qdrant"

	"github.com/google/uuid"
)

func ingestOne(path string, clip *embed.CLIPClient, client *qdrant.Client) error {
	pixels, w, h, err := cbf.ReadCBF(path)
	if err != nil {
		panic(err)
	}
	//w, h = cbf.ReconcileDimensions(pixels, w, h)

	vec := embed.ImageToEmbedding(pixels, w, h, 224)

	err = client.Upsert(uuid.New().String(), vec, map[string]any{
		"path":   path,
		"width":  w,
		"height": h,
	})
	return err
}

func BatchIngest(
	files []string,
	workers int,
	clip *embed.CLIPClient,
	qdrant *qdrant.Client,
) error {

	jobs := make(chan string)
	errs := make(chan error, workers)

	for i := 0; i < workers; i++ {
		go func() {
			for path := range jobs {
				err := ingestOne(path, clip, qdrant)
				if err != nil {
					errs <- err
				}
			}
		}()
	}

	for _, f := range files {
		jobs <- f
	}
	close(jobs)

	for i := 0; i < workers; i++ {
		select {
		case err := <-errs:
			return err
		default:
		}
	}

	return nil
}
