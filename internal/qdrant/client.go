package qdrant

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	qdrant "github.com/qdrant/go-client/qdrant"
)

type Client struct {
	URL          string
	Collection   string
	QdrantClient *qdrant.Client
}

func NewQdrantClient(qurl, col string) (*Client, error) {
	qclient, err := qdrant.NewClient(&qdrant.Config{
		Host: "localhost",
		Port: 6334,
	})
	if err != nil {
		return nil, err
	}

	return &Client{
		URL:          qurl,
		Collection:   col,
		QdrantClient: qclient,
	}, nil
}

func (c *Client) Search(vec []float32, limit int) ([]map[string]any, error) {
	body := map[string]any{
		"vector": vec,
		"limit":  limit,
	}

	b, _ := json.Marshal(body)
	rurl := fmt.Sprintf("%s/collections/%s/points/search", c.URL, c.Collection)
	log.Println("search", rurl)
	resp, err := http.Post(
		rurl,
		"application/json",
		bytes.NewReader(b),
	)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var out struct {
		Result []map[string]any `json:"result"`
	}
	err = json.NewDecoder(resp.Body).Decode(&out)
	return out.Result, err
}

func (c *Client) SearchWithFilter(
	vec []float32,
	limit int,
	filter map[string]any,
) ([]map[string]any, error) {

	body := map[string]any{
		"vector": vec,
		"limit":  limit,
		"filter": filter,
	}

	b, _ := json.Marshal(body)
	resp, err := http.Post(
		fmt.Sprintf("%s/collections/%s/points/search", c.URL, c.Collection),
		"application/json",
		bytes.NewReader(b),
	)
	fmt.Println("### response", resp, err)
	rec := make(map[string]any)
	var out []map[string]any
	out = append(out, rec)
	return out, nil
}

/* example of filter
filter := map[string]any{
	"must": []map[string]any{
		{
			"key": "detector",
			"match": map[string]any{"value": "PILATUS3_6M"},
		},
		{
			"key": "energy_ev",
			"range": map[string]any{"gte": 14000, "lte": 16000},
		},
	},
}
*/

func (c *Client) Upsert(
	ctx context.Context,
	id string,
	vec []float32,
	payload map[string]any,
) error {

	// Convert payload: map[string]any → map[string]*qdrant.Value
	qPayload := make(map[string]*qdrant.Value, len(payload))
	for k, v := range payload {
		switch t := v.(type) {
		case string:
			qPayload[k] = qdrant.NewValueString(t)
		case int:
			qPayload[k] = qdrant.NewValueInt(int64(t))
		case int64:
			qPayload[k] = qdrant.NewValueInt(t)
		case float64:
			qPayload[k] = qdrant.NewValueDouble(t)
		case bool:
			qPayload[k] = qdrant.NewValueBool(t)
		default:
			return fmt.Errorf("unsupported payload type for key %q: %T", k, v)
		}
	}

	point := &qdrant.PointStruct{
		Id:      qdrant.NewIDUUID(id),
		Vectors: qdrant.NewVectors(vec...),
		Payload: qPayload,
	}

	wait := true
	_, err := c.QdrantClient.Upsert(ctx, &qdrant.UpsertPoints{
		CollectionName: c.Collection,
		Points:         []*qdrant.PointStruct{point},
		Wait:           &wait,
	})

	return err
}
