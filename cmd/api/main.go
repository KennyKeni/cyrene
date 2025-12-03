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

	"cyrene/internal/platform/config"
	"cyrene/internal/platform/server"
)

func main() {
	config.Load()
	cfg := config.Get()

	// Build routes
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", handleHello)
	mux.HandleFunc("GET /health", handleHealth)

	// TODO: Wire feature handlers here
	// Example:
	// queryHandler := query.NewHandler(querySvc)
	// queryHandler.RegisterRoutes(mux)

	// Create server with middleware
	handler := server.CORSMiddleware(mux)
	srv := server.New(cfg, handler)

	// Graceful shutdown
	done := make(chan bool, 1)
	go gracefulShutdown(srv, done)

	log.Printf("Server starting on %s", cfg.Server.Addr)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		panic(fmt.Sprintf("http server error: %s", err))
	}

	<-done
	log.Println("Graceful shutdown complete.")
}

func gracefulShutdown(srv *http.Server, done chan bool) {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	<-ctx.Done()
	log.Println("shutting down gracefully, press Ctrl+C again to force")
	stop()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown with error: %v", err)
	}

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
