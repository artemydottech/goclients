package handlers

import (
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/artemydottech/goclients/internal/models"
)

type AssignmentServ interface {
	SetEmployeeServices(employeeID int, serviceIDs []int) error
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
	employeeID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid ID")
		return
	}

	var input employeeServicesInput
	if !decodeJSON(w, r, &input) {
		return
	}

	err = h.service.SetEmployeeServices(employeeID, input.ServiceIDs)
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
		log.Printf("SetEmployeeServices: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to save employee services")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *AssignmentHandler) GetEmployeeServices(w http.ResponseWriter, r *http.Request) {
	employeeID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid ID")
		return
	}

	services, err := h.service.GetServicesByEmployee(employeeID)
	if err != nil {
		log.Printf("GetEmployeeServices: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to load employee services")
		return
	}

	writeJSON(w, services)
}

func (h *AssignmentHandler) GetServiceEmployees(w http.ResponseWriter, r *http.Request) {
	serviceID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid ID")
		return
	}

	employees, err := h.service.GetEmployeesByService(serviceID)
	if err != nil {
		log.Printf("GetServiceEmployees: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to load service employees")
		return
	}

	writeJSON(w, employees)
}
