//go:build integration

package ingest

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testDB *sql.DB

func TestMain(m *testing.M) {
	connStr := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		getEnv("DB_USERNAME", "cyrene"),
		getEnv("DB_PASSWORD", "password1234"),
		getEnv("DB_HOST", "localhost"),
		getEnv("DB_PORT", "5432"),
		getEnv("DB_DATABASE", "cyrene"),
	)

	var err error
	testDB, err = sql.Open("pgx", connStr)
	if err != nil {
		fmt.Printf("Failed to connect to database: %v\n", err)
		os.Exit(1)
	}

	if err := testDB.Ping(); err != nil {
		fmt.Printf("Failed to ping database: %v\n", err)
		os.Exit(1)
	}

	code := m.Run()

	testDB.Close()
	os.Exit(code)
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func cleanupTestData(t *testing.T, externalIDs ...string) {
	t.Helper()
	for _, id := range externalIDs {
		_, err := testDB.ExecContext(
			context.Background(),
			"DELETE FROM ingested_documents WHERE external_id = $1",
			id,
		)
		if err != nil {
			t.Logf("Warning: cleanup failed for %s: %v", id, err)
		}
	}
}

func TestRepository_Upsert_Insert(t *testing.T) {
	cleanupTestData(t, "test-pokemon-1")
	defer cleanupTestData(t, "test-pokemon-1")

	ctx := context.Background()
	repo := NewRepository(testDB)

	doc := &IngestedDocument{
		ID:           uuid.Must(uuid.NewV7()),
		DocumentType: DocumentTypePokemon,
		ExternalID:   "test-pokemon-1",
	}

	err := repo.Upsert(ctx, doc)
	require.NoError(t, err)

	found, err := repo.FindByRef(ctx, DocumentTypePokemon, "test-pokemon-1")
	require.NoError(t, err)
	assert.Equal(t, doc.ID, found.ID)
	assert.Equal(t, DocumentTypePokemon, found.DocumentType)
	assert.Equal(t, "test-pokemon-1", found.ExternalID)
	assert.False(t, found.CreatedAt.IsZero())
	assert.False(t, found.UpdatedAt.IsZero())
}

func TestRepository_Upsert_Update(t *testing.T) {
	cleanupTestData(t, "test-pokemon-2")
	defer cleanupTestData(t, "test-pokemon-2")

	ctx := context.Background()
	repo := NewRepository(testDB)

	doc1 := &IngestedDocument{
		ID:           uuid.Must(uuid.NewV7()),
		DocumentType: DocumentTypePokemon,
		ExternalID:   "test-pokemon-2",
	}
	err := repo.Upsert(ctx, doc1)
	require.NoError(t, err)

	original, err := repo.FindByRef(ctx, DocumentTypePokemon, "test-pokemon-2")
	require.NoError(t, err)

	time.Sleep(10 * time.Millisecond)

	doc2 := &IngestedDocument{
		ID:           uuid.Must(uuid.NewV7()),
		DocumentType: DocumentTypePokemon,
		ExternalID:   "test-pokemon-2",
	}
	err = repo.Upsert(ctx, doc2)
	require.NoError(t, err)

	updated, err := repo.FindByRef(ctx, DocumentTypePokemon, "test-pokemon-2")
	require.NoError(t, err)

	assert.Equal(t, original.ID, updated.ID)
	assert.True(t, updated.UpdatedAt.After(original.UpdatedAt) || updated.UpdatedAt.Equal(original.UpdatedAt))
}

func TestRepository_FindByRef_Found(t *testing.T) {
	cleanupTestData(t, "test-pokemon-find")
	defer cleanupTestData(t, "test-pokemon-find")

	ctx := context.Background()
	repo := NewRepository(testDB)

	doc := &IngestedDocument{
		ID:           uuid.Must(uuid.NewV7()),
		DocumentType: DocumentTypePokemon,
		ExternalID:   "test-pokemon-find",
	}
	err := repo.Upsert(ctx, doc)
	require.NoError(t, err)

	found, err := repo.FindByRef(ctx, DocumentTypePokemon, "test-pokemon-find")
	require.NoError(t, err)
	assert.Equal(t, doc.ID, found.ID)
	assert.Equal(t, "test-pokemon-find", found.ExternalID)
}

func TestRepository_FindByRef_NotFound(t *testing.T) {
	ctx := context.Background()
	repo := NewRepository(testDB)

	_, err := repo.FindByRef(ctx, DocumentTypePokemon, "nonexistent-pokemon-xyz")
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestRepository_Delete(t *testing.T) {
	cleanupTestData(t, "test-pokemon-delete")
	defer cleanupTestData(t, "test-pokemon-delete")

	ctx := context.Background()
	repo := NewRepository(testDB)

	doc := &IngestedDocument{
		ID:           uuid.Must(uuid.NewV7()),
		DocumentType: DocumentTypePokemon,
		ExternalID:   "test-pokemon-delete",
	}
	err := repo.Upsert(ctx, doc)
	require.NoError(t, err)

	err = repo.Delete(ctx, doc.ID)
	require.NoError(t, err)

	_, err = repo.FindByRef(ctx, DocumentTypePokemon, "test-pokemon-delete")
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestRepository_DeleteByRef(t *testing.T) {
	cleanupTestData(t, "test-pokemon-deleteref")
	defer cleanupTestData(t, "test-pokemon-deleteref")

	ctx := context.Background()
	repo := NewRepository(testDB)

	doc := &IngestedDocument{
		ID:           uuid.Must(uuid.NewV7()),
		DocumentType: DocumentTypePokemon,
		ExternalID:   "test-pokemon-deleteref",
	}
	err := repo.Upsert(ctx, doc)
	require.NoError(t, err)

	err = repo.DeleteByRef(ctx, DocumentTypePokemon, "test-pokemon-deleteref")
	require.NoError(t, err)

	_, err = repo.FindByRef(ctx, DocumentTypePokemon, "test-pokemon-deleteref")
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestRepository_InTx_Commit(t *testing.T) {
	cleanupTestData(t, "test-tx-commit")
	defer cleanupTestData(t, "test-tx-commit")

	ctx := context.Background()
	repo := NewRepository(testDB)

	err := repo.InTx(ctx, func(txRepo Repository) error {
		doc := &IngestedDocument{
			ID:           uuid.Must(uuid.NewV7()),
			DocumentType: DocumentTypePokemon,
			ExternalID:   "test-tx-commit",
		}
		return txRepo.Upsert(ctx, doc)
	})
	require.NoError(t, err)

	found, err := repo.FindByRef(ctx, DocumentTypePokemon, "test-tx-commit")
	require.NoError(t, err)
	assert.Equal(t, "test-tx-commit", found.ExternalID)
}

func TestRepository_InTx_Rollback(t *testing.T) {
	cleanupTestData(t, "test-tx-rollback")
	defer cleanupTestData(t, "test-tx-rollback")

	ctx := context.Background()
	repo := NewRepository(testDB)

	expectedErr := errors.New("forced rollback")

	err := repo.InTx(ctx, func(txRepo Repository) error {
		doc := &IngestedDocument{
			ID:           uuid.Must(uuid.NewV7()),
			DocumentType: DocumentTypePokemon,
			ExternalID:   "test-tx-rollback",
		}
		if err := txRepo.Upsert(ctx, doc); err != nil {
			return err
		}
		return expectedErr
	})

	require.ErrorIs(t, err, expectedErr)

	_, err = repo.FindByRef(ctx, DocumentTypePokemon, "test-tx-rollback")
	assert.ErrorIs(t, err, ErrNotFound)
}
