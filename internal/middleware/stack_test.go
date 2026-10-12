package middleware

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newStack(logger *slog.Logger, inner http.Handler) http.Handler {
	return Chain(inner,
		RequestID,
		AccessLog(logger),
		Recover,
		SecureHeaders,
		BodyLimit(16),
	)
}

func TestStackLogsPanickingRequestWithItsID(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(NewContextHandler(slog.NewTextHandler(&buf, nil)))
	slog.SetDefault(logger)
	t.Cleanup(func() { slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, nil))) })

	h := newStack(logger, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("kaboom")
	}))
	req := httptest.NewRequest(http.MethodGet, "/boom", nil)
	req.Header.Set("X-Request-ID", "trace-1")
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", rec.Code)
	}
	if rec.Header().Get("X-Request-ID") != "trace-1" {
		t.Errorf("response lost the request id")
	}
	if !strings.Contains(buf.String(), "panic in handler") || !strings.Contains(buf.String(), "request_id=trace-1") {
		t.Errorf("panic was not logged with the request id: %s", buf.String())
	}
	if !strings.Contains(buf.String(), "status=500") {
		t.Errorf("access log has no 500 for the panicking request: %s", buf.String())
	}
}

func TestStackLimitsBodyAndKeepsHeaders(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(NewContextHandler(slog.NewTextHandler(&buf, nil)))
	h := newStack(logger, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := io.ReadAll(r.Body); err != nil {
			http.Error(w, "too big", http.StatusRequestEntityTooLarge)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/", strings.NewReader(strings.Repeat("x", 64))))

	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d, want 413", rec.Code)
	}
	if rec.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Errorf("security headers are missing")
	}
	if !strings.Contains(buf.String(), "status=413") {
		t.Errorf("access log has no 413: %s", buf.String())
	}
}
