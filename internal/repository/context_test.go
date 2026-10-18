package repository

import (
	"context"
	"errors"
	"testing"
)

func cancelledContext() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	return ctx
}

func assertCancelled(t *testing.T, err error) {
	t.Helper()

	if !errors.Is(err, context.Canceled) {
		t.Fatalf("got %v, want context.Canceled", err)
	}
}
