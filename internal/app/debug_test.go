package app

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/artemydottech/goclients/internal/repository"
)

func debugClient(t *testing.T, opts Options) client {
	t.Helper()

	db, err := repository.Open(context.Background(), filepath.Join(t.TempDir(), "debug.db"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	server := httptest.NewServer(New(db, opts))
	t.Cleanup(server.Close)

	return client{t: t, base: server.URL}
}

func TestDebugEndpointsAreHiddenByDefault(t *testing.T) {
	api := debugClient(t, Options{})

	api.expect(http.MethodGet, "/debug/vars", nil, http.StatusNotFound)
	api.expect(http.MethodGet, "/debug/pprof/", nil, http.StatusNotFound)
}

func TestDebugEndpointsCanBeEnabled(t *testing.T) {
	api := debugClient(t, Options{DebugEndpoints: true})

	api.expect(http.MethodGet, "/debug/vars", nil, http.StatusOK)
	api.expect(http.MethodGet, "/debug/pprof/", nil, http.StatusOK)
	api.expect(http.MethodGet, "/debug/pprof/cmdline", nil, http.StatusOK)
}
