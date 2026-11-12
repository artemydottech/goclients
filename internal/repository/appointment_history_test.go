package repository

import (
	"context"
	"github.com/artemydottech/goclients/internal/models"
	"testing"
	"time"
)

func TestAppointmentHistoryCapturesLifecycleAndSkipsNoops(t *testing.T) {
	f := newReminderFixture(t)
	ctx := context.Background()
	id := f.book(t, time.Now().UTC().Add(time.Hour), models.AppointmentPending)
	if err := f.repo.UpdateStatus(ctx, id, models.AppointmentConfirmed); err != nil {
		t.Fatal(err)
	}
	if err := f.repo.UpdateStatus(ctx, id, models.AppointmentConfirmed); err != nil {
		t.Fatal(err)
	}
	if err := f.repo.UpdateComment(ctx, id, "Updated note"); err != nil {
		t.Fatal(err)
	}
	a, err := f.repo.GetAppointmentById(ctx, id)
	if err != nil {
		t.Fatal(err)
	}
	a.StartsAt = a.StartsAt.Add(time.Hour)
	a.EndsAt = a.EndsAt.Add(time.Hour)
	if busy, err := f.repo.MoveIfFree(ctx, id, a); err != nil || busy {
		t.Fatalf("move: %v %v", busy, err)
	}
	events, err := f.repo.GetAppointmentHistory(ctx, id, models.Page{})
	if err != nil || len(events) != 4 {
		t.Fatalf("events: %+v %v", events, err)
	}
	for i, kind := range []string{"rescheduled", "comment", "status", "created"} {
		if events[i].Kind != kind || events[i].OccurredAt.IsZero() || events[i].After.ID != id {
			t.Fatalf("event: %+v", events[i])
		}
	}
	if events[0].Before == nil || !events[0].After.StartsAt.After(events[0].Before.StartsAt) || events[3].Before != nil {
		t.Fatalf("snapshots: %+v", events)
	}
	page, err := f.repo.GetAppointmentHistory(ctx, id, models.Page{Limit: 1, Offset: 2})
	if err != nil || len(page) != 1 || page[0].Kind != "status" {
		t.Fatalf("page: %+v %v", page, err)
	}
}
