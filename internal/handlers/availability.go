package handlers

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/artemydottech/goclients/internal/models"
)

type AvailabilitySearcher interface {
	Search(context.Context, models.AvailabilityQuery) ([]models.OpenSlot, error)
}

type AvailabilityHandler struct{ service AvailabilitySearcher }

func NewAvailabilityHandler(s AvailabilitySearcher) *AvailabilityHandler {
	return &AvailabilityHandler{service: s}
}

func (h *AvailabilityHandler) Search(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	q := models.AvailabilityQuery{ServiceID: id, Days: 7, Step: 15, Limit: 50}
	date, err := time.Parse("2006-01-02", r.URL.Query().Get("date"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "date must be YYYY-MM-DD")
		return
	}
	q.Date = date
	for _, field := range []struct {
		name string
		dest *int
	}{{"employee_id", &q.EmployeeID}, {"days", &q.Days}, {"step", &q.Step}, {"limit", &q.Limit}} {
		if raw := r.URL.Query().Get(field.name); raw != "" {
			value, err := strconv.Atoi(raw)
			if err != nil || value <= 0 {
				writeError(w, http.StatusBadRequest, "invalid "+field.name)
				return
			}
			*field.dest = value
		}
	}
	slots, err := h.service.Search(r.Context(), q)
	var invalid models.ValidationError
	switch {
	case errors.As(err, &invalid):
		writeError(w, http.StatusBadRequest, err.Error())
		return
	case errors.Is(err, models.ErrNotFound):
		writeError(w, http.StatusNotFound, "service not found")
		return
	case err != nil:
		slog.ErrorContext(r.Context(), "availability search failed", "err", err)
		writeError(w, http.StatusInternalServerError, "failed to search availability")
		return
	}
	writeJSON(w, slots)
}
