package ingest

import (
	"context"

	"cyrene/internal/vectorstore"
)

type Service struct {
	embedder       embedder
	store          vectorstore.Store
	pokemonService pokemonGetter
}

func New(embedder embedder, store vectorstore.Store, pokemonService pokemonGetter) *Service {
	return &Service{
		embedder:       embedder,
		store:          store,
		pokemonService: pokemonService,
	}
}

func (s *Service) IngestPokemon(ctx context.Context, pokemonID string) error {
	_, err := s.pokemonService.GetPokemonByID(ctx, pokemonID)
	if err != nil {
		return err
	}
	// TODO: embed and upsert
	return nil
}
