package pokemon

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"cyrene/internal/platform/config"
)

type Service struct {
	client  *http.Client
	baseURL string
}

func NewService(cfg config.PokemonAPIConfig) *Service {
	return &Service{
		client:  &http.Client{},
		baseURL: cfg.BaseURL,
	}
}

func (s *Service) fetch(ctx context.Context, path string) ([]byte, error) {
	urlPath := fmt.Sprintf("%s%s", s.baseURL, path)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlPath, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch: %w", err)
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

func (s *Service) GetMoveByID(ctx context.Context, id string) (*Move, error) {
	data, err := s.fetch(ctx, fmt.Sprintf("/moves/%s", id))
	if err != nil {
		return nil, fmt.Errorf("fetch move: %w", err)
	}

	var move Move
	if err := json.Unmarshal(data, &move); err != nil {
		return nil, fmt.Errorf("decode move: %w", err)
	}

	move.RawJSON = string(data)
	return &move, nil
}

func (s *Service) SearchMoves(ctx context.Context, query string, limit int) ([]MoveSearchResult, error) {
	q := url.Values{}
	q.Set("q", query)
	if limit > 0 {
		q.Set("limit", strconv.Itoa(limit))
	}

	data, err := s.fetch(ctx, "/moves/search?"+q.Encode())
	if err != nil {
		return nil, fmt.Errorf("search moves: %w", err)
	}

	var results []MoveSearchResult
	if err := json.Unmarshal(data, &results); err != nil {
		return nil, fmt.Errorf("decode move search: %w", err)
	}

	return results, nil
}

func (s *Service) GetAbilityByID(ctx context.Context, id string) (*Ability, error) {
	data, err := s.fetch(ctx, fmt.Sprintf("/abilities/%s", id))
	if err != nil {
		return nil, fmt.Errorf("fetch ability: %w", err)
	}

	var ability Ability
	if err := json.Unmarshal(data, &ability); err != nil {
		return nil, fmt.Errorf("decode ability: %w", err)
	}

	ability.RawJSON = string(data)
	return &ability, nil
}

func (s *Service) SearchAbilities(ctx context.Context, query string, limit int) ([]AbilitySearchResult, error) {
	q := url.Values{}
	q.Set("q", query)
	if limit > 0 {
		q.Set("limit", strconv.Itoa(limit))
	}

	data, err := s.fetch(ctx, "/abilities/search?"+q.Encode())
	if err != nil {
		return nil, fmt.Errorf("search abilities: %w", err)
	}

	var results []AbilitySearchResult
	if err := json.Unmarshal(data, &results); err != nil {
		return nil, fmt.Errorf("decode ability search: %w", err)
	}

	return results, nil
}

func (s *Service) GetArticleByID(ctx context.Context, id string) (*Article, error) {
	data, err := s.fetch(ctx, fmt.Sprintf("/articles/%s", id))
	if err != nil {
		return nil, fmt.Errorf("fetch article: %w", err)
	}

	var article Article
	if err := json.Unmarshal(data, &article); err != nil {
		return nil, fmt.Errorf("decode article: %w", err)
	}

	article.RawJSON = string(data)
	return &article, nil
}

func (s *Service) SearchArticles(ctx context.Context, query string, limit int) ([]ArticleSearchResult, error) {
	q := url.Values{}
	q.Set("q", query)
	if limit > 0 {
		q.Set("limit", strconv.Itoa(limit))
	}

	data, err := s.fetch(ctx, "/articles/search?"+q.Encode())
	if err != nil {
		return nil, fmt.Errorf("search articles: %w", err)
	}

	var results []ArticleSearchResult
	if err := json.Unmarshal(data, &results); err != nil {
		return nil, fmt.Errorf("decode article search: %w", err)
	}

	return results, nil
}

func (s *Service) SearchForms(ctx context.Context, params FormSearchParams) (*FormSearchResponse, error) {
	q := url.Values{}

	if params.Query != "" {
		q.Set("q", params.Query)
	}
	if params.FormID != "" {
		q.Set("formId", params.FormID)
	}
	if params.SpeciesID != "" {
		q.Set("speciesId", params.SpeciesID)
	}
	if params.VariationID != "" {
		q.Set("variationId", params.VariationID)
	}
	if len(params.Types) > 0 {
		q.Set("types", strings.Join(params.Types, ","))
	}
	if len(params.Abilities) > 0 {
		q.Set("abilities", strings.Join(params.Abilities, ","))
	}
	if len(params.Moves) > 0 {
		q.Set("moves", strings.Join(params.Moves, ","))
	}
	if params.Generation != nil {
		q.Set("generation", strconv.Itoa(*params.Generation))
	}
	if params.MinHP != nil {
		q.Set("minHp", strconv.Itoa(*params.MinHP))
	}
	if params.MaxHP != nil {
		q.Set("maxHp", strconv.Itoa(*params.MaxHP))
	}
	if params.MinAttack != nil {
		q.Set("minAttack", strconv.Itoa(*params.MinAttack))
	}
	if params.MaxAttack != nil {
		q.Set("maxAttack", strconv.Itoa(*params.MaxAttack))
	}
	if params.MinDefense != nil {
		q.Set("minDefense", strconv.Itoa(*params.MinDefense))
	}
	if params.MaxDefense != nil {
		q.Set("maxDefense", strconv.Itoa(*params.MaxDefense))
	}
	if params.MinSpecialAttack != nil {
		q.Set("minSpecialAttack", strconv.Itoa(*params.MinSpecialAttack))
	}
	if params.MaxSpecialAttack != nil {
		q.Set("maxSpecialAttack", strconv.Itoa(*params.MaxSpecialAttack))
	}
	if params.MinSpecialDefense != nil {
		q.Set("minSpecialDefense", strconv.Itoa(*params.MinSpecialDefense))
	}
	if params.MaxSpecialDefense != nil {
		q.Set("maxSpecialDefense", strconv.Itoa(*params.MaxSpecialDefense))
	}
	if params.MinSpeed != nil {
		q.Set("minSpeed", strconv.Itoa(*params.MinSpeed))
	}
	if params.MaxSpeed != nil {
		q.Set("maxSpeed", strconv.Itoa(*params.MaxSpeed))
	}
	if params.MinBST != nil {
		q.Set("minBst", strconv.Itoa(*params.MinBST))
	}
	if params.MaxBST != nil {
		q.Set("maxBst", strconv.Itoa(*params.MaxBST))
	}
	if len(params.Include) > 0 {
		q.Set("include", strings.Join(params.Include, ","))
	}
	if params.Limit > 0 {
		q.Set("limit", strconv.Itoa(params.Limit))
	}
	if params.Offset > 0 {
		q.Set("offset", strconv.Itoa(params.Offset))
	}

	path := "/pokemon/search"
	if encoded := q.Encode(); encoded != "" {
		path += "?" + encoded
	}

	data, err := s.fetch(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("search forms: %w", err)
	}

	var resp FormSearchResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("decode search response: %w", err)
	}

	return &resp, nil
}
