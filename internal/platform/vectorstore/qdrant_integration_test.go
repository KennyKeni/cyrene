//go:build integration

package vectorstore

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"testing"

	"cyrene/internal/platform/config"
	platformqdrant "cyrene/internal/platform/qdrant"

	"github.com/google/uuid"
	"github.com/qdrant/go-client/qdrant"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	testStore      *QdrantStore
	testClient     *platformqdrant.Client
	testCollection = "test_vectorstore"
	testDimension  = uint64(4) // Small dimension for tests
)

func TestMain(m *testing.M) {
	host := getEnv("QDRANT_HOST", "localhost")
	port, _ := strconv.Atoi(getEnv("QDRANT_PORT", "6334"))

	cfg := &config.QdrantConfig{
		Host:       host,
		Port:       port,
		Collection: testCollection,
	}

	var err error
	testClient, err = platformqdrant.New(cfg)
	if err != nil {
		fmt.Printf("Failed to create Qdrant client: %v\n", err)
		os.Exit(1)
	}

	ctx := context.Background()

	_ = testClient.Conn().DeleteCollection(ctx, testCollection)

	err = testClient.Conn().CreateCollection(ctx, &qdrant.CreateCollection{
		CollectionName: testCollection,
		VectorsConfig: qdrant.NewVectorsConfig(&qdrant.VectorParams{
			Size:     testDimension,
			Distance: qdrant.Distance_Cosine,
		}),
	})
	if err != nil {
		fmt.Printf("Failed to create test collection: %v\n", err)
		os.Exit(1)
	}

	testStore = NewQdrantStore(testClient, cfg)

	code := m.Run()

	_ = testClient.Conn().DeleteCollection(ctx, testCollection)
	testClient.Close()

	os.Exit(code)
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func cleanupTestPoints(t *testing.T, reference string) {
	t.Helper()
	ctx := context.Background()
	_ = testStore.Delete(ctx, Filter{
		StringFilters: []StringFilter{
			{Field: "reference", Value: reference, Op: FilterAND},
		},
	})
}

func TestQdrantStore_Upsert_Single(t *testing.T) {
	reference := "test-ref-single-" + uuid.NewString()[:8]
	cleanupTestPoints(t, reference)
	defer cleanupTestPoints(t, reference)

	ctx := context.Background()

	point := Point{
		ID:      uuid.NewString(),
		Vector:  []float32{1.0, 0.0, 0.0, 0.0},
		Payload: map[string]any{"reference": reference, "type": "pokemon"},
	}

	err := testStore.Upsert(ctx, point)
	require.NoError(t, err)

	filter := &Filter{
		StringFilters: []StringFilter{
			{Field: "reference", Value: reference, Op: FilterAND},
		},
	}
	results, err := testStore.Search(ctx, []float32{1.0, 0.0, 0.0, 0.0}, 10, filter)
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, point.ID, results[0].ID)
	assert.Equal(t, reference, results[0].Payload["reference"])
}

func TestQdrantStore_Upsert_Multiple(t *testing.T) {
	reference := "test-ref-multi-" + uuid.NewString()[:8]
	cleanupTestPoints(t, reference)
	defer cleanupTestPoints(t, reference)

	ctx := context.Background()

	points := []Point{
		{
			ID:      uuid.NewString(),
			Vector:  []float32{1.0, 0.0, 0.0, 0.0},
			Payload: map[string]any{"reference": reference, "index": 0},
		},
		{
			ID:      uuid.NewString(),
			Vector:  []float32{0.0, 1.0, 0.0, 0.0},
			Payload: map[string]any{"reference": reference, "index": 1},
		},
		{
			ID:      uuid.NewString(),
			Vector:  []float32{0.0, 0.0, 1.0, 0.0},
			Payload: map[string]any{"reference": reference, "index": 2},
		},
	}

	err := testStore.Upsert(ctx, points...)
	require.NoError(t, err)

	filter := &Filter{
		StringFilters: []StringFilter{
			{Field: "reference", Value: reference, Op: FilterAND},
		},
	}
	results, err := testStore.Search(ctx, []float32{1.0, 0.0, 0.0, 0.0}, 10, filter)
	require.NoError(t, err)
	assert.Len(t, results, 3)
}

func TestQdrantStore_Search_Basic(t *testing.T) {
	reference := "test-ref-search-" + uuid.NewString()[:8]
	cleanupTestPoints(t, reference)
	defer cleanupTestPoints(t, reference)

	ctx := context.Background()

	points := []Point{
		{
			ID:      uuid.NewString(),
			Vector:  []float32{1.0, 0.0, 0.0, 0.0},
			Payload: map[string]any{"reference": reference, "name": "close"},
		},
		{
			ID:      uuid.NewString(),
			Vector:  []float32{0.0, 0.0, 0.0, 1.0},
			Payload: map[string]any{"reference": reference, "name": "far"},
		},
	}

	err := testStore.Upsert(ctx, points...)
	require.NoError(t, err)

	filter := &Filter{
		StringFilters: []StringFilter{
			{Field: "reference", Value: reference, Op: FilterAND},
		},
	}
	results, err := testStore.Search(ctx, []float32{1.0, 0.0, 0.0, 0.0}, 10, filter)
	require.NoError(t, err)
	require.Len(t, results, 2)

	assert.Equal(t, "close", results[0].Payload["name"])
	assert.Greater(t, results[0].Score, results[1].Score)
}

func TestQdrantStore_Search_WithFilter(t *testing.T) {
	refA := "test-filter-a-" + uuid.NewString()[:8]
	refB := "test-filter-b-" + uuid.NewString()[:8]
	cleanupTestPoints(t, refA)
	cleanupTestPoints(t, refB)
	defer cleanupTestPoints(t, refA)
	defer cleanupTestPoints(t, refB)

	ctx := context.Background()

	points := []Point{
		{
			ID:      uuid.NewString(),
			Vector:  []float32{1.0, 0.0, 0.0, 0.0},
			Payload: map[string]any{"reference": refA, "type": "pokemon"},
		},
		{
			ID:      uuid.NewString(),
			Vector:  []float32{0.9, 0.1, 0.0, 0.0},
			Payload: map[string]any{"reference": refB, "type": "pokemon"},
		},
	}

	err := testStore.Upsert(ctx, points...)
	require.NoError(t, err)

	filter := &Filter{
		StringFilters: []StringFilter{
			{Field: "reference", Value: refA, Op: FilterAND},
		},
	}

	results, err := testStore.Search(ctx, []float32{1.0, 0.0, 0.0, 0.0}, 10, filter)
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, refA, results[0].Payload["reference"])
}

func TestQdrantStore_Search_Limit(t *testing.T) {
	reference := "test-ref-limit-" + uuid.NewString()[:8]
	cleanupTestPoints(t, reference)
	defer cleanupTestPoints(t, reference)

	ctx := context.Background()

	points := make([]Point, 5)
	for i := 0; i < 5; i++ {
		vec := []float32{0.0, 0.0, 0.0, 0.0}
		vec[i%4] = 1.0
		points[i] = Point{
			ID:      uuid.NewString(),
			Vector:  vec,
			Payload: map[string]any{"reference": reference, "index": i},
		}
	}

	err := testStore.Upsert(ctx, points...)
	require.NoError(t, err)

	filter := &Filter{
		StringFilters: []StringFilter{
			{Field: "reference", Value: reference, Op: FilterAND},
		},
	}

	results, err := testStore.Search(ctx, []float32{1.0, 0.0, 0.0, 0.0}, 2, filter)
	require.NoError(t, err)
	assert.Len(t, results, 2)
}

func TestQdrantStore_Delete_ByFilter(t *testing.T) {
	refKeep := "test-delete-keep-" + uuid.NewString()[:8]
	refRemove := "test-delete-remove-" + uuid.NewString()[:8]
	cleanupTestPoints(t, refKeep)
	cleanupTestPoints(t, refRemove)
	defer cleanupTestPoints(t, refKeep)
	defer cleanupTestPoints(t, refRemove)

	ctx := context.Background()

	keepPoint := Point{
		ID:      uuid.NewString(),
		Vector:  []float32{1.0, 0.0, 0.0, 0.0},
		Payload: map[string]any{"reference": refKeep},
	}
	removePoint := Point{
		ID:      uuid.NewString(),
		Vector:  []float32{0.0, 1.0, 0.0, 0.0},
		Payload: map[string]any{"reference": refRemove},
	}

	err := testStore.Upsert(ctx, keepPoint, removePoint)
	require.NoError(t, err)

	err = testStore.Delete(ctx, Filter{
		StringFilters: []StringFilter{
			{Field: "reference", Value: refRemove, Op: FilterAND},
		},
	})
	require.NoError(t, err)

	filterKeep := &Filter{
		StringFilters: []StringFilter{
			{Field: "reference", Value: refKeep, Op: FilterAND},
		},
	}
	results, err := testStore.Search(ctx, []float32{1.0, 0.0, 0.0, 0.0}, 10, filterKeep)
	require.NoError(t, err)
	assert.Len(t, results, 1)

	filterRemove := &Filter{
		StringFilters: []StringFilter{
			{Field: "reference", Value: refRemove, Op: FilterAND},
		},
	}
	results, err = testStore.Search(ctx, []float32{0.0, 1.0, 0.0, 0.0}, 10, filterRemove)
	require.NoError(t, err)
	assert.Empty(t, results)
}

func TestQdrantStore_DeleteByID(t *testing.T) {
	reference := "test-deleteid-" + uuid.NewString()[:8]
	cleanupTestPoints(t, reference)
	defer cleanupTestPoints(t, reference)

	ctx := context.Background()

	point1ID := uuid.NewString()
	point2ID := uuid.NewString()

	points := []Point{
		{ID: point1ID, Vector: []float32{1.0, 0.0, 0.0, 0.0}, Payload: map[string]any{"reference": reference}},
		{ID: point2ID, Vector: []float32{0.0, 1.0, 0.0, 0.0}, Payload: map[string]any{"reference": reference}},
	}

	err := testStore.Upsert(ctx, points...)
	require.NoError(t, err)

	err = testStore.DeleteByID(ctx, point1ID)
	require.NoError(t, err)

	filter := &Filter{
		StringFilters: []StringFilter{
			{Field: "reference", Value: reference, Op: FilterAND},
		},
	}
	results, err := testStore.Search(ctx, []float32{0.0, 1.0, 0.0, 0.0}, 10, filter)
	require.NoError(t, err)

	require.Len(t, results, 1)
	assert.Equal(t, point2ID, results[0].ID)
}
