package handlers

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/artemydottech/goclients/internal/models"
)

type AssignmentServ interface {
	SetEmployeeServices(ctx context.Context, employeeID int, serviceIDs []int) error
	GetServicesByEmployee(employeeID int) ([]models.Service, error)
	GetEmployeesByService(serviceID int) ([]models.Employee, error)
}

type AssignmentHandler struct {
	service AssignmentServ
}

func NewAssignmentHandler(s AssignmentServ) *AssignmentHandler {
	return &AssignmentHandler{service: s}
}

type employeeServicesInput struct {
	ServiceIDs []int `json:"service_ids"`
}

func (h *AssignmentHandler) SetEmployeeServices(w http.ResponseWriter, r *http.Request) {
	employeeID, ok := pathID(w, r)
	if !ok {
		return
	}

	var input employeeServicesInput
	if !decodeJSON(w, r, &input) {
		return
	}

	err := h.service.SetEmployeeServices(r.Context(), employeeID, input.ServiceIDs)
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
		slog.ErrorContext(r.Context(), "SetEmployeeServices failed", "err", err)
		writeError(w, http.StatusInternalServerError, "failed to save employee services")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *AssignmentHandler) GetEmployeeServices(w http.ResponseWriter, r *http.Request) {
	employeeID, ok := pathID(w, r)
	if !ok {
		return
	}

	services, err := h.service.GetServicesByEmployee(employeeID)
	if err != nil {
		slog.ErrorContext(r.Context(), "GetEmployeeServices failed", "err", err)
		writeError(w, http.StatusInternalServerError, "failed to load employee services")
		return
	}

	writeJSON(w, services)
}

func (h *AssignmentHandler) GetServiceEmployees(w http.ResponseWriter, r *http.Request) {
	serviceID, ok := pathID(w, r)
	if !ok {
		return
	}

	employees, err := h.service.GetEmployeesByService(serviceID)
	if err != nil {
		slog.ErrorContext(r.Context(), "GetServiceEmployees failed", "err", err)
		writeError(w, http.StatusInternalServerError, "failed to load service employees")
		return
	}

	writeJSON(w, employees)
}
