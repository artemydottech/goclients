package handlers

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/artemydottech/goclients/internal/models"
)

type ClientTagger interface {
	Get(context.Context, int) ([]string, error)
	Replace(context.Context, int, []string) error
}

type ClientTagHandler struct{ service ClientTagger }

func NewClientTagHandler(s ClientTagger) *ClientTagHandler { return &ClientTagHandler{service: s} }

func tagError(w http.ResponseWriter, r *http.Request, err error) bool {
	if err == nil {
		return false
	}
	var invalid models.ValidationError
	switch {
	case errors.As(err, &invalid):
		writeError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, models.ErrNotFound):
		writeError(w, http.StatusNotFound, "client not found")
	default:
		slog.ErrorContext(r.Context(), "client tags failed", "err", err)
		writeError(w, http.StatusInternalServerError, "failed to load or save client tags")
	}
	return true
}

func (h *ClientTagHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	tags, err := h.service.Get(r.Context(), id)
	if tagError(w, r, err) {
		return
	}
	writeJSON(w, tags)
}

func (h *ClientTagHandler) Replace(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	var body struct {
		Tags *[]string `json:"tags"`
	}
	if !decodeJSON(w, r, &body) {
		return
	}
	if body.Tags == nil {
		writeError(w, http.StatusBadRequest, "tags must be an explicit array")
		return
	}
	if tagError(w, r, h.service.Replace(r.Context(), id, *body.Tags)) {
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
