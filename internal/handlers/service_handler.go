package handlers

import (
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

	if !decodeJSON(w, r, &input) {
		return
	}

	id, err := h.service.CreateService(input)
	var validationErr models.ValidationError
	if errors.As(err, &validationErr) {
		writeError(w, http.StatusBadRequest, validationErr.Error())
		return
	}
	if err != nil {
		log.Printf("CreateService: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to save service")
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
			writeError(w, http.StatusBadRequest, "invalid company_id")
			return
		}
		services, err = h.service.GetServicesByCompany(companyID)
	} else {
		services, err = h.service.GetAllServices()
	}

	if err != nil {
		log.Printf("GetAllServices: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to load services")
		return
	}

	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(services)
}

func (h *ServiceHandler) GetServiceById(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid ID")
		return
	}

	item, err := h.service.GetServiceById(id)
	if errors.Is(err, models.ErrNotFound) {
		writeError(w, http.StatusNotFound, "service not found")
		return
	}
	if err != nil {
		log.Printf("GetServiceById: %v", err)
		writeError(w, http.StatusInternalServerError, "request failed")
		return
	}

	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(item)
}

func (h *ServiceHandler) DeleteService(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid ID")
		return
	}

	err = h.service.DeleteServiceById(id)
	if errors.Is(err, models.ErrNotFound) {
		writeError(w, http.StatusNotFound, "service not found")
		return
	}
	if err != nil {
		log.Printf("DeleteService: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to delete")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
