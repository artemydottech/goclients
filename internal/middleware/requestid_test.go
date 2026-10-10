package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequestIDGeneratedWhenMissing(t *testing.T) {
	var seen string
	h := RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = RequestIDFrom(r.Context())
	}))
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if seen == "" {
		t.Fatal("request id was not put into the context")
	}
	if got := rec.Header().Get("X-Request-ID"); got != seen {
		t.Errorf("header = %q, context = %q", got, seen)
	}
}

func TestRequestIDKeepsIncomingValue(t *testing.T) {
	h := RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Request-ID", "abc-123")
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if got := rec.Header().Get("X-Request-ID"); got != "abc-123" {
		t.Errorf("header = %q, want abc-123", got)
	}
}

func TestRequestIDReplacesOversizedValue(t *testing.T) {
	h := RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	long := make([]byte, 200)
	for i := range long {
		long[i] = 'a'
	}
	req.Header.Set("X-Request-ID", string(long))
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if got := rec.Header().Get("X-Request-ID"); len(got) > 64 {
		t.Errorf("oversized id was kept: %d chars", len(got))
	}
}
