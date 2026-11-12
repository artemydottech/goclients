package handlers

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/artemydottech/goclients/internal/models"
)

type ClientSearcher interface {
	SearchClients(context.Context, models.ClientSearch) ([]models.Client, error)
}

type ClientSearchHandler struct{ service ClientSearcher }

func NewClientSearchHandler(s ClientSearcher) *ClientSearchHandler {
	return &ClientSearchHandler{service: s}
}

func (h *ClientSearchHandler) Search(w http.ResponseWriter, r *http.Request) {
	f, ok := parseClientSearch(w, r)
	if !ok {
		return
	}
	items, err := h.service.SearchClients(r.Context(), f)
	var invalid models.ValidationError
	if errors.As(err, &invalid) {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err != nil {
		slog.ErrorContext(r.Context(), "client search failed", "err", err)
		writeError(w, http.StatusInternalServerError, "failed to search clients")
		return
	}
	writeJSON(w, items)
}

func parseClientSearch(w http.ResponseWriter, r *http.Request) (models.ClientSearch, bool) {
	page, ok := parsePage(w, r)
	if !ok {
		return models.ClientSearch{}, false
	}
	id, err := strconv.Atoi(r.URL.Query().Get("company_id"))
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "invalid company_id")
		return models.ClientSearch{}, false
	}
	return models.ClientSearch{CompanyID: id, Query: r.URL.Query().Get("q"), Tag: r.URL.Query().Get("tag"), Phone: r.URL.Query().Get("phone"), Page: page}, true
}

func (h *ClientSearchHandler) Export(w http.ResponseWriter, r *http.Request) {
	f, ok := parseClientSearch(w, r)
	if !ok {
		return
	}
	items, err := h.service.SearchClients(r.Context(), f)
	var invalid models.ValidationError
	if errors.As(err, &invalid) {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err != nil {
		slog.ErrorContext(r.Context(), "client export failed", "err", err)
		writeError(w, http.StatusInternalServerError, "failed to export clients")
		return
	}
	rows := make([][]string, 0, len(items))
	for _, c := range items {
		rows = append(rows, []string{strconv.Itoa(c.ID), strconv.Itoa(c.CompanyID), c.Name, c.Phone, c.Email, c.Comment})
	}
	writeCSV(w, "clients.csv", []string{"id", "company_id", "name", "phone", "email", "comment"}, rows)
}
