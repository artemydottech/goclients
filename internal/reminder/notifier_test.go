package reminder

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/artemydottech/goclients/internal/models"
)

func TestLogNotifierWritesTheReminder(t *testing.T) {
	var buf bytes.Buffer
	notifier := LogNotifier{Logger: slog.New(slog.NewTextHandler(&buf, nil))}

	err := notifier.Notify(context.Background(), models.Reminder{
		Appointment: models.Appointment{ID: 12, StartsAt: time.Date(2026, 10, 25, 10, 0, 0, 0, time.UTC)},
		ClientName:  "Ivan",
		ClientPhone: "79990000000",
	})
	if err != nil {
		t.Fatalf("Notify: %v", err)
	}

	line := buf.String()
	for _, want := range []string{"appointment_id=12", "client=Ivan", "phone=79990000000"} {
		if !strings.Contains(line, want) {
			t.Errorf("log line %q does not contain %q", line, want)
		}
	}
}

func TestLogNotifierFallsBackToTheDefaultLogger(t *testing.T) {
	if err := (LogNotifier{}).Notify(context.Background(), models.Reminder{}); err != nil {
		t.Fatalf("Notify: %v", err)
	}
}
