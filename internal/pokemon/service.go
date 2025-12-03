package pokemon

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"cyrene/internal/platform/config"
)

type Service struct {
	client  *http.Client
	baseURL string
}

func New(cfg config.PokemonAPIConfig) *Service {
	return &Service{
		client:  &http.Client{},
		baseURL: cfg.BaseURL,
	}
}

func (s *Service) GetPokemonByID(ctx context.Context, id string) (*Pokemon, error) {
	url := fmt.Sprintf("%s/pokemon/%s", s.baseURL, id)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch pokemon: %w", err)
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	var raw map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	rawBytes, _ := json.Marshal(raw)

	return &Pokemon{
		ID:         id,
		Identifier: raw["name"].(string),
		RawJSON:    string(rawBytes),
		Metadata:   raw,
	}, nil
}
