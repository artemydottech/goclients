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

type CatalogServ interface {
	CreateService(s models.Service) (int64, error)
	GetAllServices() ([]models.Service, error)
	GetServicesByCompany(companyID int) ([]models.Service, error)
	GetServiceById(id int) (models.Service, error)
	DeleteServiceById(id int) error
}

type ServiceHandler struct {
	service CatalogServ
}

func NewServiceHandler(s CatalogServ) *ServiceHandler {
	return &ServiceHandler{service: s}
}

func (h *ServiceHandler) CreateService(w http.ResponseWriter, r *http.Request) {
	var input models.Service

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	id, err := h.service.CreateService(input)
	var validationErr models.ValidationError
	if errors.As(err, &validationErr) {
		http.Error(w, validationErr.Error(), http.StatusBadRequest)
		return
	}
	if err != nil {
		log.Printf("CreateService: %v", err)
		http.Error(w, "Ошибка сохранения услуги", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]int64{"id": id})
}

func (h *ServiceHandler) GetAllServices(w http.ResponseWriter, r *http.Request) {
	var (
		services []models.Service
		err      error
	)

	if raw := r.URL.Query().Get("company_id"); raw != "" {
		companyID, convErr := strconv.Atoi(raw)
		if convErr != nil {
			http.Error(w, "Неправильный company_id", http.StatusBadRequest)
			return
		}
		services, err = h.service.GetServicesByCompany(companyID)
	} else {
		services, err = h.service.GetAllServices()
	}

	if err != nil {
		log.Printf("GetAllServices: %v", err)
		http.Error(w, "Ошибка запроса услуг", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(services)
}

func (h *ServiceHandler) GetServiceById(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "Неправильный ID", http.StatusBadRequest)
		return
	}

	item, err := h.service.GetServiceById(id)
	if errors.Is(err, sql.ErrNoRows) {
		http.Error(w, "Услуга не найдена", http.StatusNotFound)
		return
	}
	if err != nil {
		log.Printf("GetServiceById: %v", err)
		http.Error(w, "Ошибка запроса", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(item)
}

func (h *ServiceHandler) DeleteService(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "Неправильный ID", http.StatusBadRequest)
		return
	}

	err = h.service.DeleteServiceById(id)
	if errors.Is(err, sql.ErrNoRows) {
		http.Error(w, "Услуга не найдена", http.StatusNotFound)
		return
	}
	if err != nil {
		log.Printf("DeleteService: %v", err)
		http.Error(w, "Ошибка удаления", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
