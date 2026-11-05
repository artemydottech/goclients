package app

import (
	"encoding/json"
	"github.com/artemydottech/goclients/internal/models"
	"net/http"
	"net/url"
	"testing"
	"time"
)

func TestAppointmentSearchHTTPValidation(t *testing.T) {
	api := newClient(t)
	for _, q := range []string{"", "company_id=-1", "company_id=1&employee_id=0", "company_id=1&status=invalid", "company_id=1&from=yesterday", "company_id=1&limit=201", "company_id=1&offset=-1", "company_id=1&from=2026-11-06T00:00:00Z&to=2026-11-05T00:00:00Z"} {
		api.expect(http.MethodGet, "/appointments/search?"+q, nil, http.StatusBadRequest)
	}
}
func TestAppointmentSearchHTTPScopesCalendar(t *testing.T) {
	s := newSalon(t)
	id := s.mustBook(s.at(10, 0))
	q := url.Values{"company_id": {itoa(s.companyID)}, "employee_id": {itoa(s.employeeID)}, "client_id": {itoa(s.clientID)}, "from": {s.at(10, 0).Format(time.RFC3339)}, "to": {s.at(11, 0).Format(time.RFC3339)}, "status": {"pending"}}
	raw := s.api.expect(http.MethodGet, "/appointments/search?"+q.Encode(), nil, http.StatusOK)
	var got []models.Appointment
	if err := json.Unmarshal(raw, &got); err != nil || len(got) != 1 || got[0].ID != id {
		t.Fatalf("search: %s %v", raw, err)
	}
	q.Set("company_id", itoa(newCompany(t, s.api)))
	raw = s.api.expect(http.MethodGet, "/appointments/search?"+q.Encode(), nil, http.StatusOK)
	if string(raw) != "[]\n" {
		t.Fatalf("foreign calendar: %s", raw)
	}
}
