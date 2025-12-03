package server

import (
	"net/http"
	"time"

	"cyrene/internal/platform/config"
)

// New creates a new HTTP server with the given handler.
// Platform only provides the server setup - routes are registered by features.
func New(cfg *config.Config, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:         cfg.Server.Addr,
		Handler:      handler,
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}
}
