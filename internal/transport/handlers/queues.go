package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi"
)

func (h *Handler) CreateQueue(w http.ResponseWriter, r *http.Request) {
	qName := chi.URLParam(r, "queue")
	if qName == "" {
		writeJSON(w, http.StatusBadRequest, "queue name required")
		return
	}
	maxDelivery := r.URL.Query().Get("maxDelivery")
	deliveryInt := 5

	if maxDelivery != "" {
		i, err := strconv.Atoi(maxDelivery)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, "invalid maxDelivery")
			return
		} else {
			deliveryInt = i
		}
	}
	err := h.store.CreateQueue(r.Context(), qName, deliveryInt)
	if err != nil && strings.Contains(err.Error(), "UNIQUE constraint failed") {
		writeJSON(w, http.StatusConflict, "queue already exists")
		return
	} else if err != nil {
		writeJSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, "queue created")
}

func (h *Handler) Peek(w http.ResponseWriter, r *http.Request) {
	qName := chi.URLParam(r, "queue")
	if qName == "" {
		writeJSON(w, http.StatusBadRequest, "queue name required")
		return
	}

	msg, err := h.store.Peek(r.Context(), qName)
	if err != nil {
		writeJSON(w, http.StatusNoContent, "nothing to peek")
		return
	}

	writeJSON(w, http.StatusOK, MessageResponse{
		ID:   msg.ID,
		Body: string(msg.Body),
	})
}

func (h *Handler) GetStats(w http.ResponseWriter, r *http.Request) {
	queues, err := h.store.ListQueues(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, "unable to retrieve queues")
		return
	}

	result := []QueueResponse{}

	for _, queue := range queues {
		stats, err := h.store.GetStats(r.Context(), queue.Name)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, "unable to get stats")
			return
		}
		// add stats to response
		result = append(result, QueueResponse{
			QueueName:         queue.Name,
			ActiveMessages:    stats.ActiveMessages,
			AvailableMessages: stats.AvailableMessages,
			DLQCount:          stats.DLQCount,
		})
	}

	writeJSON(w, http.StatusOK, result)
}

func (h *Handler) DeleteQueue(w http.ResponseWriter, r *http.Request) {
	queueName := chi.URLParam(r, "queue")
	if queueName == "" {
		writeJSON(w, http.StatusBadRequest, "queue name required")
		return
	}
	stats, err := h.store.GetStats(r.Context(), queueName)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, "error validating queue stats")
		return
	}

	messagesInQueue := stats.ActiveMessages + stats.AvailableMessages

	force := r.URL.Query().Get("force")

	if messagesInQueue > 0 && force != "true" {
		writeJSON(w, http.StatusConflict, "queue has messages, use ?force=true to delete")
		return
	}

	err = h.store.DeleteQueue(r.Context(), queueName)
	if err != nil {
		fmt.Println(err)
		writeJSON(w, http.StatusInternalServerError, "error deleting queue")
		return
	}

	writeJSON(w, http.StatusNoContent, "queue deleted")
}
