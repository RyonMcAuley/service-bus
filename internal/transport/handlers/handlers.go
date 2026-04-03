package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/RyonMcAuley/servicebus/internal/store"
)

type Handler struct {
	store store.MessageStore
}

type Response struct {
	Message string `json:"message"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type MessageResponse struct {
	ID        string `json:"id"`
	Body      string `json:"body"`
	LockToken string `json:"lockToken"`
}

func NewHandler(s store.MessageStore) Handler {
	return Handler{store: s}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
