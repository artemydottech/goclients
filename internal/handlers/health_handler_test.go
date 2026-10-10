package handlers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

type stubPinger struct {
	err error
}

func (s stubPinger) PingContext(ctx context.Context) error {
	return s.err
}

func TestLiveAlwaysOK(t *testing.T) {
	h := NewHealthHandler(stubPinger{err: errors.New("db down")})
	rec := httptest.NewRecorder()

	h.Live(rec, httptest.NewRequest(http.MethodGet, "/healthz", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
}

func TestReadyReportsDatabaseState(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{"database answers", nil, http.StatusOK},
		{"database is down", errors.New("db down"), http.StatusServiceUnavailable},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewHealthHandler(stubPinger{err: tt.err})
			rec := httptest.NewRecorder()

			h.Ready(rec, httptest.NewRequest(http.MethodGet, "/readyz", nil))

			if rec.Code != tt.want {
				t.Fatalf("status = %d, want %d", rec.Code, tt.want)
			}
		})
	}
}
