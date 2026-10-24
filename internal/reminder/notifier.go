package reminder

import (
	"context"
	"log/slog"

	"github.com/artemydottech/goclients/internal/models"
)

type Notifier interface {
	Notify(ctx context.Context, r models.Reminder) error
}

type LogNotifier struct {
	Logger *slog.Logger
}

func (n LogNotifier) Notify(ctx context.Context, r models.Reminder) error {
	logger := n.Logger
	if logger == nil {
		logger = slog.Default()
	}

	logger.InfoContext(ctx, "appointment reminder",
		"appointment_id", r.Appointment.ID,
		"client", r.ClientName,
		"phone", r.ClientPhone,
		"starts_at", r.Appointment.StartsAt,
	)

	return nil
}
