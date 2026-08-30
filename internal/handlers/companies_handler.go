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

type CompanyServ interface {
	CreateCompany(c models.Company) (int64, error)
	GetAllCompanies() ([]models.Company, error)
	GetCompanyById(id int) (models.Company, error)
	DeleteCompanyById(id int) error
}

type CompanyHandler struct {
	service CompanyServ
}

func NewCompanyHandler(s CompanyServ) *CompanyHandler {
	return &CompanyHandler{service: s}
}

func (h *CompanyHandler) CreateCompany(w http.ResponseWriter, r *http.Request) {
	var input models.Company

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	id, err := h.service.CreateCompany(input)
	var validationErr models.ValidationError
	if errors.As(err, &validationErr) {
		http.Error(w, validationErr.Error(), http.StatusBadRequest)
		return
	}
	if err != nil {
		http.Error(w, "Ошибка сохранения компании", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]int64{"id": id})
}

func (h *CompanyHandler) GetAllCompanies(w http.ResponseWriter, r *http.Request) {
	companies, err := h.service.GetAllCompanies()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(companies)
}

func (h *CompanyHandler) GetCompanyById(w http.ResponseWriter, r *http.Request) {
	idString := strings.TrimPrefix(r.URL.Path, "/companies/")

	if idString == "" {
		http.Error(w, "ID обязателен", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idString)

	if err != nil {
		http.Error(w, "Неправильный ID", http.StatusBadRequest)
		return
	}

	company, err := h.service.GetCompanyById(id)
	if errors.Is(err, sql.ErrNoRows) {
		http.Error(w, "Компания не найдена", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "Ошибка запроса", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(company)
}

func (h *CompanyHandler) DeleteCompany(w http.ResponseWriter, r *http.Request) {
	idString := strings.TrimPrefix(r.URL.Path, "/companies/")
	if idString == "" {
		http.Error(w, "ID обязателен", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idString)
	if err != nil {
		http.Error(w, "Неправильный ID", http.StatusBadRequest)
		return
	}

	err = h.service.DeleteCompanyById(id)
	if errors.Is(err, sql.ErrNoRows) {
		http.Error(w, "Компания не найдена", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "Ошибка удаления", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent)
}
