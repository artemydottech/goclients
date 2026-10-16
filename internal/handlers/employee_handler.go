package handlers

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/artemydottech/goclients/internal/models"
)

type EmployeeServ interface {
	CreateEmployee(ctx context.Context, e models.Employee) (int64, error)
	GetAllEmployees(ctx context.Context) ([]models.Employee, error)
	GetEmployeeById(id int) (models.Employee, error)
	DeleteEmployeeById(id int) error
}

type EmployeeHandler struct {
	service EmployeeServ
}

func NewEmployeeHandler(s EmployeeServ) *EmployeeHandler {
	return &EmployeeHandler{service: s}
}

func (h *EmployeeHandler) CreateEmployee(w http.ResponseWriter, r *http.Request) {
	var input models.Employee

	if !decodeJSON(w, r, &input) {
		return
	}

	id, err := h.service.CreateEmployee(r.Context(), input)
	var validationErr models.ValidationError
	if errors.As(err, &validationErr) {
		writeError(w, http.StatusBadRequest, validationErr.Error())
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to save employee")
		return
	}

	writeStatusJSON(w, http.StatusCreated, map[string]int64{"id": id})
}

func (h *EmployeeHandler) GetAllEmployees(w http.ResponseWriter, r *http.Request) {
	employees, err := h.service.GetAllEmployees(r.Context())
	if err != nil {
		slog.ErrorContext(r.Context(), "GetAllEmployees failed", "err", err)
		writeError(w, http.StatusInternalServerError, "failed to load employees")
		return
	}

	writeJSON(w, employees)
}

func (h *EmployeeHandler) GetEmployeeById(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}

	employee, err := h.service.GetEmployeeById(id)
	if errors.Is(err, models.ErrNotFound) {
		writeError(w, http.StatusNotFound, "employee not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "request failed")
		return
	}

	writeJSON(w, employee)
}

func (h *EmployeeHandler) DeleteEmployee(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}

	err := h.service.DeleteEmployeeById(id)
	if errors.Is(err, models.ErrNotFound) {
		writeError(w, http.StatusNotFound, "employee not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
