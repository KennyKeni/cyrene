package pokemon

import (
	"context"
	"cyrene/internal/config"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Pokemon struct {
	ID         string
	Identifier string
	RawJSON    string
	Metadata   map[string]any
}

type Service interface {
	GetPokemonByID(ctx context.Context, id string) (*Pokemon, error)
}

type service struct {
	client  *http.Client
	baseURL string
}

func New(cfg config.PokemonAPIConfig) Service {
	s := &service{
		client:  &http.Client{},
		baseURL: cfg.BaseURL,
	}

	return s
}

func (s *service) GetPokemonByID(ctx context.Context, id string) (*Pokemon, error) {
	url := fmt.Sprintf("%s/pokemon/%s", s.baseURL, id)

	req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)

	// TODO Switch to JSON v2
	var raw map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, err
	}

	rawBytes, _ := json.Marshal(raw)

	return &Pokemon{
		ID:         fmt.Sprintf("%d", id),
		Identifier: raw["name"].(string),
		RawJSON:    string(rawBytes),
		Metadata:   raw,
	}, nil
}
