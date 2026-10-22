package handlers

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/artemydottech/goclients/internal/models"
)

type CatalogServ interface {
	CreateService(ctx context.Context, s models.Service) (int64, error)
	GetAllServices(ctx context.Context, page models.Page) ([]models.Service, error)
	GetServicesByCompany(ctx context.Context, companyID int, page models.Page) ([]models.Service, error)
	GetServiceById(ctx context.Context, id int) (models.Service, error)
	UpdateService(ctx context.Context, id int, s models.Service) error
	DeleteServiceById(ctx context.Context, id int) error
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

	id, err := h.service.CreateService(r.Context(), input)
	var validationErr models.ValidationError
	if errors.As(err, &validationErr) {
		writeError(w, http.StatusBadRequest, validationErr.Error())
		return
	}
	if err != nil {
		slog.ErrorContext(r.Context(), "CreateService failed", "err", err)
		writeError(w, http.StatusInternalServerError, "failed to save service")
		return
	}

	writeStatusJSON(w, http.StatusCreated, map[string]int64{"id": id})
}

func (h *ServiceHandler) GetAllServices(w http.ResponseWriter, r *http.Request) {
	page, ok := parsePage(w, r)
	if !ok {
		return
	}

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
		services, err = h.service.GetServicesByCompany(r.Context(), companyID, page)
	} else {
		services, err = h.service.GetAllServices(r.Context(), page)
	}

	if err != nil {
		slog.ErrorContext(r.Context(), "GetAllServices failed", "err", err)
		writeError(w, http.StatusInternalServerError, "failed to load services")
		return
	}

	writeJSON(w, services)
}

func (h *ServiceHandler) GetServiceById(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}

	item, err := h.service.GetServiceById(r.Context(), id)
	if errors.Is(err, models.ErrNotFound) {
		writeError(w, http.StatusNotFound, "service not found")
		return
	}
	if err != nil {
		slog.ErrorContext(r.Context(), "GetServiceById failed", "err", err)
		writeError(w, http.StatusInternalServerError, "request failed")
		return
	}

	writeJSON(w, item)
}

func (h *ServiceHandler) UpdateService(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}

	var input models.Service
	if !decodeJSON(w, r, &input) {
		return
	}

	err := h.service.UpdateService(r.Context(), id, input)
	var validationErr models.ValidationError
	if errors.As(err, &validationErr) {
		writeError(w, http.StatusBadRequest, validationErr.Error())
		return
	}
	if errors.Is(err, models.ErrNotFound) {
		writeError(w, http.StatusNotFound, "service not found")
		return
	}
	if err != nil {
		slog.ErrorContext(r.Context(), "UpdateService failed", "err", err)
		writeError(w, http.StatusInternalServerError, "failed to update service")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ServiceHandler) DeleteService(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}

	err := h.service.DeleteServiceById(r.Context(), id)
	if errors.Is(err, models.ErrNotFound) {
		writeError(w, http.StatusNotFound, "service not found")
		return
	}
	if err != nil {
		slog.ErrorContext(r.Context(), "DeleteService failed", "err", err)
		writeError(w, http.StatusInternalServerError, "failed to delete")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
