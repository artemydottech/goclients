package repository

import (
	"context"
	"github.com/artemydottech/goclients/internal/models"
	"testing"
	"time"
)

func TestAppointmentSearchCombinesFilters(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	company, employee, item, client := bookingFixture(t, db)
	repo := NewAppointmentRepository(db)
	at := time.Date(2027, 1, 1, 10, 0, 0, 0, time.UTC)
	for _, status := range []models.AppointmentStatus{models.AppointmentPending, models.AppointmentConfirmed} {
		if _, err := repo.Create(ctx, models.Appointment{CompanyID: company, EmployeeID: employee, ServiceID: item, ClientID: client, StartsAt: at, EndsAt: at.Add(time.Hour), Status: status}); err != nil {
			t.Fatal(err)
		}
	}
	got, err := repo.SearchAppointments(ctx, models.AppointmentFilter{CompanyID: company, EmployeeID: employee, ServiceID: item, ClientID: client, Status: models.AppointmentConfirmed})
	if err != nil || len(got) != 1 || got[0].Status != models.AppointmentConfirmed {
		t.Fatalf("search: %+v %v", got, err)
	}
	for _, f := range []models.AppointmentFilter{{CompanyID: company + 99}, {CompanyID: company, ClientID: client + 99}, {CompanyID: company, EmployeeID: employee + 99}, {CompanyID: company, ServiceID: item + 99}} {
		got, err := repo.SearchAppointments(ctx, f)
		if err != nil || len(got) != 0 {
			t.Fatalf("foreign match: %+v %v", got, err)
		}
	}
}
func TestAppointmentSearchHalfOpenRangeAndStablePages(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	company, employee, item, client := bookingFixture(t, db)
	repo := NewAppointmentRepository(db)
	at := time.Date(2027, 1, 1, 10, 0, 0, 0, time.UTC)
	var ids []int64
	for _, start := range []time.Time{at, at, at.Add(time.Hour)} {
		id, err := repo.Create(ctx, models.Appointment{CompanyID: company, EmployeeID: employee, ServiceID: item, ClientID: client, StartsAt: start, EndsAt: start.Add(time.Hour), Status: models.AppointmentPending})
		if err != nil {
			t.Fatal(err)
		}
		ids = append(ids, id)
	}
	f := models.AppointmentFilter{CompanyID: company, From: at, To: at.Add(time.Hour), Page: models.Page{Limit: 1, Offset: 1}}
	got, err := repo.SearchAppointments(ctx, f)
	if err != nil || len(got) != 1 || got[0].ID != int(ids[1]) {
		t.Fatalf("page: %+v %v", got, err)
	}
	f.Page = models.Page{}
	got, err = repo.SearchAppointments(ctx, f)
	if err != nil || len(got) != 2 {
		t.Fatalf("half-open: %+v %v", got, err)
	}
}
