package reminder

import (
	"context"
	"log/slog"
	"time"

	"github.com/artemydottech/goclients/internal/models"
)

type Store interface {
	GetDueReminders(ctx context.Context, from, to time.Time) ([]models.Reminder, error)
	MarkReminded(ctx context.Context, id int, at time.Time) error
}

type Worker struct {
	store    Store
	notifier Notifier
	lead     time.Duration
	now      func() time.Time
}

func NewWorker(store Store, notifier Notifier, lead time.Duration) *Worker {
	return &Worker{store: store, notifier: notifier, lead: lead, now: time.Now}
}

func (w *Worker) RunOnce(ctx context.Context) (int, error) {
	now := w.now()

	due, err := w.store.GetDueReminders(ctx, now, now.Add(w.lead))
	if err != nil {
		return 0, err
	}

	sent := 0
	for _, reminder := range due {
		if err := ctx.Err(); err != nil {
			return sent, err
		}

		if err := w.notifier.Notify(ctx, reminder); err != nil {
			slog.ErrorContext(ctx, "reminder failed", "appointment_id", reminder.Appointment.ID, "err", err)
			continue
		}

		if err := w.store.MarkReminded(ctx, reminder.Appointment.ID, now); err != nil {
			slog.ErrorContext(ctx, "marking reminder failed", "appointment_id", reminder.Appointment.ID, "err", err)
			continue
		}

		sent++
	}

	return sent, nil
}
