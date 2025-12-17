package ingest

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

var ErrNotFound = errors.New("document not found")

const (
	typeKey    = "type"
	contentKey = "content"

	// Form metadata
	formIDKey     = "form_id"
	speciesIDKey  = "species_id"
	variantIDsKey = "variant_ids"
	formNameKey   = "form_name"

	// Move metadata
	moveIDKey = "move_id"

	// Ability metadata
	abilityIDKey = "ability_id"

	// Article metadata
	articleIDKey = "article_id"
)

type Topic string

const (
	TopicIngestion Topic = "ingestion"
)

type DocumentType string

const (
	DocumentTypeSpecies   DocumentType = "species"
	DocumentTypeVariation DocumentType = "variation"
	DocumentTypeForm      DocumentType = "form"
	DocumentTypeMove      DocumentType = "move"
	DocumentTypeAbility   DocumentType = "ability"
	DocumentTypeArticle   DocumentType = "article"
)

func NewDocumentID(d DocumentType, id string) string {
	return fmt.Sprintf("%s_%s", d, id)
}

type IngestedDocument struct {
	ID           uuid.UUID
	DocumentType DocumentType
	ExternalID   string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type IngestionEvent struct {
	Type DocumentType `json:"type"`
	ID   string       `json:"id"`
}
