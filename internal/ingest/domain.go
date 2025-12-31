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
	formIDKey = "form_id"

	// Move metadata
	moveIDKey = "move_id"

	// Ability metadata
	abilityIDKey = "ability_id"

	// Item metadata
	itemIDKey = "item_id"

	// Article metadata
	articleIDKey = "article_id"
)

type Topic string

const (
	TopicIngestion Topic = "entity-events"
)

type EntityType string

const (
	EntityTypeSpecies EntityType = "species"
	EntityTypeForm    EntityType = "form"
	EntityTypeMove    EntityType = "move"
	EntityTypeAbility EntityType = "ability"
	EntityTypeItem    EntityType = "item"
	EntityTypeArticle EntityType = "article"
)

type Operation string

const (
	OperationCreate Operation = "CREATE"
	OperationUpdate Operation = "UPDATE"
	OperationDelete Operation = "DELETE"
)

type DocumentType string

const (
	DocumentTypeForm    DocumentType = "form"
	DocumentTypeMove    DocumentType = "move"
	DocumentTypeAbility DocumentType = "ability"
	DocumentTypeItem    DocumentType = "item"
	DocumentTypeArticle DocumentType = "article"
)

func (DocumentType) JSONSchema() *jsonschema.Schema {
	return &jsonschema.Schema{
		Type: "string",
		Enum: []any{
			string(DocumentTypeForm),
			string(DocumentTypeMove),
			string(DocumentTypeAbility),
			string(DocumentTypeItem),
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
	EntityType EntityType `json:"entityType"`
	EntityID   string     `json:"entityId"`
	Operation  Operation  `json:"operation"`
}
