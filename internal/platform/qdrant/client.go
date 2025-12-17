package qdrant

import (
	"context"
	"fmt"

	"cyrene/internal/platform/config"

	"github.com/qdrant/go-client/qdrant"
)

// Client wraps the Qdrant SDK connection.
type Client struct {
	conn *qdrant.Client
}

func New(cfg *config.QdrantConfig) (*Client, error) {
	conn, err := qdrant.NewClient(&qdrant.Config{
		Host:   cfg.Host,
		Port:   cfg.Port,
		APIKey: cfg.APIKey,
		UseTLS: cfg.UseTLS,
	})
	if err != nil {
		return nil, err
	}

	return &Client{conn: conn}, nil
}

func (c *Client) Conn() *qdrant.Client {
	return c.conn
}

func (c *Client) Close() error {
	return c.conn.Close()
}

func (c *Client) EnsureCollection(ctx context.Context, name string, vectorSize uint64) error {
	exists, err := c.conn.CollectionExists(ctx, name)
	if err != nil {
		return fmt.Errorf("check collection exists: %w", err)
	}
	if exists {
		return nil
	}

	err = c.conn.CreateCollection(ctx, &qdrant.CreateCollection{
		CollectionName: name,
		VectorsConfig: qdrant.NewVectorsConfig(&qdrant.VectorParams{
			Size:     vectorSize,
			Distance: qdrant.Distance_Cosine,
		}),
	})
	if err != nil {
		return fmt.Errorf("create collection: %w", err)
	}
	return nil
}

type IndexType string

const (
	IndexTypeKeyword IndexType = "keyword"
	IndexTypeInteger IndexType = "integer"
)

func (c *Client) EnsureIndexes(ctx context.Context, collection string, indexes map[string]IndexType) error {
	for field, indexType := range indexes {
		var fieldType qdrant.FieldType
		var params *qdrant.PayloadIndexParams

		switch indexType {
		case IndexTypeKeyword:
			fieldType = qdrant.FieldType_FieldTypeKeyword
		case IndexTypeInteger:
			fieldType = qdrant.FieldType_FieldTypeInteger
			params = &qdrant.PayloadIndexParams{
				IndexParams: &qdrant.PayloadIndexParams_IntegerIndexParams{
					IntegerIndexParams: &qdrant.IntegerIndexParams{
						Lookup: qdrant.PtrOf(true),
						Range:  qdrant.PtrOf(true),
					},
				},
			}
		}

		_, err := c.conn.CreateFieldIndex(ctx, &qdrant.CreateFieldIndexCollection{
			CollectionName:   collection,
			FieldName:        field,
			FieldType:        &fieldType,
			FieldIndexParams: params,
		})
		if err != nil {
			return fmt.Errorf("create index for %s: %w", field, err)
		}
	}
	return nil
}
