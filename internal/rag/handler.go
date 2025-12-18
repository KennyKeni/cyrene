package rag

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type ChatRequest struct {
	Message string `json:"message"`
	User    string `json:"user"`
}

type ChatResponse struct {
	Response string `json:"response"`
}

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /", h.chat)
	mux.HandleFunc("POST /stream/", h.chatStream)
	return mux
}

// @Summary      Chat with Pokemon knowledge base
// @Description  Query the Pokemon RAG system with a message
// @Tags         chat
// @Accept       json
// @Produce      json
// @Param        request  body      ChatRequest   true  "Chat request"
// @Success      200      {object}  ChatResponse
// @Failure      400      {string}  string  "invalid request body / message is required / user is required"
// @Failure      500      {string}  string  "internal server error"
// @Router       /chat/ [post]
func (h *Handler) chat(w http.ResponseWriter, r *http.Request) {
	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Message == "" {
		http.Error(w, "message is required", http.StatusBadRequest)
		return
	}
	if req.User == "" {
		http.Error(w, "user is required", http.StatusBadRequest)
		return
	}

	response, err := h.service.Chat(r.Context(), req.Message, req.User)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ChatResponse{Response: response})
}

// @Summary      Stream chat with Pokemon knowledge base
// @Description  Query the Pokemon RAG system with streaming response (SSE)
// @Tags         chat
// @Accept       json
// @Produce      text/event-stream
// @Param        request  body      ChatRequest   true  "Chat request"
// @Success      200      {string}  string  "SSE stream"
// @Failure      400      {string}  string  "invalid request body / message is required / user is required"
// @Failure      500      {string}  string  "internal server error"
// @Router       /chat/stream [post]
func (h *Handler) chatStream(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Message == "" {
		http.Error(w, "message is required", http.StatusBadRequest)
		return
	}
	if req.User == "" {
		http.Error(w, "user is required", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	err := h.service.ChatStream(r.Context(), req.Message, req.User, func(chunk string) error {
		_, err := fmt.Fprintf(w, "data: %s\n\n", chunk)
		if err != nil {
			return err
		}
		flusher.Flush()
		return nil
	})

	if err != nil {
		fmt.Fprintf(w, "event: error\ndata: %s\n\n", err.Error())
		flusher.Flush()
		return
	}

	fmt.Fprintf(w, "data: [DONE]\n\n")
	flusher.Flush()
}
