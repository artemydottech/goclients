package handlers

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/artemydottech/goclients/internal/models"
)

type CompanyServ interface {
	CreateCompany(ctx context.Context, c models.Company) (int64, error)
	GetAllCompanies(ctx context.Context) ([]models.Company, error)
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

	if !decodeJSON(w, r, &input) {
		return
	}

	id, err := h.service.CreateCompany(r.Context(), input)
	var validationErr models.ValidationError
	if errors.As(err, &validationErr) {
		writeError(w, http.StatusBadRequest, validationErr.Error())
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to save company")
		return
	}

	writeStatusJSON(w, http.StatusCreated, map[string]int64{"id": id})
}

func (h *CompanyHandler) GetAllCompanies(w http.ResponseWriter, r *http.Request) {
	companies, err := h.service.GetAllCompanies(r.Context())
	if err != nil {
		slog.ErrorContext(r.Context(), "GetAllCompanies failed", "err", err)
		writeError(w, http.StatusInternalServerError, "failed to load companies")
		return
	}

	writeJSON(w, companies)
}

func (h *CompanyHandler) GetCompanyById(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}

	company, err := h.service.GetCompanyById(id)
	if errors.Is(err, models.ErrNotFound) {
		writeError(w, http.StatusNotFound, "company not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "request failed")
		return
	}

	writeJSON(w, company)
}

func (h *CompanyHandler) DeleteCompany(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}

	err := h.service.DeleteCompanyById(id)
	if errors.Is(err, models.ErrNotFound) {
		writeError(w, http.StatusNotFound, "company not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
