package pokemon

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"

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

func (s *Service) SearchPokemon(ctx context.Context, params AgentPokemonParams) (*PaginatedResponse[AgentPokemon], error) {
	q := url.Values{}

	for _, name := range params.Names {
		q.Add("name", name)
	}
	for _, t := range params.Types {
		q.Add("type", t)
	}
	for _, a := range params.Abilities {
		q.Add("abilities", a)
	}
	for _, m := range params.Moves {
		q.Add("moves", m)
	}
	for _, eg := range params.EggGroups {
		q.Add("eggGroups", eg)
	}
	for _, l := range params.Labels {
		q.Add("labels", l)
	}
	for _, d := range params.DropsItems {
		q.Add("dropsItems", d)
	}
	for _, g := range params.Generation {
		q.Add("generation", strconv.Itoa(g))
	}

	if params.IncludeDescription {
		q.Set("includeDescription", "true")
	}
	if params.IncludeGeneration {
		q.Set("includeGeneration", "true")
	}
	if params.IncludeStats {
		q.Set("includeStats", "true")
	}
	if params.IncludeEvYield {
		q.Set("includeEvYield", "true")
	}
	if params.IncludePhysical {
		q.Set("includePhysical", "true")
	}
	if params.IncludeTypes {
		q.Set("includeTypes", "true")
	}
	if params.IncludeAbilities {
		q.Set("includeAbilities", "true")
	}
	if params.IncludeMoves {
		q.Set("includeMoves", "true")
	}
	if params.IncludeDrops {
		q.Set("includeDrops", "true")
	}
	if params.IncludeBreeding {
		q.Set("includeBreeding", "true")
	}
	if params.IncludeEggGroups {
		q.Set("includeEggGroups", "true")
	}
	if params.IncludeExpGroup {
		q.Set("includeExperienceGroup", "true")
	}
	if params.IncludeLabels {
		q.Set("includeLabels", "true")
	}
	if params.IncludeAspects {
		q.Set("includeAspects", "true")
	}
	if params.IncludeHitboxes {
		q.Set("includeHitboxes", "true")
	}
	if params.IncludeLighting {
		q.Set("includeLighting", "true")
	}
	if params.IncludeRiding {
		q.Set("includeRiding", "true")
	}
	if params.IncludeBehaviour {
		q.Set("includeBehaviour", "true")
	}
	if params.IncludeSpawns {
		q.Set("includeSpawns", "true")
	}
	if params.Limit > 0 {
		q.Set("limit", strconv.Itoa(params.Limit))
	}
	if params.Offset > 0 {
		q.Set("offset", strconv.Itoa(params.Offset))
	}

	path := "/agent/pokemon"
	if encoded := q.Encode(); encoded != "" {
		path += "?" + encoded
	}

	data, err := s.fetch(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("search pokemon: %w", err)
	}

	var resp PaginatedResponse[AgentPokemon]
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("decode pokemon response: %w", err)
	}

	return &resp, nil
}

func (s *Service) SearchMoves(ctx context.Context, params AgentMoveParams) (*PaginatedResponse[AgentMove], error) {
	q := url.Values{}

	for _, name := range params.Names {
		q.Add("names", name)
	}
	for _, t := range params.Types {
		q.Add("types", t)
	}
	for _, c := range params.Categories {
		q.Add("categories", c)
	}

	if params.IncludeDescription {
		q.Set("includeDescription", "true")
	}
	if params.IncludeFlags {
		q.Set("includeFlags", "true")
	}
	if params.IncludeBoosts {
		q.Set("includeBoosts", "true")
	}
	if params.IncludeEffects {
		q.Set("includeEffects", "true")
	}
	if params.IncludeZData {
		q.Set("includeZData", "true")
	}
	if params.Limit > 0 {
		q.Set("limit", strconv.Itoa(params.Limit))
	}
	if params.Offset > 0 {
		q.Set("offset", strconv.Itoa(params.Offset))
	}

	path := "/agent/moves"
	if encoded := q.Encode(); encoded != "" {
		path += "?" + encoded
	}

	data, err := s.fetch(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("search moves: %w", err)
	}

	var resp PaginatedResponse[AgentMove]
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("decode moves response: %w", err)
	}

	return &resp, nil
}

func (s *Service) SearchAbilities(ctx context.Context, params AgentAbilityParams) (*PaginatedResponse[AgentAbility], error) {
	q := url.Values{}

	for _, name := range params.Names {
		q.Add("names", name)
	}

	if params.IncludeDescription {
		q.Set("includeDescription", "true")
	}
	if params.IncludeFlags {
		q.Set("includeFlags", "true")
	}
	if params.Limit > 0 {
		q.Set("limit", strconv.Itoa(params.Limit))
	}
	if params.Offset > 0 {
		q.Set("offset", strconv.Itoa(params.Offset))
	}

	path := "/agent/abilities"
	if encoded := q.Encode(); encoded != "" {
		path += "?" + encoded
	}

	data, err := s.fetch(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("search abilities: %w", err)
	}

	var resp PaginatedResponse[AgentAbility]
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("decode abilities response: %w", err)
	}

	return &resp, nil
}

func (s *Service) SearchItems(ctx context.Context, params AgentItemParams) (*PaginatedResponse[AgentItem], error) {
	q := url.Values{}

	for _, name := range params.Names {
		q.Add("names", name)
	}
	for _, tag := range params.Tags {
		q.Add("tags", tag)
	}

	if params.IncludeDescription {
		q.Set("includeDescription", "true")
	}
	if params.IncludeBoosts {
		q.Set("includeBoosts", "true")
	}
	if params.IncludeTags {
		q.Set("includeTags", "true")
	}
	if params.IncludeRecipes {
		q.Set("includeRecipes", "true")
	}
	if params.Limit > 0 {
		q.Set("limit", strconv.Itoa(params.Limit))
	}
	if params.Offset > 0 {
		q.Set("offset", strconv.Itoa(params.Offset))
	}

	path := "/agent/items"
	if encoded := q.Encode(); encoded != "" {
		path += "?" + encoded
	}

	data, err := s.fetch(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("search items: %w", err)
	}

	var resp PaginatedResponse[AgentItem]
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("decode items response: %w", err)
	}

	return &resp, nil
}

func (s *Service) SearchArticles(ctx context.Context, params AgentArticleParams) (*PaginatedResponse[AgentArticleSearch], error) {
	q := url.Values{}

	for _, title := range params.Titles {
		q.Add("titles", title)
	}
	for _, cat := range params.Categories {
		q.Add("categories", cat)
	}

	if params.IncludeBody {
		q.Set("includeBody", "true")
	}
	if params.IncludeCategories {
		q.Set("includeCategories", "true")
	}
	if params.Limit > 0 {
		q.Set("limit", strconv.Itoa(params.Limit))
	}
	if params.Offset > 0 {
		q.Set("offset", strconv.Itoa(params.Offset))
	}

	path := "/agent/articles"
	if encoded := q.Encode(); encoded != "" {
		path += "?" + encoded
	}

	data, err := s.fetch(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("search articles: %w", err)
	}

	var resp PaginatedResponse[AgentArticleSearch]
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("decode articles response: %w", err)
	}

	return &resp, nil
}

func (s *Service) GetArticleBySlug(ctx context.Context, slug string) (*AgentArticle, error) {
	data, err := s.fetch(ctx, fmt.Sprintf("/agent/article/%s", slug))
	if err != nil {
		return nil, fmt.Errorf("fetch article: %w", err)
	}

	var article AgentArticle
	if err := json.Unmarshal(data, &article); err != nil {
		return nil, fmt.Errorf("decode article: %w", err)
	}

	return &article, nil
}

// Get-one endpoints for ingestion

func (s *Service) GetSpecies(ctx context.Context, identifier string) (*Species, error) {
	path := fmt.Sprintf("/pokemon/%s", identifier)
	data, err := s.fetch(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("fetch species: %w", err)
	}

	var species Species
	if err := json.Unmarshal(data, &species); err != nil {
		return nil, fmt.Errorf("decode species: %w", err)
	}

	return &species, nil
}

func (s *Service) GetPokemon(ctx context.Context, identifier string) (*Pokemon, error) {
	q := url.Values{}
	q.Set("includeTypes", "true")
	q.Set("includeAbilities", "true")
	q.Set("includeMoves", "true")
	q.Set("includeLabels", "true")
	q.Set("includeEggGroups", "true")

	path := fmt.Sprintf("/pokemon/form/%s?%s", identifier, q.Encode())
	data, err := s.fetch(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("fetch pokemon form: %w", err)
	}

	var pokemon Pokemon
	if err := json.Unmarshal(data, &pokemon); err != nil {
		return nil, fmt.Errorf("decode pokemon form: %w", err)
	}

	return &pokemon, nil
}

func (s *Service) GetMove(ctx context.Context, identifier string) (*Move, error) {
	q := url.Values{}
	q.Set("includeFlags", "true")

	path := fmt.Sprintf("/moves/%s?%s", identifier, q.Encode())
	data, err := s.fetch(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("fetch move: %w", err)
	}

	var move Move
	if err := json.Unmarshal(data, &move); err != nil {
		return nil, fmt.Errorf("decode move: %w", err)
	}

	return &move, nil
}

func (s *Service) GetAbility(ctx context.Context, identifier string) (*Ability, error) {
	q := url.Values{}
	q.Set("includeFlags", "true")

	path := fmt.Sprintf("/abilities/%s?%s", identifier, q.Encode())
	data, err := s.fetch(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("fetch ability: %w", err)
	}

	var ability Ability
	if err := json.Unmarshal(data, &ability); err != nil {
		return nil, fmt.Errorf("decode ability: %w", err)
	}

	return &ability, nil
}

func (s *Service) GetItem(ctx context.Context, identifier string) (*Item, error) {
	q := url.Values{}
	q.Set("includeBoosts", "true")
	q.Set("includeFlags", "true")
	q.Set("includeTags", "true")

	path := fmt.Sprintf("/items/%s?%s", identifier, q.Encode())
	data, err := s.fetch(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("fetch item: %w", err)
	}

	var item Item
	if err := json.Unmarshal(data, &item); err != nil {
		return nil, fmt.Errorf("decode item: %w", err)
	}

	return &item, nil
}

func (s *Service) GetArticle(ctx context.Context, identifier string) (*Article, error) {
	q := url.Values{}
	q.Set("includeCategories", "true")

	path := fmt.Sprintf("/articles/%s?%s", identifier, q.Encode())
	data, err := s.fetch(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("fetch article: %w", err)
	}

	var article Article
	if err := json.Unmarshal(data, &article); err != nil {
		return nil, fmt.Errorf("decode article: %w", err)
	}

	return &article, nil
}
