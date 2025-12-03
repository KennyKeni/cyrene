package vectorstore

import (
	"context"

	"cyrene/internal/platform/config"
	platformqdrant "cyrene/internal/platform/qdrant"

	"github.com/qdrant/go-client/qdrant"
)

// QdrantStore implements Store using Qdrant as the backend.
type QdrantStore struct {
	client     *qdrant.Client
	collection string
}

// NewQdrantStore creates a new QdrantStore.
func NewQdrantStore(client *platformqdrant.Client, cfg *config.QdrantConfig) *QdrantStore {
	return &QdrantStore{
		client:     client.Conn(),
		collection: cfg.Collection,
	}
}

// Compile safety
var _ Store = (*QdrantStore)(nil)

func (s *QdrantStore) Upsert(ctx context.Context, points []Point) error {
	p := make([]*qdrant.PointStruct, len(points))
	for i, point := range points {
		p[i] = &qdrant.PointStruct{
			Id:      qdrant.NewID(point.ID),
			Vectors: qdrant.NewVectors(point.Vector...),
			Payload: qdrant.NewValueMap(point.Payload),
		}
	}

	_, err := s.client.Upsert(ctx, &qdrant.UpsertPoints{
		CollectionName: s.collection,
		Points:         p,
	})

	return err
}

func (s *QdrantStore) Search(ctx context.Context, vector []float32, limit int) ([]SearchResult, error) {
	resp, err := s.client.Query(ctx, &qdrant.QueryPoints{
		CollectionName: s.collection,
		Query:          qdrant.NewQuery(vector...),
		Limit:          qdrant.PtrOf(uint64(limit)),
		WithPayload:    qdrant.NewWithPayload(true),
	})
	if err != nil {
		return nil, err
	}

	results := make([]SearchResult, len(resp))
	for i, p := range resp {
		results[i] = SearchResult{
			ID:      p.Id.GetUuid(),
			Score:   p.Score,
			Payload: extractPayload(p.Payload),
		}
	}

	return results, nil
}

func (s *QdrantStore) Delete(ctx context.Context, ids []string) error {
	p := make([]*qdrant.PointId, len(ids))
	for i, id := range ids {
		p[i] = qdrant.NewID(id)
	}

	_, err := s.client.Delete(ctx, &qdrant.DeletePoints{
		CollectionName: s.collection,
		Points:         qdrant.NewPointsSelector(p...),
	})

	return err
}

func extractPayload(payload map[string]*qdrant.Value) map[string]any {
	if payload == nil {
		return nil
	}
	result := make(map[string]any, len(payload))
	for k, v := range payload {
		result[k] = valueToAny(v)
	}
	return result
}

func valueToAny(v *qdrant.Value) any {
	if v == nil {
		return nil
	}
	switch val := v.Kind.(type) {
	case *qdrant.Value_StringValue:
		return val.StringValue
	case *qdrant.Value_IntegerValue:
		return val.IntegerValue
	case *qdrant.Value_DoubleValue:
		return val.DoubleValue
	case *qdrant.Value_BoolValue:
		return val.BoolValue
	case *qdrant.Value_ListValue:
		list := make([]any, len(val.ListValue.Values))
		for i, item := range val.ListValue.Values {
			list[i] = valueToAny(item)
		}
		return list
	case *qdrant.Value_StructValue:
		return extractPayload(val.StructValue.Fields)
	default:
		return nil
	}
}
