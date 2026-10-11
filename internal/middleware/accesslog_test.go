package middleware

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAccessLogWritesRequestLine(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(NewContextHandler(slog.NewTextHandler(&buf, nil)))
	h := RequestID(AccessLog(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	})))
	req := httptest.NewRequest(http.MethodGet, "/tea", nil)
	req.Header.Set("X-Request-ID", "req-1")

	h.ServeHTTP(httptest.NewRecorder(), req)

	line := buf.String()
	for _, want := range []string{"method=GET", "path=/tea", "status=418", "request_id=req-1"} {
		if !strings.Contains(line, want) {
			t.Errorf("log line %q does not contain %q", line, want)
		}
	}
}
