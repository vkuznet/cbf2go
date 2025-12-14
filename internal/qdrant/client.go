package qdrant

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type Client struct {
	URL        string
	Collection string
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

func (c *Client) Upsert(id string, vec []float32, payload map[string]any) error {
	body := map[string]any{
		"points": []map[string]any{
			{
				"id":      id,
				"vector":  vec,
				"payload": payload,
			},
		},
	}

	b, _ := json.Marshal(body)
	resp, err := http.Post(
		fmt.Sprintf("%s/collections/%s/points", c.URL, c.Collection),
		"application/json",
		bytes.NewReader(b),
	)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
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
