package embed

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type EmbedClient struct {
	BaseURL    string
	HttpClient *http.Client
}

func NewEmbedClient(baseURL string) *EmbedClient {
	return &EmbedClient{
		BaseURL: baseURL,
		HttpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

type PixelRequest struct {
	Pixels []float32 `json:"pixels"`
	Height int       `json:"height"`
	Width  int       `json:"width"`
	DType  string    `json:"dtype"`
}

type EmbedResponse struct {
	Embedding []float32 `json:"embedding"`
	Dim       int       `json:"dim"`
}

func (c *EmbedClient) EmbedPixels(pixels []float32, h, w int) ([]float32, error) {
	reqBody := PixelRequest{
		Pixels: pixels,
		Height: h,
		Width:  w,
		DType:  "float32",
	}

	data, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(
		"POST",
		c.BaseURL+"/embed/pixels",
		bytes.NewReader(data),
	)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server error: %s", resp.Status)
	}

	var out EmbedResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}

	return out.Embedding, nil
}
