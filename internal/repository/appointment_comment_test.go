package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/artemydottech/goclients/internal/models"
)

func TestAppointmentCommentKeepsBookingDetails(t *testing.T) {
	f := newReminderFixture(t)
	ctx := context.Background()
	id := f.book(t, time.Now().UTC().Add(time.Hour), models.AppointmentConfirmed)
	before, err := f.repo.GetAppointmentById(ctx, id)
	if err != nil {
		t.Fatal(err)
	}
	for _, comment := range []string{"Уточнить пожелания", ""} {
		if err := f.repo.UpdateComment(ctx, id, comment); err != nil {
			t.Fatal(err)
		}
		got, err := f.repo.GetAppointmentById(ctx, id)
		if err != nil {
			t.Fatal(err)
		}
		want := before
		want.Comment = comment
		if got != want {
			t.Fatalf("booking changed: %+v != %+v", got, want)
		}
	}
	if err := f.repo.UpdateComment(ctx, 99999, "missing"); !errors.Is(err, models.ErrNotFound) {
		t.Fatalf("missing: %v", err)
	}
}
