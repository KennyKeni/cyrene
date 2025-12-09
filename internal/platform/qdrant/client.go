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
		UseTLS: cfg.APIKey != "", // Use TLS when API key is provided
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
