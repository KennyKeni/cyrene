package ingest

import (
	"context"
	"cyrene/internal/platform/postgres/jet/cyrene/public/model"
	"cyrene/internal/platform/postgres/jet/cyrene/public/table"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/go-jet/jet/v2/postgres"
	"github.com/google/uuid"
)

type db interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}

type postgresRepository struct {
	db   db
	conn *sql.DB
}

func NewRepository(conn *sql.DB) Repository {
	return &postgresRepository{
		db:   conn,
		conn: conn,
	}
}

func (r *postgresRepository) Upsert(ctx context.Context, doc *IngestedDocument) error {
	now := time.Now()

	stmt := table.IngestedDocuments.INSERT(
		table.IngestedDocuments.ID,
		table.IngestedDocuments.DocumentType,
		table.IngestedDocuments.ExternalID,
		table.IngestedDocuments.CreatedAt,
		table.IngestedDocuments.UpdatedAt,
	).VALUES(
		doc.ID,
		doc.DocumentType,
		doc.ExternalID,
		now,
		now,
	).ON_CONFLICT(
		table.IngestedDocuments.DocumentType,
		table.IngestedDocuments.ExternalID,
	).DO_UPDATE(
		postgres.SET(
			table.IngestedDocuments.UpdatedAt.SET(postgres.TimestampzT(now)),
		),
	)

	_, err := stmt.ExecContext(ctx, r.db)
	if err != nil {
		return fmt.Errorf("upsert document: %w", err)
	}
	return nil
}

func (r *postgresRepository) Delete(ctx context.Context, id uuid.UUID) error {
	stmt := table.IngestedDocuments.DELETE().WHERE(table.IngestedDocuments.ID.EQ(postgres.UUID(id)))

	_, err := stmt.ExecContext(ctx, r.db)
	if err != nil {
		return err
	}
	return nil
}

func (r *postgresRepository) DeleteByRef(ctx context.Context, dt DocumentType, externalID string) error {
	stmt := table.IngestedDocuments.DELETE().WHERE(
		table.IngestedDocuments.DocumentType.EQ(postgres.String(string(dt))).
			AND(table.IngestedDocuments.ExternalID.EQ(postgres.String(externalID))),
	)

	_, err := stmt.ExecContext(ctx, r.db)
	if err != nil {
		return err
	}
	return nil
}

func (r *postgresRepository) FindByRef(ctx context.Context, dt DocumentType, externalID string) (*IngestedDocument, error) {
	stmt := postgres.SELECT(table.IngestedDocuments.AllColumns).
		FROM(table.IngestedDocuments).
		WHERE(
			table.IngestedDocuments.DocumentType.EQ(postgres.String(string(dt))).
				AND(table.IngestedDocuments.ExternalID.EQ(postgres.String(externalID))),
		)

	var dest model.IngestedDocuments
	err := stmt.QueryContext(ctx, r.db, &dest)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("find document: %w", err)
	}

	return toDomain(&dest), nil
}

func (r *postgresRepository) InTx(ctx context.Context, fn func(Repository) error) error {
	tx, err := r.conn.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func(tx *sql.Tx) {
		err := tx.Rollback()
		if err != nil {

		}
	}(tx)

	txRepo := &postgresRepository{db: tx, conn: r.conn}
	if err := fn(txRepo); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}
	return nil
}

// toDomain maps ingestedDocuments Jet model to domain struct
func toDomain(m *model.IngestedDocuments) *IngestedDocument {
	return &IngestedDocument{
		ID:           m.ID,
		DocumentType: DocumentType(m.DocumentType),
		ExternalID:   m.ExternalID,
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
	}
}
