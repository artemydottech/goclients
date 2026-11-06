package handlers

import (
	"context"
	"errors"
	"github.com/artemydottech/goclients/internal/models"
	"log/slog"
	"net/http"
	"time"
)

type CompanyReporter interface {
	GetCompanyReport(context.Context, models.ReportQuery) (models.CompanyReport, error)
}
type CompanyReportHandler struct{ service CompanyReporter }

func NewCompanyReportHandler(s CompanyReporter) *CompanyReportHandler {
	return &CompanyReportHandler{service: s}
}
func (h *CompanyReportHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	from, e1 := time.Parse(time.RFC3339, r.URL.Query().Get("from"))
	to, e2 := time.Parse(time.RFC3339, r.URL.Query().Get("to"))
	if e1 != nil || e2 != nil {
		writeError(w, 400, "from and to must be RFC3339 timestamps")
		return
	}
	result, err := h.service.GetCompanyReport(r.Context(), models.ReportQuery{CompanyID: id, From: from, To: to})
	var invalid models.ValidationError
	if errors.As(err, &invalid) {
		writeError(w, 400, err.Error())
		return
	}
	if errors.Is(err, models.ErrNotFound) {
		writeError(w, 404, "company not found")
		return
	}
	if err != nil {
		slog.ErrorContext(r.Context(), "company report failed", "err", err)
		writeError(w, 500, "failed to load report")
		return
	}
	writeJSON(w, result)
}
