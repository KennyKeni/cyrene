//go:build integration

package ingest

import (
	"context"
	"encoding/json"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"cyrene/internal/platform/config"
	"cyrene/internal/platform/kafka"
	platformqdrant "cyrene/internal/platform/qdrant"
	"cyrene/internal/platform/vectorstore"
	"cyrene/internal/pokemon"

	"github.com/google/uuid"
	"github.com/qdrant/go-client/qdrant"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/twmb/franz-go/pkg/kadm"
	"github.com/twmb/franz-go/pkg/kgo"
)

const (
	testCollection = "test_pipeline"
	testDimension  = uint64(4)
)

type stubEmbedService struct{}

func (s *stubEmbedService) Embed(ctx context.Context, texts ...string) ([][]float32, error) {
	return [][]float32{{1.0, 0.0, 0.0, 0.0}}, nil
}

type stubPokemonService struct {
	calledWith string
}

func (s *stubPokemonService) GetPokemonByID(ctx context.Context, id string) (*pokemon.Pokemon, error) {
	s.calledWith = id
	return &pokemon.Pokemon{
		ID:      id,
		RawJSON: `{"id":25,"name":"pikachu"}`,
	}, nil
}

func TestPipeline_ProduceConsumeIngest(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	testTopic := "test-ingest-" + uuid.NewString()[:8]

	// Setup Kafka
	kafkaCfg := &config.KafkaConfig{
		Brokers:       strings.Split(getEnvOrDefault("KAFKA_BROKERS", "localhost:9092"), ","),
		ConsumerGroup: "test-pipeline-" + uuid.NewString()[:8],
	}

	producer, err := kafka.NewProducer(kafkaCfg)
	require.NoError(t, err, "create producer")
	defer producer.Close()

	adminClient, err := kgo.NewClient(kgo.SeedBrokers(kafkaCfg.Brokers...))
	require.NoError(t, err, "create admin client")
	defer adminClient.Close()

	admin := kadm.NewClient(adminClient)
	_, err = admin.CreateTopic(ctx, 1, 1, nil, testTopic)
	require.NoError(t, err, "create topic")

	// Setup Qdrant
	qdrantHost := getEnvOrDefault("QDRANT_HOST", "localhost")
	qdrantPort, _ := strconv.Atoi(getEnvOrDefault("QDRANT_PORT", "6334"))

	qdrantCfg := &config.QdrantConfig{
		Host:       qdrantHost,
		Port:       qdrantPort,
		Collection: testCollection,
	}

	qdrantClient, err := platformqdrant.New(qdrantCfg)
	require.NoError(t, err, "create qdrant client")
	defer qdrantClient.Close()

	_ = qdrantClient.Conn().DeleteCollection(ctx, testCollection)
	err = qdrantClient.Conn().CreateCollection(ctx, &qdrant.CreateCollection{
		CollectionName: testCollection,
		VectorsConfig: qdrant.NewVectorsConfig(&qdrant.VectorParams{
			Size:     testDimension,
			Distance: qdrant.Distance_Cosine,
		}),
	})
	require.NoError(t, err, "create test collection")
	defer qdrantClient.Conn().DeleteCollection(ctx, testCollection)

	store := vectorstore.NewQdrantStore(qdrantClient, qdrantCfg)

	// Setup Repository
	repo := NewRepository(testDB)

	// Setup stubs
	embedStub := &stubEmbedService{}
	pokemonStub := &stubPokemonService{}

	// Create service and handler
	svc := NewService(embedStub, store, pokemonStub, repo)
	handler := NewHandler(svc)

	// Track if handler was called
	handlerCalled := make(chan string, 1)
	wrappedHandler := func(ctx context.Context, payload []byte) error {
		err := handler.HandleIngest(ctx, payload)
		if err == nil {
			var event IngestionEvent
			json.Unmarshal(payload, &event)
			handlerCalled <- event.ID
		}
		return err
	}

	// Create consumer
	consumer, err := kafka.NewConsumer(kafkaCfg, map[string]kafka.Handler{
		testTopic: wrappedHandler,
	})
	require.NoError(t, err, "create consumer")

	// Run consumer in background
	consumerCtx, cancelConsumer := context.WithCancel(ctx)
	consumerDone := make(chan error, 1)
	go func() {
		consumerDone <- consumer.Run(consumerCtx)
	}()

	// Give consumer time to join group
	time.Sleep(2 * time.Second)

	// Produce a message
	testPokemonID := "test-pokemon-" + uuid.NewString()[:8]
	event := IngestionEvent{
		Type: DocumentTypePokemon,
		ID:   testPokemonID,
	}
	payload, err := json.Marshal(event)
	require.NoError(t, err)

	err = producer.Produce(ctx, testTopic, []byte(testPokemonID), payload)
	require.NoError(t, err, "produce message")

	// Wait for handler to process
	select {
	case processedID := <-handlerCalled:
		assert.Equal(t, testPokemonID, processedID)
	case <-time.After(10 * time.Second):
		t.Fatal("timeout waiting for handler to process message")
	}

	// Stop consumer
	cancelConsumer()
	<-consumerDone

	// Verify pokemon stub was called
	assert.Equal(t, testPokemonID, pokemonStub.calledWith)

	// Verify document in Postgres
	doc, err := repo.FindByRef(ctx, DocumentTypePokemon, testPokemonID)
	require.NoError(t, err, "find document in postgres")
	assert.Equal(t, testPokemonID, doc.ExternalID)
	assert.Equal(t, DocumentTypePokemon, doc.DocumentType)

	// Verify vectors in Qdrant
	reference := NewDocumentID(DocumentTypePokemon, testPokemonID)
	results, err := store.Search(ctx, []float32{1.0, 0.0, 0.0, 0.0}, 10, &vectorstore.Filter{
		StringFilters: []vectorstore.StringFilter{
			{Field: "reference", Value: reference, Op: vectorstore.FilterAND},
		},
	})
	require.NoError(t, err, "search qdrant")
	assert.Len(t, results, 1)
	assert.Equal(t, reference, results[0].Payload["reference"])

	// Cleanup
	cleanupTestData(t, testPokemonID)
}

func getEnvOrDefault(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return fallback
}
