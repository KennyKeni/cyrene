package ingest

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

var ErrNotFound = errors.New("document not found")

const referenceKey = "reference"
const typeKey = "type"
const contentKey = "content"

type Topic string

const (
	TopicIngestion Topic = "ingestion"
)

type DocumentType string

const (
	DocumentTypePokemon DocumentType = "pokemon"
	DocumentTypeMove    DocumentType = "move"
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
