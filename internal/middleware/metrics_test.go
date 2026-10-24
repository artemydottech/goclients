package middleware

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

func classCount(t *testing.T, class string) int64 {
	t.Helper()

	value := requestsByClass.Get(class)
	if value == nil {
		return 0
	}

	n, err := strconv.ParseInt(value.String(), 10, 64)
	if err != nil {
		t.Fatalf("parse %q: %v", value.String(), err)
	}

	return n
}

func TestMetricsCountRequestsByStatusClass(t *testing.T) {
	totalBefore := requestsTotal.Value()
	okBefore := classCount(t, "2xx")
	notFoundBefore := classCount(t, "4xx")
	bytesBefore := responseBytes.Value()

	h := Metrics(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/missing" {
			http.NotFound(w, r)
			return
		}
		w.Write([]byte("hello"))
	}))

	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/ok", nil))
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/ok", nil))
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/missing", nil))

	if got := requestsTotal.Value() - totalBefore; got != 3 {
		t.Errorf("total grew by %d, want 3", got)
	}
	if got := classCount(t, "2xx") - okBefore; got != 2 {
		t.Errorf("2xx grew by %d, want 2", got)
	}
	if got := classCount(t, "4xx") - notFoundBefore; got != 1 {
		t.Errorf("4xx grew by %d, want 1", got)
	}
	if got := responseBytes.Value() - bytesBefore; got < 10 {
		t.Errorf("response bytes grew by %d, want at least 10", got)
	}
}

func TestMetricsTreatAnImplicitOKAsSuccess(t *testing.T) {
	before := classCount(t, "2xx")
	h := Metrics(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))

	if got := classCount(t, "2xx") - before; got != 1 {
		t.Errorf("2xx grew by %d, want 1", got)
	}
}
