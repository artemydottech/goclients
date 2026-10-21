package app

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"
)

type salon struct {
	api        client
	companyID  int
	employeeID int
	serviceID  int
	clientID   int
	day        time.Time
}

func (s salon) at(hour, minute int) time.Time {
	return time.Date(s.day.Year(), s.day.Month(), s.day.Day(), hour, minute, 0, 0, time.UTC)
}

func (s salon) book(start time.Time) (int, []byte) {
	return s.api.do(http.MethodPost, "/appointments", map[string]any{
		"client_id":   s.clientID,
		"employee_id": s.employeeID,
		"service_id":  s.serviceID,
		"starts_at":   start.Format(time.RFC3339),
	})
}

func (s salon) mustBook(start time.Time) int {
	s.api.t.Helper()

	status, raw := s.book(start)
	if status != http.StatusCreated {
		s.api.t.Fatalf("booking %s = %d: %s", start.Format(time.RFC3339), status, raw)
	}

	var created struct {
		ID int `json:"id"`
	}
	if err := json.Unmarshal(raw, &created); err != nil {
		s.api.t.Fatalf("decode: %v", err)
	}

	return created.ID
}

func (s salon) slots() []time.Time {
	s.api.t.Helper()

	path := "/slots?employee_id=" + itoa(s.employeeID) + "&service_id=" + itoa(s.serviceID) +
		"&date=" + s.day.Format("2006-01-02") + "&step=60"
	raw := s.api.expect(http.MethodGet, path, nil, http.StatusOK)

	var slots []time.Time
	if err := json.Unmarshal(raw, &slots); err != nil {
		s.api.t.Fatalf("decode slots: %v", err)
	}

	return slots
}

func hasSlot(slots []time.Time, want time.Time) bool {
	for _, slot := range slots {
		if slot.Equal(want) {
			return true
		}
	}

	return false
}

func newSalon(t *testing.T) salon {
	t.Helper()

	api := newClient(t)
	companyID := newCompany(t, api)

	employeeID := api.create("/employees", map[string]any{"company_id": companyID, "name": "Anna", "surname": "Ivanova"})
	serviceID := api.create("/services", map[string]any{
		"company_id": companyID, "name": "Haircut", "duration": 60, "price": 1000,
	})
	clientID := api.create("/clients", map[string]any{"company_id": companyID, "name": "Ivan", "phone": "79990000000"})

	api.expect(http.MethodPut, "/employees/"+itoa(employeeID)+"/services", map[string]any{"service_ids": []int{serviceID}}, http.StatusNoContent)

	days := make([]map[string]any, 0, 7)
	for weekday := 0; weekday < 7; weekday++ {
		days = append(days, map[string]any{
			"weekday": weekday, "starts_at": "09:00", "ends_at": "18:00",
			"break_starts_at": "13:00", "break_ends_at": "14:00",
		})
	}
	api.expect(http.MethodPut, "/employees/"+itoa(employeeID)+"/schedule", map[string]any{"days": days}, http.StatusNoContent)

	return salon{
		api:        api,
		companyID:  companyID,
		employeeID: employeeID,
		serviceID:  serviceID,
		clientID:   clientID,
		day:        time.Now().UTC().AddDate(0, 0, 14),
	}
}
