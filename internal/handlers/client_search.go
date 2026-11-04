package handlers

import (
	"context"
	"errors"
	"github.com/artemydottech/goclients/internal/models"
	"log/slog"
	"net/http"
	"strconv"
)

type ClientSearcher interface {
	SearchClients(context.Context, models.ClientSearch) ([]models.Client, error)
}
type ClientSearchHandler struct{ service ClientSearcher }

func NewClientSearchHandler(s ClientSearcher) *ClientSearchHandler {
	return &ClientSearchHandler{service: s}
}
func (h *ClientSearchHandler) Search(w http.ResponseWriter, r *http.Request) {
	page, ok := parsePage(w, r)
	if !ok {
		return
	}
	id, err := strconv.Atoi(r.URL.Query().Get("company_id"))
	if err != nil || id <= 0 {
		writeError(w, 400, "invalid company_id")
		return
	}
	items, err := h.service.SearchClients(r.Context(), models.ClientSearch{CompanyID: id, Query: r.URL.Query().Get("q"), Phone: r.URL.Query().Get("phone"), Page: page})
	var invalid models.ValidationError
	if errors.As(err, &invalid) {
		writeError(w, 400, err.Error())
		return
	}
	if err != nil {
		slog.ErrorContext(r.Context(), "client search failed", "err", err)
		writeError(w, 500, "failed to search clients")
		return
	}
	writeJSON(w, items)
}
