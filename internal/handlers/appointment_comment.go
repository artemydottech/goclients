package handlers

import (
	"context"
	"errors"
	"github.com/artemydottech/goclients/internal/models"
	"log/slog"
	"net/http"
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
		writeError(w, 400, "comment is required; use an empty string to clear it")
		return
	}
	err := h.service.UpdateComment(r.Context(), id, *input.Comment)
	var invalid models.ValidationError
	if errors.As(err, &invalid) {
		writeError(w, 400, err.Error())
		return
	}
	if errors.Is(err, models.ErrNotFound) {
		writeError(w, 404, "appointment not found")
		return
	}
	if err != nil {
		slog.ErrorContext(r.Context(), "update appointment comment failed", "err", err)
		writeError(w, 500, "failed to update comment")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
