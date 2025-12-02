package indexer

import (
	"context"
	"cyrene/internal/llm"
	"cyrene/internal/pokemon"
)

type Service interface {
	Index(ctx context.Context, id int) error
}

type service struct {
	llm         llm.Service
	vectorstore any
	pokemonAPI  pokemon.Service
}
