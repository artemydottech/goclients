package handlers

import (
	"database/sql"
	"encoding/json"
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
		http.Error(w, "Неправильный ID", http.StatusBadRequest)
		return
	}

	var input employeeServicesInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err = h.service.SetEmployeeServices(employeeID, input.ServiceIDs)
	var validationErr models.ValidationError
	if errors.As(err, &validationErr) {
		http.Error(w, validationErr.Error(), http.StatusBadRequest)
		return
	}
	if errors.Is(err, sql.ErrNoRows) {
		http.Error(w, "Сотрудник не найден", http.StatusNotFound)
		return
	}
	if err != nil {
		log.Printf("SetEmployeeServices: %v", err)
		http.Error(w, "Ошибка сохранения услуг сотрудника", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *AssignmentHandler) GetEmployeeServices(w http.ResponseWriter, r *http.Request) {
	employeeID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "Неправильный ID", http.StatusBadRequest)
		return
	}

	services, err := h.service.GetServicesByEmployee(employeeID)
	if err != nil {
		log.Printf("GetEmployeeServices: %v", err)
		http.Error(w, "Ошибка запроса услуг сотрудника", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(services)
}

func (h *AssignmentHandler) GetServiceEmployees(w http.ResponseWriter, r *http.Request) {
	serviceID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "Неправильный ID", http.StatusBadRequest)
		return
	}

	employees, err := h.service.GetEmployeesByService(serviceID)
	if err != nil {
		log.Printf("GetServiceEmployees: %v", err)
		http.Error(w, "Ошибка запроса мастеров услуги", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(employees)
}
