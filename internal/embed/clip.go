package embed

import (
	"bytes"
	"encoding/json"
	"net/http"
)

type CLIPClient struct {
	BaseURL string
}

func (c *CLIPClient) EmbedImage(imageBytes []byte) ([]float32, error) {
	return c.post("/embed/image", imageBytes)
}

func (c *CLIPClient) EmbedText(text string) ([]float32, error) {
	b, _ := json.Marshal(map[string]string{"text": text})
	return c.postJSON("/embed/text", b)
}

func (c *CLIPClient) post(path string, raw []byte) ([]float32, error) {
	resp, err := http.Post(
		c.BaseURL+path,
		"application/octet-stream",
		bytes.NewReader(raw),
	)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var out struct {
		Embedding []float32 `json:"embedding"`
	}
	err = json.NewDecoder(resp.Body).Decode(&out)
	return out.Embedding, err
}

func (c *CLIPClient) postJSON(path string, raw []byte) ([]float32, error) {
	resp, err := http.Post(
		c.BaseURL+path,
		"application/json",
		bytes.NewReader(raw),
	)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var out struct {
		Embedding []float32 `json:"embedding"`
	}
	err = json.NewDecoder(resp.Body).Decode(&out)
	return out.Embedding, err
}
