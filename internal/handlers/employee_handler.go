package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

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
		http.Error(w, "Ошибка сохранения сотрудника", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]int64{"id": id})
}

func (h *EmployeeHandler) GetAllEmployees(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	employees, err := h.service.GetAllEmployees()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(employees)
}

func (h *EmployeeHandler) GetEmployeeById(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idString := strings.TrimPrefix(r.URL.Path, "/employees/")

	if idString == "" {
		http.Error(w, "ID обязателен", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idString)

	if err != nil {
		http.Error(w, "Неправильный ID", http.StatusBadRequest)
		return
	}

	employee, err := h.service.GetEmployeeById(id)
	if errors.Is(err, sql.ErrNoRows) {
		http.Error(w, "Сотрудник не найден", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "Ошибка запроса", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(employee)
}

func (h *EmployeeHandler) DeleteEmployee(w http.ResponseWriter, r *http.Request) {
	if r.Method != "DELETE" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idString := strings.TrimPrefix(r.URL.Path, "/employees/")
	if idString == "" {
		http.Error(w, "ID обязателен", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idString)
	if err != nil {
		http.Error(w, "Неправильный ID", http.StatusBadRequest)
		return
	}

	err = h.service.DeleteEmployeeById(id)
	if errors.Is(err, sql.ErrNoRows) {
		http.Error(w, "Сотрудник не найден", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "Ошибка удаления", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent)
}
