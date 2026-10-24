package reminder

import (
	"context"
	"errors"
	"log/slog"
	"time"
)

func (w *Worker) Run(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		sent, err := w.RunOnce(ctx)
		switch {
		case errors.Is(err, context.Canceled):
			return
		case err != nil:
			slog.ErrorContext(ctx, "reminder run failed", "err", err)
		case sent > 0:
			slog.InfoContext(ctx, "reminders sent", "count", sent)
		}

		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}
