package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func serve(t *testing.T, pattern string, handler http.HandlerFunc, method, target, body string) *httptest.ResponseRecorder {
	t.Helper()

	return serveWithContext(t, context.Background(), pattern, handler, method, target, body)
}

type ctxKey struct{}

func markedContext() context.Context {
	return context.WithValue(context.Background(), ctxKey{}, "request-scope")
}

func assertMarked(t *testing.T, got context.Context) {
	t.Helper()

	if got == nil {
		t.Fatal("service was called without a context")
	}
	if got.Value(ctxKey{}) != "request-scope" {
		t.Fatal("service did not receive the request context")
	}
}

func serveWithContext(t *testing.T, ctx context.Context, pattern string, handler http.HandlerFunc, method, target, body string) *httptest.ResponseRecorder {
	t.Helper()

	mux := http.NewServeMux()
	mux.HandleFunc(pattern, handler)

	req := httptest.NewRequest(method, target, strings.NewReader(body)).WithContext(ctx)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	return rec
}

func assertStatus(t *testing.T, rec *httptest.ResponseRecorder, want int) {
	t.Helper()

	if rec.Code != want {
		t.Fatalf("status = %d, want %d, body: %s", rec.Code, want, rec.Body.String())
	}
}
