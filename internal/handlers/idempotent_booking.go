package handlers

import (
	"context"
	"errors"
	"github.com/artemydottech/goclients/internal/models"
	"log/slog"
	"net/http"
)

type IdempotentBooker interface {
	Book(context.Context, string, models.Appointment) (models.BookingResult, error)
}
type IdempotentBookingHandler struct{ service IdempotentBooker }

func NewIdempotentBookingHandler(s IdempotentBooker) *IdempotentBookingHandler {
	return &IdempotentBookingHandler{service: s}
}
func (h *IdempotentBookingHandler) Create(w http.ResponseWriter, r *http.Request) {
	var input models.Appointment
	if !decodeJSON(w, r, &input) {
		return
	}
	key := r.Header.Get("Idempotency-Key")
	if _, present := r.Header["Idempotency-Key"]; present && key == "" {
		writeError(w, http.StatusBadRequest, "Idempotency-Key cannot be empty")
		return
	}
	result, err := h.service.Book(r.Context(), key, input)
	var invalid models.ValidationError
	if errors.As(err, &invalid) {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if errors.Is(err, models.ErrConflict) {
		writeError(w, http.StatusConflict, "Idempotency-Key has already been used for a different request")
		return
	}
	if err != nil {
		slog.ErrorContext(r.Context(), "booking failed", "err", err)
		writeError(w, http.StatusInternalServerError, "failed to save appointment")
		return
	}
	if result.Replayed {
		w.Header().Set("Idempotent-Replayed", "true")
	}
	writeStatusJSON(w, http.StatusCreated, map[string]int64{"id": result.ID})
}
