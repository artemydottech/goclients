package app

import (
	"encoding/json"
	"github.com/artemydottech/goclients/internal/models"
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
func TestAvailabilityFindsAnotherEmployeeWhenFirstIsBooked(t *testing.T) {
	s := newSalon(t)
	s.mustBook(s.at(10, 0))
	second := s.api.create("/employees", map[string]any{"company_id": s.companyID, "name": "Second", "surname": "Master"})
	s.api.expect(http.MethodPut, "/employees/"+itoa(second)+"/services", map[string]any{"service_ids": []int{s.serviceID}}, 204)
	s.api.expect(http.MethodPut, "/employees/"+itoa(second)+"/schedule", map[string]any{"days": []map[string]any{{"weekday": int(s.day.Weekday()), "starts_at": "10:00", "ends_at": "11:00"}}}, 204)
	raw := s.api.expect(http.MethodGet, "/services/"+itoa(s.serviceID)+"/availability?date="+s.day.Format("2006-01-02")+"&days=1&step=60&limit=200", nil, 200)
	var got []models.OpenSlot
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatal(err)
	}
	found := false
	for _, slot := range got {
		if slot.StartsAt.Equal(s.at(10, 0)) {
			if slot.EmployeeID != second {
				t.Fatalf("busy master: %+v", slot)
			}
			found = true
		}
	}
	if !found {
		t.Fatalf("second employee not offered: %s", raw)
	}
}
