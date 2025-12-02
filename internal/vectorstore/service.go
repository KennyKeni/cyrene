package vectorstore

import (
	"cyrene/internal/qdrant"
)

type Service interface {
}

type service struct {
	client qdrant.Service
}

func New(client qdrant.Service) Service {
	return &service{client: client}
}

//func (s *service) Upsert(ctx context.Context, id string, vector []float64, payload map[string]interface{}) error {
//	return s.client.Upsert(ctx, id, vector, payload)
//}
