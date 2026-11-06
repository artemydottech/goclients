package repository

import (
	"context"
	"database/sql"
	"github.com/artemydottech/goclients/internal/models"
)

func (r *AppointmentRepository) GetCompanyReport(ctx context.Context, q models.ReportQuery) (models.CompanyReport, error) {
	result := models.CompanyReport{CompanyID: q.CompanyID, From: q.From.UTC(), To: q.To.UTC(), Services: []models.ReportGroup{}, Employees: []models.ReportGroup{}}
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return result, err
	}
	defer tx.Rollback()
	s := &result.Summary
	err = tx.QueryRowContext(ctx, `SELECT COUNT(*),
 COALESCE(SUM(status IN ('pending','confirmed')),0),COALESCE(SUM(status='completed'),0),
 COALESCE(SUM(status='cancelled'),0),COALESCE(SUM(status='no_show'),0),COUNT(DISTINCT client_id),
 COALESCE(SUM(CASE WHEN status='completed' THEN price ELSE 0 END),0),
 COALESCE(SUM(CASE WHEN status IN ('pending','confirmed') THEN price ELSE 0 END),0)
 FROM appointments WHERE company_id=? AND starts_at>=? AND starts_at<?`, q.CompanyID, q.From.UTC().Format(timeLayout), q.To.UTC().Format(timeLayout)).Scan(&s.Appointments, &s.Active, &s.Completed, &s.Cancelled, &s.NoShows, &s.UniqueClients, &s.Revenue, &s.ExpectedRevenue)
	if err != nil {
		return result, err
	}
	return result, tx.Commit()
}
