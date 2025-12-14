package httpapi

import (
	"cbf2go/internal/embed"
	"cbf2go/internal/qdrant"
	"fmt"
)

func HybridSearch(
	qdrant *qdrant.Client,
	clip *embed.CLIPClient,
	imageBytes []byte,
	text string,
) ([]map[string]any, error) {

	var vec []float32
	var err error

	switch {
	case imageBytes != nil:
		vec, err = clip.EmbedImage(imageBytes)
	case text != "":
		vec, err = clip.EmbedText(text)
	default:
		return nil, fmt.Errorf("no query provided")
	}
	if err != nil {
		return nil, err
	}

	return qdrant.Search(vec, 10)
}
