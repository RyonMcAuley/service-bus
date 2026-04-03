package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/chi"
)

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

	err := h.store.Enqueue(r.Context(), qName, []byte(body))
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, "message queued")
}

func (h *Handler) Receive(w http.ResponseWriter, r *http.Request) {
	qName := chi.URLParam(r, "queue")
	if qName == "" {
		writeJSON(w, http.StatusBadRequest, "queue name required")
		return
	}

	msg, err := h.store.Receive(r.Context(), qName)

	if strings.Contains(err.Error(), "no rows in result set") {
		writeJSON(w, http.StatusNoContent, "nothing in queue")
		return
	} else if err != nil {
		fmt.Println(err)
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
