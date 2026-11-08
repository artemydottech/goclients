package handlers

import (
	"context"
	"fmt"
	"github.com/artemydottech/goclients/internal/models"
	"log/slog"
	"net/http"
	"strconv"
	"time"
)

type AppointmentSearcher interface {
	SearchAppointments(context.Context, models.AppointmentFilter) ([]models.Appointment, error)
}
type AppointmentSearchHandler struct{ service AppointmentSearcher }

func NewAppointmentSearchHandler(s AppointmentSearcher) *AppointmentSearchHandler {
	return &AppointmentSearchHandler{service: s}
}
func parseAppointmentFilter(w http.ResponseWriter, r *http.Request) (models.AppointmentFilter, bool) {
	var f models.AppointmentFilter
	page, ok := parsePage(w, r)
	if !ok {
		return f, false
	}
	f.Page = page
	q := r.URL.Query()
	for _, part := range []struct {
		key  string
		dest *int
	}{{"company_id", &f.CompanyID}, {"client_id", &f.ClientID}, {"employee_id", &f.EmployeeID}, {"service_id", &f.ServiceID}} {
		raw := q.Get(part.key)
		if raw == "" {
			continue
		}
		id, err := strconv.Atoi(raw)
		if err != nil || id <= 0 {
			writeError(w, 400, "invalid "+part.key)
			return f, false
		}
		*part.dest = id
	}
	for _, part := range []struct {
		key  string
		dest *time.Time
	}{{"from", &f.From}, {"to", &f.To}} {
		if raw := q.Get(part.key); raw != "" {
			at, err := time.Parse(time.RFC3339, raw)
			if err != nil {
				writeError(w, 400, "invalid "+part.key)
				return f, false
			}
			*part.dest = at.UTC()
		}
	}
	f.Status = models.AppointmentStatus(q.Get("status"))
	if err := f.Validate(); err != nil {
		writeError(w, 400, err.Error())
		return f, false
	}
	return f, true
}
func (h *AppointmentSearchHandler) Search(w http.ResponseWriter, r *http.Request) {
	f, ok := parseAppointmentFilter(w, r)
	if !ok {
		return
	}
	items, err := h.service.SearchAppointments(r.Context(), f)
	if err != nil {
		slog.ErrorContext(r.Context(), "appointment search failed", "err", err)
		writeError(w, 500, "failed to search appointments")
		return
	}
	writeJSON(w, items)
}
func (h *AppointmentSearchHandler) Export(w http.ResponseWriter, r *http.Request) {
	f, ok := parseAppointmentFilter(w, r)
	if !ok {
		return
	}
	items, err := h.service.SearchAppointments(r.Context(), f)
	if err != nil {
		slog.ErrorContext(r.Context(), "appointment export failed", "err", err)
		writeError(w, 500, "failed to export appointments")
		return
	}
	rows := make([][]string, 0, len(items))
	for _, a := range items {
		rows = append(rows, []string{strconv.Itoa(a.ID), strconv.Itoa(a.CompanyID), strconv.Itoa(a.ClientID), strconv.Itoa(a.EmployeeID), strconv.Itoa(a.ServiceID), a.StartsAt.UTC().Format(time.RFC3339), a.EndsAt.UTC().Format(time.RFC3339), string(a.Status), fmt.Sprintf("%.2f", a.Price), a.Comment})
	}
	writeCSV(w, "appointments.csv", []string{"id", "company_id", "client_id", "employee_id", "service_id", "starts_at", "ends_at", "status", "price", "comment"}, rows)
}
