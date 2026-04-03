package handlers

import (
	"net/http"
	"strconv"

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
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("invalid maxDelivery"))
			return
		} else {
			deliveryInt = i
		}
	}
	err := h.store.CreateQueue(r.Context(), qName, deliveryInt)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("unable to create queue\n"))
		w.Write([]byte(err.Error()))
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("queue created"))
}

func (h *Handler) Enqueue(w http.ResponseWriter, r *http.Request) {
	qName := chi.URLParam(r, "queue")
	if qName == "" {
		writeJSON(w, http.StatusBadRequest, "queue name required")
		return
	}

	body := r.URL.Query().Get("body")

	if body == "" {
		writeJSON(w, http.StatusBadRequest, "must include message body")
		return
	}

	h.store.Enqueue(r.Context(), qName, []byte(body))
	writeJSON(w, http.StatusCreated, "message queued")
}

func (h *Handler) Receive(w http.ResponseWriter, r *http.Request) {
	qName := chi.URLParam(r, "queue")
	if qName == "" {
		writeJSON(w, http.StatusBadRequest, "queue name required")
		return
	}

	msg, err := h.store.Receive(r.Context(), qName)

	if err != nil {
		writeJSON(w, http.StatusInternalServerError, "unable to receive a message")
		return
	}

	writeJSON(w, http.StatusOK, MessageResponse{
		ID:        msg.ID,
		Body:      string(msg.Body),
		LockToken: *msg.LockToken})
}

func (h *Handler) Ack(w http.ResponseWriter, r *http.Request) {
	lockToken := r.URL.Query().Get("lockToken")

	if lockToken == "" {
		writeJSON(w, http.StatusBadRequest, "must pass lock token")
		return
	}

	err := h.store.Ack(r.Context(), lockToken)

	if err != nil {
		writeJSON(w, http.StatusInternalServerError, "failed to ack with supplied token")
		return
	}

	writeJSON(w, http.StatusNoContent, "message completed")
}

func (h *Handler) Nack(w http.ResponseWriter, r *http.Request) {
	lockToken := r.URL.Query().Get("lockToken")

	if lockToken == "" {
		writeJSON(w, http.StatusBadRequest, "must pass lock token")
		return
	}

	err := h.store.Nack(r.Context(), lockToken)

	if err != nil {
		writeJSON(w, http.StatusInternalServerError, "failed to nack with supplied token")
		return
	}

	writeJSON(w, http.StatusNoContent, "message abandoned")
}
