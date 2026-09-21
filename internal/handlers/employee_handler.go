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

type EmployeeServ interface {
	CreateEmployee(e models.Employee) (int64, error)
	GetAllEmployees() ([]models.Employee, error)
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

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	id, err := h.service.CreateEmployee(input)
	var validationErr models.ValidationError
	if errors.As(err, &validationErr) {
		http.Error(w, validationErr.Error(), http.StatusBadRequest)
		return
	}
	if err != nil {
		http.Error(w, "failed to save employee", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]int64{"id": id})
}

func (h *EmployeeHandler) GetAllEmployees(w http.ResponseWriter, r *http.Request) {
	employees, err := h.service.GetAllEmployees()
	if err != nil {
		log.Printf("GetAllEmployees: %v", err)
		http.Error(w, "failed to load employees", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(employees)
}

func (h *EmployeeHandler) GetEmployeeById(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))

	if err != nil {
		http.Error(w, "invalid ID", http.StatusBadRequest)
		return
	}

	employee, err := h.service.GetEmployeeById(id)
	if errors.Is(err, sql.ErrNoRows) {
		http.Error(w, "employee not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "request failed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(employee)
}

func (h *EmployeeHandler) DeleteEmployee(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "invalid ID", http.StatusBadRequest)
		return
	}

	err = h.service.DeleteEmployeeById(id)
	if errors.Is(err, sql.ErrNoRows) {
		http.Error(w, "employee not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "failed to delete", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
