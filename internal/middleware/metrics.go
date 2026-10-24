package middleware

import (
	"expvar"
	"net/http"
	"strconv"
)

var (
	requestsTotal   = expvar.NewInt("http_requests_total")
	requestsByClass = expvar.NewMap("http_requests_by_status_class")
	responseBytes   = expvar.NewInt("http_response_bytes_total")
)

func Metrics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec := &statusRecorder{ResponseWriter: w}

		next.ServeHTTP(rec, r)

		status := rec.status
		if status == 0 {
			status = http.StatusOK
		}

		requestsTotal.Add(1)
		requestsByClass.Add(strconv.Itoa(status/100)+"xx", 1)
		responseBytes.Add(int64(rec.bytes))
	})
}
