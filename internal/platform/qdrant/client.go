package qdrant

import (
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
		UseTLS: cfg.APIKey != "", // Use TLS when API key is provided (cloud)
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
