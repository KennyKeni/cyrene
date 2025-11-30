package server

import (
	"fmt"
	"net/http"
	"time"

	"cyrene/internal/config"
	"cyrene/internal/database"
)

type Server struct {
	port int

	db database.Service
}

func NewServer() *http.Server {
	cfg := config.Get()
	NewServer := &Server{
		port: cfg.Port,
		db:   database.New(),
	}

	// Declare Server config
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", NewServer.port),
		Handler:      NewServer.RegisterRoutes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	return server
}
