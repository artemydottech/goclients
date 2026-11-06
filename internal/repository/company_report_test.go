package repository

import (
	"context"
	"github.com/artemydottech/goclients/internal/models"
	"testing"
	"time"
)

func TestCompanyReportRevenueAndBoundaries(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	company, employee, item, client := bookingFixture(t, db)
	repo := NewAppointmentRepository(db)
	at := time.Date(2026, 11, 6, 0, 0, 0, 0, time.UTC)
	for _, status := range []models.AppointmentStatus{models.AppointmentPending, models.AppointmentConfirmed, models.AppointmentCompleted, models.AppointmentCancelled, models.AppointmentNoShow} {
		if _, err := repo.Create(ctx, models.Appointment{CompanyID: company, EmployeeID: employee, ServiceID: item, ClientID: client, StartsAt: at, EndsAt: at.Add(time.Hour), Status: status, Price: 1250}); err != nil {
			t.Fatal(err)
		}
	}
	q := models.ReportQuery{CompanyID: company, From: at, To: at.Add(24 * time.Hour)}
	got, err := repo.GetCompanyReport(ctx, q)
	if err != nil {
		t.Fatal(err)
	}
	s := got.Summary
	if s.Appointments != 5 || s.Active != 2 || s.Completed != 1 || s.Cancelled != 1 || s.NoShows != 1 || s.UniqueClients != 1 || s.Revenue != 1250 || s.ExpectedRevenue != 2500 {
		t.Fatalf("summary: %+v", s)
	}
	q.From = q.To
	q.To = q.To.Add(time.Hour)
	got, err = repo.GetCompanyReport(ctx, q)
	if err != nil || got.Summary.Appointments != 0 {
		t.Fatalf("outside: %+v %v", got, err)
	}
	q.CompanyID = company + 99
	got, err = repo.GetCompanyReport(ctx, q)
	if err != nil || got.Summary.Revenue != 0 {
		t.Fatalf("foreign: %+v %v", got, err)
	}
}
