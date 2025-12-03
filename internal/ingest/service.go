package ingest

import (
	"context"
	"cyrene/internal/pokemon"
	"cyrene/internal/qdrant"
)

type embedder interface {
	Embed(ctx context.Context, texts []string) ([][]float32, error)
}

type vectorstore interface {
	Upsert(ctx context.Context, collectionName string, points []qdrant.Point) error
}

type Service struct {
	Embedder       embedder
	Vectorstore    vectorstore
	PokemonService pokemon.Service
}

func New(embedder embedder, vectorstore vectorstore, pokemonService pokemon.Service) *Service {
	return &Service{
		Embedder:       embedder,
		Vectorstore:    vectorstore,
		PokemonService: pokemonService,
	}
}

func (s *Service) IngestPokemon(ctx context.Context, pokemonId string) {
	pokemon, err := s.PokemonService.GetPokemonByID(ctx, pokemonId)
}
