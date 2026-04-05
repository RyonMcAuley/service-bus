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

	r.Post("/queues/{queue}", h.CreateQueue)
	r.Get("/queues/{queue}", h.Peek)
	r.Get("/queues/stats", h.GetStats)

	r.Post("/message/{queue}", h.Enqueue)
	r.Get("/message/{queue}", h.Receive)
	r.Post("/ack", h.Ack)
	r.Post("/nack", h.Nack)

	return r
}
