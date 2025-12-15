package qdrant

import (
	"context"
	"fmt"

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

func payloadToMap(p map[string]*qdrant.Value) map[string]any {
	out := make(map[string]any, len(p))
	for k, v := range p {
		switch x := v.Kind.(type) {
		case *qdrant.Value_StringValue:
			out[k] = x.StringValue
		case *qdrant.Value_IntegerValue:
			out[k] = x.IntegerValue
		case *qdrant.Value_DoubleValue:
			out[k] = x.DoubleValue
		case *qdrant.Value_BoolValue:
			out[k] = x.BoolValue
		default:
			out[k] = nil
		}
	}
	return out
}

func buildQdrantFilter(filter map[string]any) (*qdrant.Filter, error) {
	mustRaw, ok := filter["must"].([]map[string]any)
	if !ok {
		return nil, fmt.Errorf("only 'must' filters are supported")
	}

	var must []*qdrant.Condition

	for _, f := range mustRaw {
		key := f["key"].(string)

		if match, ok := f["match"].(map[string]any); ok {
			must = append(must,
				qdrant.NewMatchKeyword(key, fmt.Sprint(match["value"])),
			)
			continue
		}

		if rng, ok := f["range"].(map[string]any); ok {
			r := &qdrant.Range{}
			if gte, ok := rng["gte"]; ok {
				switch v := gte.(type) {
				case float64:
					r.Gte = &v
				}
			}
			if lte, ok := rng["lte"]; ok {
				switch v := lte.(type) {
				case float64:
					r.Lte = &v
				}
			}
			// TODO: I need to add my range to condition
			cond := &qdrant.Condition{}
			must = append(must, cond)
			continue
		}

		return nil, fmt.Errorf("unsupported filter condition: %v", f)
	}

	return &qdrant.Filter{Must: must}, nil
}

func (c *Client) Search(vec []float32, limit int) ([]map[string]any, error) {
	ctx := context.Background()
	pointsClient := c.QdrantClient.GetPointsClient()

	lim := uint64(limit)
	resp, err := pointsClient.Search(ctx, &qdrant.SearchPoints{
		CollectionName: c.Collection,
		Vector:         vec,
		Limit:          lim,
		WithPayload:    qdrant.NewWithPayload(true),
	})
	if err != nil {
		return nil, err
	}
	res := resp.GetResult()

	out := make([]map[string]any, 0, len(res))
	for _, p := range res {
		rec := payloadToMap(p.Payload)
		rec["_score"] = p.Score
		rec["_id"] = p.Id.GetUuid()
		out = append(out, rec)
	}

	return out, nil
}

func (c *Client) SearchWithFilter(
	vec []float32,
	limit int,
	filter map[string]any,
) ([]map[string]any, error) {

	ctx := context.Background()
	pointsClient := c.QdrantClient.GetPointsClient()

	qfilter, err := buildQdrantFilter(filter)
	if err != nil {
		return nil, err
	}

	lim := uint64(limit)
	resp, err := pointsClient.Search(ctx, &qdrant.SearchPoints{
		CollectionName: c.Collection,
		Vector:         vec,
		Limit:          lim,
		Filter:         qfilter,
		WithPayload:    qdrant.NewWithPayload(true),
	})
	if err != nil {
		return nil, err
	}

	res := resp.GetResult()

	out := make([]map[string]any, 0, len(res))
	for _, p := range res {
		rec := payloadToMap(p.Payload)
		rec["_score"] = p.Score
		rec["_id"] = p.Id.GetUuid()
		out = append(out, rec)
	}

	return out, nil
}

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
