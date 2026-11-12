package repository

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/artemydottech/goclients/internal/models"
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

func TestBookingReservationRollsBackAndSerializesRetries(t *testing.T) {
	f := newReminderFixture(t)
	ctx := context.Background()
	start := time.Now().UTC().Add(time.Hour)
	a := models.Appointment{CompanyID: f.companyID, ClientID: f.clientID, EmployeeID: f.employee, ServiceID: f.service, StartsAt: start, EndsAt: start.Add(time.Hour), Status: models.AppointmentPending}
	if _, err := f.repo.ReserveBooking(ctx, "occupied", "first", a); err != nil {
		t.Fatal(err)
	}
	if _, err := f.repo.ReserveBooking(ctx, "failed", "second", a); err == nil {
		t.Fatal("expected occupied slot")
	}
	if _, err := f.repo.FindBookingRequest(ctx, a.ClientID, "failed"); !errors.Is(err, models.ErrNotFound) {
		t.Fatalf("failed key persisted: %v", err)
	}
	a.StartsAt = a.EndsAt
	a.EndsAt = a.EndsAt.Add(time.Hour)
	var wg sync.WaitGroup
	ids := make(chan int64, 12)
	for range 12 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			result, err := f.repo.ReserveBooking(ctx, "parallel", "same", a)
			if err != nil {
				t.Error(err)
				return
			}
			ids <- result.ID
		}()
	}
	wg.Wait()
	close(ids)
	var first int64
	count := 0
	for id := range ids {
		if first == 0 {
			first = id
		}
		if id != first {
			t.Fatalf("different IDs: %d %d", first, id)
		}
		count++
	}
	if count != 12 {
		t.Fatal(count)
	}
}
