package qdrant

import (
	"cbf2go/internal/cbf"
	"cbf2go/internal/embed"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	qdrant "github.com/qdrant/go-client/qdrant"
)

// GetFilesFromPath checks if the given path is a file or directory.
// - If it's a file, returns a slice with that single file.
// - If it's a directory, returns a slice of all regular files in that directory (non-recursive).
func GetFilesFromPath(path string) ([]string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("failed to stat path %q: %w", path, err)
	}

	if !info.IsDir() {
		// It's a single file
		return []string{path}, nil
	}

	// It's a directory: list all files (non-recursive)
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory %q: %w", path, err)
	}

	var files []string
	for _, entry := range entries {
		if entry.Type().IsRegular() {
			files = append(files, filepath.Join(path, entry.Name()))
		}
	}

	return files, nil
}

func (c *Client) ensureCollection(ctx context.Context, vectorSize int) error {
	_, err := c.QdrantClient.GetCollectionInfo(ctx, c.Collection)
	if err == nil {
		// Collection exists
		return nil
	}

	// Collection missing — create it
	params := &qdrant.VectorParams{
		Size:     uint64(vectorSize),
		Distance: qdrant.Distance_Cosine, // or DistanceEuclid
	}
	_, err = c.QdrantClient.GetCollectionsClient().Create(ctx, &qdrant.CreateCollection{
		CollectionName: c.Collection,
		VectorsConfig:  qdrant.NewVectorsConfig(params),
	})
	return err
}

func (c *Client) IngestOne(ctx context.Context, path string, vectorSize int) error {

	absPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	pixels, w, h, err := cbf.ReadCBF(path, c.Verbose)
	if err != nil {
		return err
	}

	vec := embed.ImageToEmbedding(pixels, w, h, vectorSize, c.Verbose)

	err = c.Upsert(ctx, uuid.New().String(), vec, map[string]any{
		"filename": filepath.Base(absPath),
		"path":     absPath,
		"width":    int(w),
		"height":   int(h),
		"method":   "pixel",
		"engine":   "cbf2go",
	})
	return err
}

func (c *Client) BatchIngest(path string, workers int, vectorSize int) error {
	t0 := time.Now()
	files, err := GetFilesFromPath(path)
	if err != nil {
		return err
	}

	// set dynamic timeout interval based on number of processing files and nworkers
	timeoutSec := len(files)/workers + 5 // +5s buffer
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutSec)*time.Second)
	defer cancel()

	// Ensure collection exists before upsert
	if err := c.ensureCollection(ctx, vectorSize); err != nil {
		fmt.Printf("failed to create collection: %v\n", err)
		return err
	}

	jobs := make(chan string)
	errs := make(chan error, len(files)) // buffered for all files
	var wg sync.WaitGroup

	// Start worker goroutines
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for f := range jobs {
				if strings.HasSuffix(f, c.FileExtension) {
					fmt.Println("inserting", f)
				} else {
					fmt.Println("skipping", f)
					continue
				}
				if err := c.IngestOne(ctx, f, vectorSize); err != nil {
					// send error and cancel context to stop all workers
					errs <- fmt.Errorf("file %s: %w", f, err)
					cancel()
					return
				}
			}
		}()
	}

	// Send jobs
	for _, f := range files {
		select {
		case <-ctx.Done():
			break
		case jobs <- f:
		}
	}
	close(jobs)

	// Wait for all workers to finish
	wg.Wait()
	close(errs)

	// Collect first error if any
	var firstErr error
	for err := range errs {
		if firstErr == nil {
			firstErr = err
		}
	}

	fmt.Printf("Batch insgestion completed %d files in %v with %d errors", len(files), time.Since(t0), len(errs))
	return firstErr
}
