package app

import (
	"net/http"
	"testing"
)

func TestAvailabilityHTTPRejectsUnboundedRequests(t *testing.T) {
	s := newSalon(t)
	path := "/services/" + itoa(s.serviceID) + "/availability"
	for _, q := range []string{"", "?date=2026-02-30", "?date=2027-01-01&days=15", "?date=2027-01-01&step=0", "?date=2027-01-01&limit=201", "?date=2027-01-01&employee_id=-1"} {
		s.api.expect(http.MethodGet, path+q, nil, 400)
	}
	s.api.expect(http.MethodGet, "/services/999999/availability?date=2027-01-01", nil, 404)
}
