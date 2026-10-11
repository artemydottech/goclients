package handlers

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/artemydottech/goclients/internal/models"
)

type TimeOffServ interface {
	CreateTimeOff(employeeID int, t models.TimeOff) (int64, error)
	GetTimeOffByEmployee(employeeID int) ([]models.TimeOff, error)
	DeleteTimeOffById(id int) error
}

type TimeOffHandler struct {
	service TimeOffServ
}

func NewTimeOffHandler(s TimeOffServ) *TimeOffHandler {
	return &TimeOffHandler{service: s}
}

func (h *TimeOffHandler) CreateTimeOff(w http.ResponseWriter, r *http.Request) {
	employeeID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid ID")
		return
	}

	var input models.TimeOff
	if !decodeJSON(w, r, &input) {
		return
	}

	id, err := h.service.CreateTimeOff(employeeID, input)
	var validationErr models.ValidationError
	if errors.As(err, &validationErr) {
		writeError(w, http.StatusBadRequest, validationErr.Error())
		return
	}
	if errors.Is(err, models.ErrNotFound) {
		writeError(w, http.StatusNotFound, "employee not found")
		return
	}
	if err != nil {
		slog.ErrorContext(r.Context(), "CreateTimeOff failed", "err", err)
		writeError(w, http.StatusInternalServerError, "failed to save time off")
		return
	}

	writeStatusJSON(w, http.StatusCreated, map[string]int64{"id": id})
}

func (h *TimeOffHandler) GetTimeOff(w http.ResponseWriter, r *http.Request) {
	employeeID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid ID")
		return
	}

	periods, err := h.service.GetTimeOffByEmployee(employeeID)
	if err != nil {
		slog.ErrorContext(r.Context(), "GetTimeOff failed", "err", err)
		writeError(w, http.StatusInternalServerError, "failed to load time off")
		return
	}

	writeJSON(w, periods)
}

func (h *TimeOffHandler) DeleteTimeOff(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid ID")
		return
	}

	err = h.service.DeleteTimeOffById(id)
	if errors.Is(err, models.ErrNotFound) {
		writeError(w, http.StatusNotFound, "time off not found")
		return
	}
	if err != nil {
		slog.ErrorContext(r.Context(), "DeleteTimeOff failed", "err", err)
		writeError(w, http.StatusInternalServerError, "failed to delete")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
