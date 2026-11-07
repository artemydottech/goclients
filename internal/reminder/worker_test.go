package reminder

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/artemydottech/goclients/internal/models"
)

type fakeStore struct {
	due      []models.Reminder
	dueErr   error
	markErr  error
	marked   []int
	from, to time.Time
	markedAt time.Time
}

func (s *fakeStore) GetDueReminders(_ context.Context, from, to time.Time) ([]models.Reminder, error) {
	s.from, s.to = from, to
	return s.due, s.dueErr
}

func (s *fakeStore) MarkRemindedIfUnchanged(_ context.Context, id int, _ time.Time, at time.Time) error {
	if s.markErr != nil {
		return s.markErr
	}
	s.marked = append(s.marked, id)
	s.markedAt = at
	return nil
}

type fakeNotifier struct {
	sent   []int
	failOn map[int]bool
}

func (n *fakeNotifier) Notify(_ context.Context, r models.Reminder) error {
	if n.failOn[r.Appointment.ID] {
		return errors.New("sms gateway is down")
	}
	n.sent = append(n.sent, r.Appointment.ID)
	return nil
}

func reminderFor(id int) models.Reminder {
	return models.Reminder{Appointment: models.Appointment{ID: id}}
}

var fixedNow = time.Date(2026, 10, 24, 9, 0, 0, 0, time.UTC)

func newTestWorker(store Store, notifier Notifier) *Worker {
	w := NewWorker(store, notifier, 24*time.Hour)
	w.now = func() time.Time { return fixedNow }

	return w
}

func TestRunOnceNotifiesAndMarksEveryDueAppointment(t *testing.T) {
	store := &fakeStore{due: []models.Reminder{reminderFor(1), reminderFor(2)}}
	notifier := &fakeNotifier{}

	sent, err := newTestWorker(store, notifier).RunOnce(context.Background())
	if err != nil {
		t.Fatalf("RunOnce: %v", err)
	}

	if sent != 2 || len(notifier.sent) != 2 || len(store.marked) != 2 {
		t.Errorf("sent=%d notified=%v marked=%v", sent, notifier.sent, store.marked)
	}
	if !store.markedAt.Equal(fixedNow) {
		t.Errorf("markedAt = %v, want %v", store.markedAt, fixedNow)
	}
}

func TestRunOnceAsksForTheLeadWindow(t *testing.T) {
	store := &fakeStore{}

	if _, err := newTestWorker(store, &fakeNotifier{}).RunOnce(context.Background()); err != nil {
		t.Fatalf("RunOnce: %v", err)
	}

	if !store.from.Equal(fixedNow) || !store.to.Equal(fixedNow.Add(24*time.Hour)) {
		t.Errorf("window = %v - %v", store.from, store.to)
	}
}

func TestFailedNotificationIsNotMarkedAndDoesNotStopTheRest(t *testing.T) {
	store := &fakeStore{due: []models.Reminder{reminderFor(1), reminderFor(2), reminderFor(3)}}
	notifier := &fakeNotifier{failOn: map[int]bool{2: true}}

	sent, err := newTestWorker(store, notifier).RunOnce(context.Background())
	if err != nil {
		t.Fatalf("RunOnce: %v", err)
	}

	if sent != 2 {
		t.Errorf("sent = %d, want 2", sent)
	}
	if len(store.marked) != 2 || store.marked[0] != 1 || store.marked[1] != 3 {
		t.Errorf("marked = %v, the failed appointment must stay unmarked so the next run retries it", store.marked)
	}
}

func TestRunOnceReportsStoreErrors(t *testing.T) {
	store := &fakeStore{dueErr: errors.New("db is locked")}

	if _, err := newTestWorker(store, &fakeNotifier{}).RunOnce(context.Background()); err == nil {
		t.Fatal("expected the store error")
	}
}

func TestMarkFailureDoesNotCountAsSent(t *testing.T) {
	store := &fakeStore{due: []models.Reminder{reminderFor(1)}, markErr: errors.New("db is locked")}

	sent, err := newTestWorker(store, &fakeNotifier{}).RunOnce(context.Background())
	if err != nil {
		t.Fatalf("RunOnce: %v", err)
	}

	if sent != 0 {
		t.Errorf("sent = %d, want 0", sent)
	}
}

func TestRunOnceStopsWhenTheContextIsCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	store := &fakeStore{due: []models.Reminder{reminderFor(1)}}
	notifier := &fakeNotifier{}

	_, err := newTestWorker(store, notifier).RunOnce(ctx)

	if !errors.Is(err, context.Canceled) {
		t.Fatalf("got %v, want context.Canceled", err)
	}
	if len(notifier.sent) != 0 {
		t.Errorf("notified %v after cancellation", notifier.sent)
	}
}
