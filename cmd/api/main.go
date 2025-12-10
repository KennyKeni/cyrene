package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"cyrene/internal/ingest"
	"cyrene/internal/platform/chatstore"
	"cyrene/internal/platform/config"
	"cyrene/internal/platform/genkit"
	"cyrene/internal/platform/kafka"
	"cyrene/internal/platform/postgres"
	"cyrene/internal/platform/qdrant"
	"cyrene/internal/platform/redis"
	"cyrene/internal/platform/server"
	"cyrene/internal/platform/vectorstore"
	"cyrene/internal/pokemon"
	"cyrene/internal/rag"

	_ "cyrene/docs"

	httpSwagger "github.com/swaggo/http-swagger/v2"
)

// @title           Cyrene API
// @version         1.0
// @description     RAG Agent for Pokemon data - indexes Pokemon from Kafka events into Qdrant, answers queries using retrieved context.

// @host            localhost:8080
// @BasePath        /

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	config.Load()
	cfg := config.Get()

	// Platform clients
	pgDB := postgres.New()
	defer func(pgDB *postgres.PostgresDB) {
		err := pgDB.Close()
		if err != nil {
			log.Printf("failed to close postgres: %v", err)
		}
	}(pgDB)

	qdrantClient, err := qdrant.New(&cfg.Qdrant)
	if err != nil {
		log.Fatalf("failed to create qdrant client: %v", err)
	}
	defer func(qdrantClient *qdrant.Client) {
		err := qdrantClient.Close()
		if err != nil {
			log.Printf("failed to close qdrant: %v", err)
		}
	}(qdrantClient)

	if err := qdrantClient.EnsureCollection(ctx, cfg.Qdrant.Collection, uint64(cfg.Qdrant.CollectionDim)); err != nil {
		log.Fatalf("failed to ensure qdrant collection: %v", err)
	}
	if err := qdrantClient.EnsureCollection(ctx, cfg.Qdrant.CacheCollection, uint64(cfg.Qdrant.CacheCollectionDim)); err != nil {
		log.Fatalf("failed to ensure qdrant cache collection: %v", err)
	}

	if err := kafka.EnsureTopics(ctx, cfg.Kafka.Brokers, []string{string(ingest.TopicIngestion)}); err != nil {
		log.Fatalf("failed to ensure kafka topics: %v", err)
	}

	genkitClients, err := genkit.New(ctx, &cfg.Genkit)
	if err != nil {
		log.Fatalf("failed to create genkit clients: %v", err)
	}

	redisClient, err := redis.New(&cfg.Redis)
	if err != nil {
		log.Fatalf("failed to create redis client: %v", err)
	}
	defer redisClient.Client.Close()

	chatStore := chatstore.NewChatStore(redisClient, cfg.ChatStore.MaxMessages, time.Duration(cfg.ChatStore.TTLMinutes)*time.Minute)

	// Services
	vectorStore := vectorstore.NewQdrantStore(qdrantClient, cfg.Qdrant.Collection, int(cfg.Qdrant.CollectionDim))
	cacheStore := vectorstore.NewQdrantStore(qdrantClient, cfg.Qdrant.CacheCollection, int(cfg.Qdrant.CacheCollectionDim))
	pokemonSvc := pokemon.NewService(cfg.PokemonAPI)
	ragSvc := rag.NewService(genkitClients, pokemonSvc, vectorStore, cacheStore, chatStore)
	ingestRepo := ingest.NewRepository(pgDB.DB())
	ingestSvc := ingest.NewService(ragSvc, vectorStore, pokemonSvc, ingestRepo)

	// Handlers
	ingestHandler := ingest.NewHandler(ingestSvc)
	ragHandler := rag.NewHandler(ragSvc)

	// Kafka consumer
	consumer, err := kafka.NewConsumer(&cfg.Kafka, map[string]kafka.Handler{
		string(ingest.TopicIngestion): ingestHandler.HandleKafka,
	})
	if err != nil {
		log.Fatalf("failed to create kafka consumer: %v", err)
	}

	go func() {
		if err := consumer.Run(ctx); err != nil {
			log.Printf("kafka consumer error: %v", err)
		}
	}()

	// Build routes
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", handleHello)
	mux.HandleFunc("GET /health", handleHealth)
	mux.Handle("/ingest/", http.StripPrefix("/ingest", ingestHandler.RegisterRoutes()))
	mux.Handle("/chat/", http.StripPrefix("/chat", ragHandler.RegisterRoutes()))
	mux.Handle("GET /swagger/", httpSwagger.Handler())

	// Create server with middleware
	handler := server.CORSMiddleware(server.TrailingSlashMiddleware(mux))
	srv := server.New(cfg, handler)

	// Graceful shutdown
	done := make(chan bool, 1)
	go gracefulShutdown(ctx, srv, done)

	log.Printf("Server starting on %s", cfg.Server.Addr)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		panic(fmt.Sprintf("http server error: %s", err))
	}

	<-done
	log.Println("Graceful shutdown complete.")
}

func gracefulShutdown(ctx context.Context, apiServer *http.Server, done chan bool) {
	// Listen for the interrupt signal.
	<-ctx.Done()

	log.Println("shutting down gracefully, press Ctrl+C again to force")

	// The context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := apiServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server forced to shutdown with error: %v", err)
	}

	log.Println("Server exiting")

	// Notify the main goroutine that the shutdown is complete
	done <- true
}

func handleHello(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Hello World"})
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
