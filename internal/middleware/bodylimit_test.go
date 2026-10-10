package middleware

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestBodyLimitRejectsLargeBodies(t *testing.T) {
	var readErr error
	h := BodyLimit(10)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, readErr = io.ReadAll(r.Body)
	}))
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(strings.Repeat("x", 100)))

	h.ServeHTTP(httptest.NewRecorder(), req)

	if readErr == nil {
		t.Fatal("expected an error while reading an oversized body")
	}
}

func TestBodyLimitAllowsSmallBodies(t *testing.T) {
	var got string
	h := BodyLimit(10)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		got = string(b)
	}))
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("hello"))

	h.ServeHTTP(httptest.NewRecorder(), req)

	if got != "hello" {
		t.Errorf("body = %q, want hello", got)
	}
}
