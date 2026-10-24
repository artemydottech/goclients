package reminder

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/artemydottech/goclients/internal/models"
)

type countingStore struct {
	calls atomic.Int32
}

func (s *countingStore) GetDueReminders(context.Context, time.Time, time.Time) ([]models.Reminder, error) {
	s.calls.Add(1)
	return nil, nil
}

func (s *countingStore) MarkReminded(context.Context, int, time.Time) error { return nil }

func TestRunPollsUntilTheContextIsCancelled(t *testing.T) {
	store := &countingStore{}
	worker := NewWorker(store, &fakeNotifier{}, time.Hour)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})

	go func() {
		worker.Run(ctx, 5*time.Millisecond)
		close(done)
	}()

	deadline := time.After(2 * time.Second)
	for store.calls.Load() < 3 {
		select {
		case <-deadline:
			t.Fatalf("worker polled only %d times", store.calls.Load())
		case <-time.After(5 * time.Millisecond):
		}
	}

	cancel()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("worker did not stop after the context was cancelled")
	}
}

func TestRunRunsOnceImmediately(t *testing.T) {
	store := &countingStore{}
	worker := NewWorker(store, &fakeNotifier{}, time.Hour)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})

	go func() {
		worker.Run(ctx, time.Hour)
		close(done)
	}()

	deadline := time.After(2 * time.Second)
	for store.calls.Load() < 1 {
		select {
		case <-deadline:
			t.Fatal("worker did not run right after the start")
		case <-time.After(2 * time.Millisecond):
		}
	}

	cancel()
	<-done
}
