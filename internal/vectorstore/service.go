package vectorstore

import "context"

// Store defines the interface for vector storage operations.
type Store interface {
	Upsert(ctx context.Context, points ...Point) error
	Search(ctx context.Context, vector []float32, limit int, filter *Filter) ([]SearchResult, error)
	Delete(ctx context.Context, filter Filter) error
	DeleteByID(ctx context.Context, ids ...string) error
}
