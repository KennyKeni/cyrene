package ingest

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/invopop/jsonschema"
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

type EventType string

const (
	EventTypeSpecies   EventType = "species"
	EventTypeVariation EventType = "variation"
	EventTypeForm      EventType = "form"
	EventTypeMove      EventType = "move"
	EventTypeAbility   EventType = "ability"
	EventTypeArticle   EventType = "article"
)

type DocumentType string

const (
	DocumentTypeForm    DocumentType = "form"
	DocumentTypeMove    DocumentType = "move"
	DocumentTypeAbility DocumentType = "ability"
	DocumentTypeArticle DocumentType = "article"
)

func (DocumentType) JSONSchema() *jsonschema.Schema {
	return &jsonschema.Schema{
		Type: "string",
		Enum: []any{
			string(DocumentTypeForm),
			string(DocumentTypeMove),
			string(DocumentTypeAbility),
			string(DocumentTypeArticle),
		},
	}
}

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
	Type EventType `json:"type"`
	ID   string    `json:"id"`
}
