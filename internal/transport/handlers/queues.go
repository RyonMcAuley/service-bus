package handlers

import (
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
	if strings.Contains(err.Error(), "UNIQUE constraint failed") {
		writeJSON(w, http.StatusConflict, "queue already exists")
		return
	} else if err != nil {
		writeJSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, "queue created")
}
