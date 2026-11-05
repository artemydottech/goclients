package repository

import (
	"context"
	"github.com/artemydottech/goclients/internal/models"
)

func (r *AppointmentRepository) SearchAppointments(ctx context.Context, f models.AppointmentFilter) ([]models.Appointment, error) {
	query := `SELECT id, company_id, client_id, employee_id, service_id, starts_at, ends_at, status, comment, price FROM appointments WHERE company_id = ?`
	args := []any{f.CompanyID}
	for _, part := range []struct {
		column string
		id     int
	}{{"client_id", f.ClientID}, {"employee_id", f.EmployeeID}, {"service_id", f.ServiceID}} {
		if part.id > 0 {
			query += " AND " + part.column + " = ?"
			args = append(args, part.id)
		}
	}
	if f.Status != "" {
		query += " AND status = ?"
		args = append(args, string(f.Status))
	}
	if !f.From.IsZero() {
		query += " AND starts_at >= ?"
		args = append(args, f.From.UTC().Format(timeLayout))
	}
	if !f.To.IsZero() {
		query += " AND starts_at < ?"
		args = append(args, f.To.UTC().Format(timeLayout))
	}
	query, args = paginate(query+" ORDER BY starts_at, id", args, f.Page)
	return r.query(ctx, query, args...)
}
