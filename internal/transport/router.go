package transport

import (
	"github.com/RyonMcAuley/servicebus/internal/store"
	"github.com/RyonMcAuley/servicebus/internal/transport/handlers"
	"github.com/go-chi/chi"
	"net/http"
)

func NewRouter(s store.MessageStore) http.Handler {

	r := chi.NewRouter()

	h := handlers.NewHandler(s)

	r.Get("/health", h.Health)

	r.Post("/queues/{queue}", h.CreateQueue)
	r.Get("/queues/{queue}", h.Peek)
	r.Get("/stats", h.GetStats)
	r.Delete("/queues/{queue}", h.DeleteQueue)

	r.Post("/queues/{queue}/message", h.Enqueue)
	r.Get("/queues/{queue}/message", h.Receive)
	r.Post("/ack", h.Ack)
	r.Post("/nack", h.Nack)

	return r
}
