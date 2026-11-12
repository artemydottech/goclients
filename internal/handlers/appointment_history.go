package handlers

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/artemydottech/goclients/internal/models"
)

type AppointmentHistorian interface {
	List(context.Context, int, models.Page) ([]models.AppointmentEvent, error)
}

type AppointmentHistoryHandler struct{ service AppointmentHistorian }

func NewAppointmentHistoryHandler(s AppointmentHistorian) *AppointmentHistoryHandler {
	return &AppointmentHistoryHandler{service: s}
}

func (h *AppointmentHistoryHandler) List(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	page, ok := parsePage(w, r)
	if !ok {
		return
	}
	if page.Limit == 0 {
		page.Limit = 50
	}
	events, err := h.service.List(r.Context(), id, page)
	var invalid models.ValidationError
	switch {
	case errors.As(err, &invalid):
		writeError(w, http.StatusBadRequest, err.Error())
		return
	case errors.Is(err, models.ErrNotFound):
		writeError(w, http.StatusNotFound, "appointment not found")
		return
	case err != nil:
		slog.ErrorContext(r.Context(), "appointment history failed", "err", err)
		writeError(w, http.StatusInternalServerError, "failed to load appointment history")
		return
	}
	writeJSON(w, events)
}
