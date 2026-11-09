package repository

import (
	"context"
	"errors"
	"github.com/artemydottech/goclients/internal/models"
	"testing"
	"time"
)

func TestBookingRequestReplayAndConflict(t *testing.T) {
	f := newReminderFixture(t)
	ctx := context.Background()
	start := time.Now().UTC().Add(time.Hour)
	a := models.Appointment{CompanyID: f.companyID, ClientID: f.clientID, EmployeeID: f.employee, ServiceID: f.service, StartsAt: start, EndsAt: start.Add(time.Hour), Status: models.AppointmentPending}
	first, err := f.repo.ReserveBooking(ctx, "request", "hash", a)
	if err != nil {
		t.Fatal(err)
	}
	retry, err := f.repo.ReserveBooking(ctx, "request", "hash", a)
	if err != nil || retry.ID != first.ID || !retry.Replayed {
		t.Fatalf("replay: %+v %v", retry, err)
	}
	if _, err := f.repo.ReserveBooking(ctx, "request", "different", a); !errors.Is(err, models.ErrConflict) {
		t.Fatalf("conflict: %v", err)
	}
	all, err := f.repo.GetAllAppointments(ctx)
	if err != nil || len(all) != 1 {
		t.Fatalf("duplicates: %+v %v", all, err)
	}
}
