package ingest

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
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
		slog.Error("failed to unmarshal ingestion event", "error", err)
		return fmt.Errorf("unmarshal payload: %w", err)
	}

	slog.Info("ingestion event received", "type", event.Type, "id", event.ID)

	if err := h.service.Ingest(ctx, event); err != nil {
		slog.Error("ingestion failed", "type", event.Type, "id", event.ID, "error", err)
		return err
	}

	slog.Info("ingestion completed", "type", event.Type, "id", event.ID)
	return nil
}

// @Summary      Ingest document
// @Description  Index a Pokemon or Move document into the vector store
// @Tags         ingest
// @Accept       json
// @Produce      json
// @Param        event  body      IngestionEvent  true  "Ingestion event"
// @Success      202    {object}  map[string]string
// @Failure      400    {string}  string  "invalid request body"
// @Failure      500    {string}  string  "internal server error"
// @Router       /ingest/ [post]
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
