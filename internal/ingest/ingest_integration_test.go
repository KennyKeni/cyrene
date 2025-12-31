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

func (s *stubEmbedService) Embed(ctx context.Context, dimensions int, texts ...string) ([][]float32, error) {
	return [][]float32{{1.0, 0.0, 0.0, 0.0}}, nil
}

type stubPokemonService struct {
	speciesCalls []string
	formCalls    []string
	moveCalls    []string
	abilityCalls []string
	itemCalls    []string
	articleCalls []string
}

func (s *stubPokemonService) GetSpecies(ctx context.Context, identifier string) (*pokemon.Species, error) {
	s.speciesCalls = append(s.speciesCalls, identifier)
	return &pokemon.Species{
		ID:   1,
		Name: "Pikachu",
		Slug: "pikachu",
		Forms: []pokemon.SpeciesForm{
			{ID: 25, Slug: "pikachu"},
		},
	}, nil
}

func (s *stubPokemonService) GetPokemon(ctx context.Context, identifier string) (*pokemon.Pokemon, error) {
	s.formCalls = append(s.formCalls, identifier)
	return &pokemon.Pokemon{
		ID:         25,
		Name:       "Pikachu",
		Slug:       "pikachu",
		Generation: 1,
		Form: &pokemon.PokemonForm{
			ID:                 25,
			Name:               "Pikachu",
			FullName:           "Pikachu",
			Slug:               "pikachu",
			BaseHP:             35,
			BaseAttack:         55,
			BaseDefence:        40,
			BaseSpecialAttack:  50,
			BaseSpecialDefence: 50,
			BaseSpeed:          90,
			Types: []pokemon.PokemonFormType{
				{Type: pokemon.NamedResource{ID: 13, Name: "Electric", Slug: "electric"}, Slot: 1},
			},
			Abilities: []pokemon.PokemonFormAbility{
				{Ability: pokemon.NamedResource{ID: 9, Name: "Static", Slug: "static"}, Slot: pokemon.NamedResource{ID: 1, Name: "Slot 1", Slug: "slot-1"}},
			},
		},
	}, nil
}

func (s *stubPokemonService) GetMove(ctx context.Context, identifier string) (*pokemon.Move, error) {
	s.moveCalls = append(s.moveCalls, identifier)
	power := 90
	accuracy := 100
	desc := "A strong electric attack"
	return &pokemon.Move{
		ID:       85,
		Name:     "Thunderbolt",
		Slug:     "thunderbolt",
		Desc:     &desc,
		Type:     pokemon.NamedResource{ID: 13, Name: "Electric", Slug: "electric"},
		Category: pokemon.NamedResource{ID: 2, Name: "Special", Slug: "special"},
		Power:    &power,
		Accuracy: &accuracy,
		PP:       15,
	}, nil
}

func (s *stubPokemonService) GetAbility(ctx context.Context, identifier string) (*pokemon.Ability, error) {
	s.abilityCalls = append(s.abilityCalls, identifier)
	desc := "Contact may cause paralysis"
	return &pokemon.Ability{
		ID:        9,
		Name:      "Static",
		Slug:      "static",
		ShortDesc: &desc,
	}, nil
}

func (s *stubPokemonService) GetItem(ctx context.Context, identifier string) (*pokemon.Item, error) {
	s.itemCalls = append(s.itemCalls, identifier)
	desc := "Restores 20 HP"
	return &pokemon.Item{
		ID:        4,
		Name:      "Potion",
		Slug:      "potion",
		ShortDesc: &desc,
	}, nil
}

func (s *stubPokemonService) GetArticle(ctx context.Context, identifier string) (*pokemon.Article, error) {
	s.articleCalls = append(s.articleCalls, identifier)
	return &pokemon.Article{
		ID:    1,
		Title: "Getting Started",
		Slug:  "getting-started",
		Body:  "Welcome to the guide",
	}, nil
}

func setupTestInfra(t *testing.T, ctx context.Context) (*kafka.Producer, *kadm.Client, *vectorstore.QdrantStore, Repository, func()) {
	t.Helper()

	kafkaCfg := &config.KafkaConfig{
		Brokers:       strings.Split(getEnvOrDefault("KAFKA_BROKERS", "localhost:9092"), ","),
		ConsumerGroup: "test-pipeline-" + uuid.NewString()[:8],
	}

	producer, err := kafka.NewProducer(kafkaCfg)
	require.NoError(t, err, "create producer")

	adminClient, err := kgo.NewClient(kgo.SeedBrokers(kafkaCfg.Brokers...))
	require.NoError(t, err, "create admin client")
	admin := kadm.NewClient(adminClient)

	qdrantHost := getEnvOrDefault("QDRANT_HOST", "localhost")
	qdrantPort, _ := strconv.Atoi(getEnvOrDefault("QDRANT_PORT", "6334"))

	qdrantCfg := &config.QdrantConfig{
		Host:       qdrantHost,
		Port:       qdrantPort,
		Collection: testCollection,
	}

	qdrantClient, err := platformqdrant.New(qdrantCfg)
	require.NoError(t, err, "create qdrant client")

	_ = qdrantClient.Conn().DeleteCollection(ctx, testCollection)
	err = qdrantClient.Conn().CreateCollection(ctx, &qdrant.CreateCollection{
		CollectionName: testCollection,
		VectorsConfig: qdrant.NewVectorsConfig(&qdrant.VectorParams{
			Size:     testDimension,
			Distance: qdrant.Distance_Cosine,
		}),
	})
	require.NoError(t, err, "create test collection")

	store := vectorstore.NewQdrantStore(qdrantClient, testCollection, int(testDimension))
	repo := NewRepository(testDB)

	cleanup := func() {
		producer.Close()
		adminClient.Close()
		qdrantClient.Conn().DeleteCollection(ctx, testCollection)
		qdrantClient.Close()
	}

	return producer, admin, store, repo, cleanup
}

func TestIngest_FormCreate(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	testTopic := "test-ingest-" + uuid.NewString()[:8]
	producer, admin, store, repo, cleanup := setupTestInfra(t, ctx)
	defer cleanup()

	_, err := admin.CreateTopic(ctx, 1, 1, nil, testTopic)
	require.NoError(t, err, "create topic")

	embedStub := &stubEmbedService{}
	pokemonStub := &stubPokemonService{}

	svc := NewService(embedStub, store, pokemonStub, repo)
	handler := NewHandler(svc)

	handlerCalled := make(chan string, 1)
	wrappedHandler := func(ctx context.Context, payload []byte) error {
		err := handler.HandleKafka(ctx, payload)
		if err == nil {
			var event IngestionEvent
			json.Unmarshal(payload, &event)
			handlerCalled <- event.EntityID
		}
		return err
	}

	kafkaCfg := &config.KafkaConfig{
		Brokers:       strings.Split(getEnvOrDefault("KAFKA_BROKERS", "localhost:9092"), ","),
		ConsumerGroup: "test-pipeline-" + uuid.NewString()[:8],
	}

	consumer, err := kafka.NewConsumer(kafkaCfg, map[string]kafka.Handler{
		testTopic: wrappedHandler,
	})
	require.NoError(t, err, "create consumer")

	consumerCtx, cancelConsumer := context.WithCancel(ctx)
	consumerDone := make(chan error, 1)
	go func() {
		consumerDone <- consumer.Run(consumerCtx)
	}()

	time.Sleep(2 * time.Second)

	testFormID := "25"
	event := IngestionEvent{
		EntityType: EntityTypeForm,
		EntityID:   testFormID,
		Operation:  OperationCreate,
	}
	payload, err := json.Marshal(event)
	require.NoError(t, err)

	err = producer.Produce(ctx, testTopic, []byte(testFormID), payload)
	require.NoError(t, err, "produce message")

	select {
	case processedID := <-handlerCalled:
		assert.Equal(t, testFormID, processedID)
	case <-time.After(10 * time.Second):
		t.Fatal("timeout waiting for handler to process message")
	}

	cancelConsumer()
	<-consumerDone

	assert.Contains(t, pokemonStub.formCalls, testFormID)

	doc, err := repo.FindByRef(ctx, DocumentTypeForm, testFormID)
	require.NoError(t, err, "find document in postgres")
	assert.Equal(t, testFormID, doc.ExternalID)
	assert.Equal(t, DocumentTypeForm, doc.DocumentType)

	results, err := store.Search(ctx, []float32{1.0, 0.0, 0.0, 0.0}, 10, &vectorstore.Filter{
		StringFilters: []vectorstore.StringFilter{
			{Field: formIDKey, Value: testFormID, Op: vectorstore.FilterAND},
		},
	})
	require.NoError(t, err, "search qdrant")
	assert.Len(t, results, 1)

	cleanupTestData(t, testFormID)
}

func TestIngest_SpeciesCreate_IngestsForms(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	testTopic := "test-ingest-" + uuid.NewString()[:8]
	producer, admin, store, repo, cleanup := setupTestInfra(t, ctx)
	defer cleanup()

	_, err := admin.CreateTopic(ctx, 1, 1, nil, testTopic)
	require.NoError(t, err, "create topic")

	embedStub := &stubEmbedService{}
	pokemonStub := &stubPokemonService{}

	svc := NewService(embedStub, store, pokemonStub, repo)
	handler := NewHandler(svc)

	handlerCalled := make(chan string, 1)
	wrappedHandler := func(ctx context.Context, payload []byte) error {
		err := handler.HandleKafka(ctx, payload)
		if err == nil {
			var event IngestionEvent
			json.Unmarshal(payload, &event)
			handlerCalled <- event.EntityID
		}
		return err
	}

	kafkaCfg := &config.KafkaConfig{
		Brokers:       strings.Split(getEnvOrDefault("KAFKA_BROKERS", "localhost:9092"), ","),
		ConsumerGroup: "test-pipeline-" + uuid.NewString()[:8],
	}

	consumer, err := kafka.NewConsumer(kafkaCfg, map[string]kafka.Handler{
		testTopic: wrappedHandler,
	})
	require.NoError(t, err, "create consumer")

	consumerCtx, cancelConsumer := context.WithCancel(ctx)
	consumerDone := make(chan error, 1)
	go func() {
		consumerDone <- consumer.Run(consumerCtx)
	}()

	time.Sleep(2 * time.Second)

	testSpeciesID := "1"
	event := IngestionEvent{
		EntityType: EntityTypeSpecies,
		EntityID:   testSpeciesID,
		Operation:  OperationCreate,
	}
	payload, err := json.Marshal(event)
	require.NoError(t, err)

	err = producer.Produce(ctx, testTopic, []byte(testSpeciesID), payload)
	require.NoError(t, err, "produce message")

	select {
	case processedID := <-handlerCalled:
		assert.Equal(t, testSpeciesID, processedID)
	case <-time.After(10 * time.Second):
		t.Fatal("timeout waiting for handler to process message")
	}

	cancelConsumer()
	<-consumerDone

	assert.Contains(t, pokemonStub.speciesCalls, testSpeciesID)
	assert.Contains(t, pokemonStub.formCalls, "25")

	cleanupTestData(t, "25")
}

func TestIngest_MoveCreate(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	testTopic := "test-ingest-" + uuid.NewString()[:8]
	producer, admin, store, repo, cleanup := setupTestInfra(t, ctx)
	defer cleanup()

	_, err := admin.CreateTopic(ctx, 1, 1, nil, testTopic)
	require.NoError(t, err, "create topic")

	embedStub := &stubEmbedService{}
	pokemonStub := &stubPokemonService{}

	svc := NewService(embedStub, store, pokemonStub, repo)
	handler := NewHandler(svc)

	handlerCalled := make(chan string, 1)
	wrappedHandler := func(ctx context.Context, payload []byte) error {
		err := handler.HandleKafka(ctx, payload)
		if err == nil {
			var event IngestionEvent
			json.Unmarshal(payload, &event)
			handlerCalled <- event.EntityID
		}
		return err
	}

	kafkaCfg := &config.KafkaConfig{
		Brokers:       strings.Split(getEnvOrDefault("KAFKA_BROKERS", "localhost:9092"), ","),
		ConsumerGroup: "test-pipeline-" + uuid.NewString()[:8],
	}

	consumer, err := kafka.NewConsumer(kafkaCfg, map[string]kafka.Handler{
		testTopic: wrappedHandler,
	})
	require.NoError(t, err, "create consumer")

	consumerCtx, cancelConsumer := context.WithCancel(ctx)
	consumerDone := make(chan error, 1)
	go func() {
		consumerDone <- consumer.Run(consumerCtx)
	}()

	time.Sleep(2 * time.Second)

	testMoveID := "85"
	event := IngestionEvent{
		EntityType: EntityTypeMove,
		EntityID:   testMoveID,
		Operation:  OperationCreate,
	}
	payload, err := json.Marshal(event)
	require.NoError(t, err)

	err = producer.Produce(ctx, testTopic, []byte(testMoveID), payload)
	require.NoError(t, err, "produce message")

	select {
	case processedID := <-handlerCalled:
		assert.Equal(t, testMoveID, processedID)
	case <-time.After(10 * time.Second):
		t.Fatal("timeout waiting for handler to process message")
	}

	cancelConsumer()
	<-consumerDone

	assert.Contains(t, pokemonStub.moveCalls, testMoveID)

	doc, err := repo.FindByRef(ctx, DocumentTypeMove, testMoveID)
	require.NoError(t, err, "find document in postgres")
	assert.Equal(t, testMoveID, doc.ExternalID)

	cleanupTestData(t, testMoveID)
}

func TestIngest_AbilityCreate(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	testTopic := "test-ingest-" + uuid.NewString()[:8]
	producer, admin, store, repo, cleanup := setupTestInfra(t, ctx)
	defer cleanup()

	_, err := admin.CreateTopic(ctx, 1, 1, nil, testTopic)
	require.NoError(t, err, "create topic")

	embedStub := &stubEmbedService{}
	pokemonStub := &stubPokemonService{}

	svc := NewService(embedStub, store, pokemonStub, repo)
	handler := NewHandler(svc)

	handlerCalled := make(chan string, 1)
	wrappedHandler := func(ctx context.Context, payload []byte) error {
		err := handler.HandleKafka(ctx, payload)
		if err == nil {
			var event IngestionEvent
			json.Unmarshal(payload, &event)
			handlerCalled <- event.EntityID
		}
		return err
	}

	kafkaCfg := &config.KafkaConfig{
		Brokers:       strings.Split(getEnvOrDefault("KAFKA_BROKERS", "localhost:9092"), ","),
		ConsumerGroup: "test-pipeline-" + uuid.NewString()[:8],
	}

	consumer, err := kafka.NewConsumer(kafkaCfg, map[string]kafka.Handler{
		testTopic: wrappedHandler,
	})
	require.NoError(t, err, "create consumer")

	consumerCtx, cancelConsumer := context.WithCancel(ctx)
	consumerDone := make(chan error, 1)
	go func() {
		consumerDone <- consumer.Run(consumerCtx)
	}()

	time.Sleep(2 * time.Second)

	testAbilityID := "9"
	event := IngestionEvent{
		EntityType: EntityTypeAbility,
		EntityID:   testAbilityID,
		Operation:  OperationCreate,
	}
	payload, err := json.Marshal(event)
	require.NoError(t, err)

	err = producer.Produce(ctx, testTopic, []byte(testAbilityID), payload)
	require.NoError(t, err, "produce message")

	select {
	case processedID := <-handlerCalled:
		assert.Equal(t, testAbilityID, processedID)
	case <-time.After(10 * time.Second):
		t.Fatal("timeout waiting for handler to process message")
	}

	cancelConsumer()
	<-consumerDone

	assert.Contains(t, pokemonStub.abilityCalls, testAbilityID)

	doc, err := repo.FindByRef(ctx, DocumentTypeAbility, testAbilityID)
	require.NoError(t, err, "find document in postgres")
	assert.Equal(t, testAbilityID, doc.ExternalID)

	cleanupTestData(t, testAbilityID)
}

func TestIngest_ItemCreate(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	testTopic := "test-ingest-" + uuid.NewString()[:8]
	producer, admin, store, repo, cleanup := setupTestInfra(t, ctx)
	defer cleanup()

	_, err := admin.CreateTopic(ctx, 1, 1, nil, testTopic)
	require.NoError(t, err, "create topic")

	embedStub := &stubEmbedService{}
	pokemonStub := &stubPokemonService{}

	svc := NewService(embedStub, store, pokemonStub, repo)
	handler := NewHandler(svc)

	handlerCalled := make(chan string, 1)
	wrappedHandler := func(ctx context.Context, payload []byte) error {
		err := handler.HandleKafka(ctx, payload)
		if err == nil {
			var event IngestionEvent
			json.Unmarshal(payload, &event)
			handlerCalled <- event.EntityID
		}
		return err
	}

	kafkaCfg := &config.KafkaConfig{
		Brokers:       strings.Split(getEnvOrDefault("KAFKA_BROKERS", "localhost:9092"), ","),
		ConsumerGroup: "test-pipeline-" + uuid.NewString()[:8],
	}

	consumer, err := kafka.NewConsumer(kafkaCfg, map[string]kafka.Handler{
		testTopic: wrappedHandler,
	})
	require.NoError(t, err, "create consumer")

	consumerCtx, cancelConsumer := context.WithCancel(ctx)
	consumerDone := make(chan error, 1)
	go func() {
		consumerDone <- consumer.Run(consumerCtx)
	}()

	time.Sleep(2 * time.Second)

	testItemID := "4"
	event := IngestionEvent{
		EntityType: EntityTypeItem,
		EntityID:   testItemID,
		Operation:  OperationCreate,
	}
	payload, err := json.Marshal(event)
	require.NoError(t, err)

	err = producer.Produce(ctx, testTopic, []byte(testItemID), payload)
	require.NoError(t, err, "produce message")

	select {
	case processedID := <-handlerCalled:
		assert.Equal(t, testItemID, processedID)
	case <-time.After(10 * time.Second):
		t.Fatal("timeout waiting for handler to process message")
	}

	cancelConsumer()
	<-consumerDone

	assert.Contains(t, pokemonStub.itemCalls, testItemID)

	doc, err := repo.FindByRef(ctx, DocumentTypeItem, testItemID)
	require.NoError(t, err, "find document in postgres")
	assert.Equal(t, testItemID, doc.ExternalID)

	cleanupTestData(t, testItemID)
}

func TestIngest_ArticleCreate(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	testTopic := "test-ingest-" + uuid.NewString()[:8]
	producer, admin, store, repo, cleanup := setupTestInfra(t, ctx)
	defer cleanup()

	_, err := admin.CreateTopic(ctx, 1, 1, nil, testTopic)
	require.NoError(t, err, "create topic")

	embedStub := &stubEmbedService{}
	pokemonStub := &stubPokemonService{}

	svc := NewService(embedStub, store, pokemonStub, repo)
	handler := NewHandler(svc)

	handlerCalled := make(chan string, 1)
	wrappedHandler := func(ctx context.Context, payload []byte) error {
		err := handler.HandleKafka(ctx, payload)
		if err == nil {
			var event IngestionEvent
			json.Unmarshal(payload, &event)
			handlerCalled <- event.EntityID
		}
		return err
	}

	kafkaCfg := &config.KafkaConfig{
		Brokers:       strings.Split(getEnvOrDefault("KAFKA_BROKERS", "localhost:9092"), ","),
		ConsumerGroup: "test-pipeline-" + uuid.NewString()[:8],
	}

	consumer, err := kafka.NewConsumer(kafkaCfg, map[string]kafka.Handler{
		testTopic: wrappedHandler,
	})
	require.NoError(t, err, "create consumer")

	consumerCtx, cancelConsumer := context.WithCancel(ctx)
	consumerDone := make(chan error, 1)
	go func() {
		consumerDone <- consumer.Run(consumerCtx)
	}()

	time.Sleep(2 * time.Second)

	testArticleID := "1"
	event := IngestionEvent{
		EntityType: EntityTypeArticle,
		EntityID:   testArticleID,
		Operation:  OperationCreate,
	}
	payload, err := json.Marshal(event)
	require.NoError(t, err)

	err = producer.Produce(ctx, testTopic, []byte(testArticleID), payload)
	require.NoError(t, err, "produce message")

	select {
	case processedID := <-handlerCalled:
		assert.Equal(t, testArticleID, processedID)
	case <-time.After(10 * time.Second):
		t.Fatal("timeout waiting for handler to process message")
	}

	cancelConsumer()
	<-consumerDone

	assert.Contains(t, pokemonStub.articleCalls, testArticleID)

	doc, err := repo.FindByRef(ctx, DocumentTypeArticle, testArticleID)
	require.NoError(t, err, "find document in postgres")
	assert.Equal(t, testArticleID, doc.ExternalID)

	cleanupTestData(t, testArticleID)
}

func TestIngest_FormDelete(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	testTopic := "test-ingest-" + uuid.NewString()[:8]
	producer, admin, store, repo, cleanup := setupTestInfra(t, ctx)
	defer cleanup()

	_, err := admin.CreateTopic(ctx, 1, 1, nil, testTopic)
	require.NoError(t, err, "create topic")

	embedStub := &stubEmbedService{}
	pokemonStub := &stubPokemonService{}

	svc := NewService(embedStub, store, pokemonStub, repo)
	handler := NewHandler(svc)

	testFormID := "25"
	createEvent := IngestionEvent{
		EntityType: EntityTypeForm,
		EntityID:   testFormID,
		Operation:  OperationCreate,
	}
	err = svc.Ingest(ctx, createEvent)
	require.NoError(t, err, "create form for delete test")

	doc, err := repo.FindByRef(ctx, DocumentTypeForm, testFormID)
	require.NoError(t, err, "verify form exists before delete")
	assert.Equal(t, testFormID, doc.ExternalID)

	handlerCalled := make(chan string, 1)
	wrappedHandler := func(ctx context.Context, payload []byte) error {
		err := handler.HandleKafka(ctx, payload)
		if err == nil {
			var event IngestionEvent
			json.Unmarshal(payload, &event)
			handlerCalled <- event.EntityID
		}
		return err
	}

	kafkaCfg := &config.KafkaConfig{
		Brokers:       strings.Split(getEnvOrDefault("KAFKA_BROKERS", "localhost:9092"), ","),
		ConsumerGroup: "test-pipeline-" + uuid.NewString()[:8],
	}

	consumer, err := kafka.NewConsumer(kafkaCfg, map[string]kafka.Handler{
		testTopic: wrappedHandler,
	})
	require.NoError(t, err, "create consumer")

	consumerCtx, cancelConsumer := context.WithCancel(ctx)
	consumerDone := make(chan error, 1)
	go func() {
		consumerDone <- consumer.Run(consumerCtx)
	}()

	time.Sleep(2 * time.Second)

	deleteEvent := IngestionEvent{
		EntityType: EntityTypeForm,
		EntityID:   testFormID,
		Operation:  OperationDelete,
	}
	payload, err := json.Marshal(deleteEvent)
	require.NoError(t, err)

	err = producer.Produce(ctx, testTopic, []byte(testFormID), payload)
	require.NoError(t, err, "produce delete message")

	select {
	case processedID := <-handlerCalled:
		assert.Equal(t, testFormID, processedID)
	case <-time.After(10 * time.Second):
		t.Fatal("timeout waiting for handler to process delete message")
	}

	cancelConsumer()
	<-consumerDone

	_, err = repo.FindByRef(ctx, DocumentTypeForm, testFormID)
	assert.ErrorIs(t, err, ErrNotFound)

	results, err := store.Search(ctx, []float32{1.0, 0.0, 0.0, 0.0}, 10, &vectorstore.Filter{
		StringFilters: []vectorstore.StringFilter{
			{Field: formIDKey, Value: testFormID, Op: vectorstore.FilterAND},
		},
	})
	require.NoError(t, err, "search qdrant after delete")
	assert.Len(t, results, 0)
}

func getEnvOrDefault(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return fallback
}
