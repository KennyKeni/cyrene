package qdrant

import (
	"context"
	"cyrene/internal/config"

	"github.com/qdrant/go-client/qdrant"
)

const partition = "partition"

type Service interface {
	NewCollection(ctx context.Context, name string, vectorSize uint64) error
	DeleteCollection(ctx context.Context, name string) error
	CollectionExists(ctx context.Context, name string) (bool, error)

	Upsert(ctx context.Context, collectionName string, points []Point) error
	Delete(ctx context.Context, collectionName string, ids []string) error
	Query(ctx context.Context, collectionName string, params SearchParams) ([]SearchResult, error)

	Health(ctx context.Context) error
	Close() error
}

type service struct {
	client *qdrant.Client
}

func New(cfg *config.QdrantConfig) (Service, error) {
	qc, err := qdrant.NewClient(&qdrant.Config{
		Host:   cfg.Host,
		Port:   cfg.Port,
		APIKey: cfg.APIKey,
		UseTLS: true,
	})
	if err != nil {
		return nil, err
	}

	return &service{
		client: qc,
	}, nil
}

// Point represents a vector point to store in Qdrant.
type Point struct {
	ID      string
	Vector  []float32
	Payload map[string]any
}

// SearchResult represents a single search result.
type SearchResult struct {
	ID      string
	Score   float32
	Payload map[string]any
}

type SearchParams struct {
	Vector        []float32
	Limit         uint64
	StringFilters []StringFilter
	IntFilters    []IntFilter
	BoolFilters   []BoolFilter
}

type FilterOp string

const (
	FilterAND FilterOp = "AND"
	FilterOR  FilterOp = "OR"
	FilterNOT FilterOp = "NOT"
)

type StringFilter struct {
	Field string
	Value string
	Op    FilterOp
}

type IntFilter struct {
	Field string
	Value int64
	Op    FilterOp
}

type BoolFilter struct {
	Field string
	Value bool
	Op    FilterOp
}

func (s *service) NewCollection(ctx context.Context, name string, vectorSize uint64) error {
	return s.client.CreateCollection(ctx, &qdrant.CreateCollection{
		CollectionName: name,
		VectorsConfig: qdrant.NewVectorsConfig(&qdrant.VectorParams{
			Size:     vectorSize,
			Distance: qdrant.Distance_Cosine,
		}),
		HnswConfig: &qdrant.HnswConfigDiff{
			M:           qdrant.PtrOf(uint64(32)),
			EfConstruct: qdrant.PtrOf(uint64(200)),
		},
	})
}

func (s *service) DeleteCollection(ctx context.Context, collectionName string) error {
	return s.client.DeleteCollection(ctx, collectionName)
}

func (s *service) CollectionExists(ctx context.Context, collectionName string) (bool, error) {
	return s.client.CollectionExists(ctx, collectionName)
}

func (s *service) Upsert(ctx context.Context, collectionName string, points []Point) error {
	p := make([]*qdrant.PointStruct, len(points))
	for i, point := range points {
		p[i] = &qdrant.PointStruct{
			Id:      qdrant.NewID(point.ID),
			Vectors: qdrant.NewVectors(point.Vector...),
			Payload: qdrant.NewValueMap(point.Payload),
		}
	}

	_, err := s.client.Upsert(ctx, &qdrant.UpsertPoints{
		CollectionName: collectionName,
		Points:         p,
	})

	return err
}

func (s *service) Delete(ctx context.Context, collectionName string, ids []string) error {
	p := make([]*qdrant.PointId, len(ids))
	for i, id := range ids {
		p[i] = qdrant.NewID(id)
	}

	_, err := s.client.Delete(ctx, &qdrant.DeletePoints{
		CollectionName: collectionName,
		Points:         qdrant.NewPointsSelector(p...),
	})

	return err
}

func (s *service) Query(ctx context.Context, collectionName string, params SearchParams) ([]SearchResult, error) {
	query := &qdrant.QueryPoints{
		CollectionName: collectionName,
		Query:          qdrant.NewQuery(params.Vector...),
		Limit:          qdrant.PtrOf(params.Limit),
		WithPayload:    qdrant.NewWithPayload(true),
	}

	var should []*qdrant.Condition
	var must []*qdrant.Condition
	var mustNot []*qdrant.Condition

	for _, f := range params.StringFilters {
		cond := qdrant.NewMatch(f.Field, f.Value)
		switch f.Op {
		case FilterOR:
			should = append(should, cond)
		case FilterNOT:
			mustNot = append(mustNot, cond)
		case FilterAND:
			must = append(must, cond)
		default:
			must = append(must, cond)
		}
	}

	for _, f := range params.IntFilters {
		cond := qdrant.NewMatchInt(f.Field, f.Value)
		switch f.Op {
		case FilterOR:
			should = append(should, cond)
		case FilterNOT:
			mustNot = append(mustNot, cond)
		case FilterAND:
			must = append(must, cond)
		default:
			must = append(must, cond)
		}
	}

	for _, f := range params.BoolFilters {
		cond := qdrant.NewMatchBool(f.Field, f.Value)
		switch f.Op {
		case FilterOR:
			should = append(should, cond)
		case FilterNOT:
			mustNot = append(mustNot, cond)
		case FilterAND:
			must = append(must, cond)
		default:
			must = append(must, cond)
		}
	}

	if len(should) > 0 || len(must) > 0 || len(mustNot) > 0 {
		query.Filter = &qdrant.Filter{
			Should:  should,
			Must:    must,
			MustNot: mustNot,
		}
	}

	resp, err := s.client.Query(ctx, query)
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

func (s *service) Health(ctx context.Context) error {
	_, err := s.client.HealthCheck(ctx)
	return err
}

func (s *service) Close() error {
	return s.client.Close()
}
