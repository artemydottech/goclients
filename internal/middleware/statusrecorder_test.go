package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestStatusRecorderDefaultsToOK(t *testing.T) {
	rec := &statusRecorder{ResponseWriter: httptest.NewRecorder()}

	rec.Write([]byte("hello"))

	if rec.status != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.status)
	}
	if rec.bytes != 5 {
		t.Errorf("bytes = %d, want 5", rec.bytes)
	}
}

func TestStatusRecorderKeepsFirstStatus(t *testing.T) {
	rec := &statusRecorder{ResponseWriter: httptest.NewRecorder()}

	rec.WriteHeader(http.StatusNotFound)
	rec.WriteHeader(http.StatusOK)

	if rec.status != http.StatusNotFound {
		t.Errorf("status = %d, want 404", rec.status)
	}
}
