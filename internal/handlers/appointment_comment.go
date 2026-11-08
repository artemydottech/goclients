package handlers

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/artemydottech/goclients/internal/models"
)

type AppointmentCommenter interface {
	UpdateComment(context.Context, int, string) error
}
type AppointmentCommentHandler struct{ service AppointmentCommenter }

func NewAppointmentCommentHandler(s AppointmentCommenter) *AppointmentCommentHandler {
	return &AppointmentCommentHandler{service: s}
}
func (h *AppointmentCommentHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	var input struct {
		Comment *string `json:"comment"`
	}
	if !decodeJSON(w, r, &input) {
		return
	}
	if input.Comment == nil {
		writeError(w, http.StatusBadRequest, "comment is required; use an empty string to clear it")
		return
	}
	err := h.service.UpdateComment(r.Context(), id, *input.Comment)
	var invalid models.ValidationError
	if errors.As(err, &invalid) {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if errors.Is(err, models.ErrNotFound) {
		writeError(w, http.StatusNotFound, "appointment not found")
		return
	}
	if err != nil {
		slog.ErrorContext(r.Context(), "update appointment comment failed", "err", err)
		writeError(w, http.StatusInternalServerError, "failed to update comment")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
