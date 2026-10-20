package app

import (
	"net/http"
	"testing"
)

func TestHealthEndpoints(t *testing.T) {
	api := newClient(t)

	api.expect(http.MethodGet, "/healthz", nil, http.StatusOK)
	api.expect(http.MethodGet, "/readyz", nil, http.StatusOK)
}

func TestUnknownRouteIsNotFound(t *testing.T) {
	api := newClient(t)

	api.expect(http.MethodGet, "/nope", nil, http.StatusNotFound)
}

func TestWrongMethodIsRejected(t *testing.T) {
	api := newClient(t)

	api.expect(http.MethodPatch, "/companies/1", nil, http.StatusMethodNotAllowed)
}
