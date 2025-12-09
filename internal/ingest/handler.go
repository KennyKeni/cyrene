package ingest

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /{$}", h.ingest)
	return mux
}

func (h *Handler) HandleKafka(ctx context.Context, payload []byte) error {
	var event IngestionEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return fmt.Errorf("unmarshal payload: %w", err)
	}
	return h.service.Ingest(ctx, event)
}

func (h *Handler) ingest(w http.ResponseWriter, r *http.Request) {
	var event IngestionEvent
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.service.Ingest(r.Context(), event); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{"status": "accepted"})
}
