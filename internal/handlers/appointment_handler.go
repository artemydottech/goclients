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

type AppointmentServ interface {
	Book(a models.Appointment) (int64, error)
	GetAllAppointments(ctx context.Context) ([]models.Appointment, error)
	GetAppointmentsByEmployee(ctx context.Context, employeeID int, from, to time.Time) ([]models.Appointment, error)
	GetAppointmentsByClient(ctx context.Context, clientID int) ([]models.Appointment, error)
	GetAppointmentById(ctx context.Context, id int) (models.Appointment, error)
	SetStatus(ctx context.Context, id int, status models.AppointmentStatus) error
	Reschedule(id int, startsAt time.Time, employeeID int) (models.Appointment, error)
	GetClientStats(ctx context.Context, clientID int) (models.ClientStats, error)
	DeleteAppointmentById(ctx context.Context, id int) error
}

type AppointmentHandler struct {
	service AppointmentServ
}

func NewAppointmentHandler(s AppointmentServ) *AppointmentHandler {
	return &AppointmentHandler{service: s}
}

func (h *AppointmentHandler) CreateAppointment(w http.ResponseWriter, r *http.Request) {
	var input models.Appointment

	if !decodeJSON(w, r, &input) {
		return
	}

	id, err := h.service.Book(input)
	var validationErr models.ValidationError
	if errors.As(err, &validationErr) {
		writeError(w, http.StatusBadRequest, validationErr.Error())
		return
	}
	if err != nil {
		slog.ErrorContext(r.Context(), "CreateAppointment failed", "err", err)
		writeError(w, http.StatusInternalServerError, "failed to save appointment")
		return
	}

	writeStatusJSON(w, http.StatusCreated, map[string]int64{"id": id})
}

func (h *AppointmentHandler) GetAllAppointments(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	if raw := query.Get("client_id"); raw != "" {
		clientID, err := strconv.Atoi(raw)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid client_id")
			return
		}

		appointments, err := h.service.GetAppointmentsByClient(r.Context(), clientID)
		if err != nil {
			slog.ErrorContext(r.Context(), "GetAppointmentsByClient failed", "err", err)
			writeError(w, http.StatusInternalServerError, "failed to load appointments")
			return
		}

		writeJSON(w, appointments)
		return
	}

	if raw := query.Get("employee_id"); raw != "" {
		employeeID, err := strconv.Atoi(raw)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid employee_id")
			return
		}

		from, to, err := parseRange(query.Get("from"), query.Get("to"))
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}

		appointments, err := h.service.GetAppointmentsByEmployee(r.Context(), employeeID, from, to)
		if err != nil {
			slog.ErrorContext(r.Context(), "GetAppointmentsByEmployee failed", "err", err)
			writeError(w, http.StatusInternalServerError, "failed to load appointments")
			return
		}

		writeJSON(w, appointments)
		return
	}

	appointments, err := h.service.GetAllAppointments(r.Context())
	if err != nil {
		slog.ErrorContext(r.Context(), "GetAllAppointments failed", "err", err)
		writeError(w, http.StatusInternalServerError, "failed to load appointments")
		return
	}

	writeJSON(w, appointments)
}

func parseRange(rawFrom, rawTo string) (time.Time, time.Time, error) {
	from := time.Now().UTC()
	to := from.AddDate(0, 0, 7)

	if rawFrom != "" {
		parsed, err := time.Parse(time.RFC3339, rawFrom)
		if err != nil {
			return time.Time{}, time.Time{}, errors.New("parameter from must be in RFC3339 format")
		}
		from = parsed.UTC()
		to = from.AddDate(0, 0, 7)
	}

	if rawTo != "" {
		parsed, err := time.Parse(time.RFC3339, rawTo)
		if err != nil {
			return time.Time{}, time.Time{}, errors.New("parameter to must be in RFC3339 format")
		}
		to = parsed.UTC()
	}

	if !to.After(from) {
		return time.Time{}, time.Time{}, errors.New("period end must be after start")
	}

	return from, to, nil
}

func (h *AppointmentHandler) GetAppointmentById(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}

	appointment, err := h.service.GetAppointmentById(r.Context(), id)
	if errors.Is(err, models.ErrNotFound) {
		writeError(w, http.StatusNotFound, "appointment not found")
		return
	}
	if err != nil {
		slog.ErrorContext(r.Context(), "GetAppointmentById failed", "err", err)
		writeError(w, http.StatusInternalServerError, "request failed")
		return
	}

	writeJSON(w, appointment)
}

type statusInput struct {
	Status models.AppointmentStatus `json:"status"`
}

func (h *AppointmentHandler) SetStatus(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}

	var input statusInput
	if !decodeJSON(w, r, &input) {
		return
	}

	err := h.service.SetStatus(r.Context(), id, input.Status)
	var validationErr models.ValidationError
	if errors.As(err, &validationErr) {
		writeError(w, http.StatusBadRequest, validationErr.Error())
		return
	}
	if errors.Is(err, models.ErrNotFound) {
		writeError(w, http.StatusNotFound, "appointment not found")
		return
	}
	if err != nil {
		slog.ErrorContext(r.Context(), "SetStatus failed", "err", err)
		writeError(w, http.StatusInternalServerError, "failed to change status")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *AppointmentHandler) DeleteAppointment(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}

	err := h.service.DeleteAppointmentById(r.Context(), id)
	if errors.Is(err, models.ErrNotFound) {
		writeError(w, http.StatusNotFound, "appointment not found")
		return
	}
	if err != nil {
		slog.ErrorContext(r.Context(), "DeleteAppointment failed", "err", err)
		writeError(w, http.StatusInternalServerError, "failed to delete")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

type rescheduleInput struct {
	StartsAt   time.Time `json:"starts_at"`
	EmployeeID int       `json:"employee_id"`
}

func (h *AppointmentHandler) Reschedule(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}

	var input rescheduleInput
	if !decodeJSON(w, r, &input) {
		return
	}

	appointment, err := h.service.Reschedule(id, input.StartsAt, input.EmployeeID)
	var validationErr models.ValidationError
	if errors.As(err, &validationErr) {
		writeError(w, http.StatusBadRequest, validationErr.Error())
		return
	}
	if errors.Is(err, models.ErrNotFound) {
		writeError(w, http.StatusNotFound, "appointment not found")
		return
	}
	if err != nil {
		slog.ErrorContext(r.Context(), "Reschedule failed", "err", err)
		writeError(w, http.StatusInternalServerError, "failed to reschedule appointment")
		return
	}

	writeJSON(w, appointment)
}

func (h *AppointmentHandler) GetClientStats(w http.ResponseWriter, r *http.Request) {
	clientID, ok := pathID(w, r)
	if !ok {
		return
	}

	stats, err := h.service.GetClientStats(r.Context(), clientID)
	if errors.Is(err, models.ErrNotFound) {
		writeError(w, http.StatusNotFound, "client not found")
		return
	}
	if err != nil {
		slog.ErrorContext(r.Context(), "GetClientStats failed", "err", err)
		writeError(w, http.StatusInternalServerError, "failed to compute stats")
		return
	}

	writeJSON(w, stats)
}
