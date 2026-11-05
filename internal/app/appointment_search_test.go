package app

import (
	"net/http"
	"testing"
)

func TestAppointmentSearchHTTPValidation(t *testing.T) {
	api := newClient(t)
	for _, q := range []string{"", "company_id=-1", "company_id=1&employee_id=0", "company_id=1&status=invalid", "company_id=1&from=yesterday", "company_id=1&limit=201", "company_id=1&offset=-1", "company_id=1&from=2026-11-06T00:00:00Z&to=2026-11-05T00:00:00Z"} {
		api.expect(http.MethodGet, "/appointments/search?"+q, nil, http.StatusBadRequest)
	}
}
