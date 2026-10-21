package app

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"
)

func TestEmployeeMustPerformTheService(t *testing.T) {
	s := newSalon(t)
	other := s.api.create("/services", map[string]any{"company_id": s.companyID, "name": "Manicure", "duration": 45})

	status, _ := s.api.do(http.MethodPost, "/appointments", map[string]any{
		"client_id": s.clientID, "employee_id": s.employeeID, "service_id": other,
		"starts_at": s.at(10, 0).Format(time.RFC3339),
	})

	if status != http.StatusBadRequest {
		t.Errorf("unassigned service = %d, want 400", status)
	}
}

func TestEverythingMustBelongToOneCompany(t *testing.T) {
	s := newSalon(t)
	foreign := s.api.create("/companies", map[string]any{"name": "Lilia"})
	stranger := s.api.create("/clients", map[string]any{"company_id": foreign, "name": "Stranger", "phone": "79991112233"})

	status, _ := s.api.do(http.MethodPost, "/appointments", map[string]any{
		"client_id": stranger, "employee_id": s.employeeID, "service_id": s.serviceID,
		"starts_at": s.at(10, 0).Format(time.RFC3339),
	})

	if status != http.StatusBadRequest {
		t.Errorf("client from another company = %d, want 400", status)
	}
}

func TestPastAppointmentsAreRejected(t *testing.T) {
	s := newSalon(t)

	status, _ := s.book(time.Now().UTC().AddDate(0, 0, -1))

	if status != http.StatusBadRequest {
		t.Errorf("booking in the past = %d, want 400", status)
	}
}

func TestPriceIsFixedAtBookingTime(t *testing.T) {
	s := newSalon(t)
	id := s.mustBook(s.at(10, 0))

	s.api.expect(http.MethodPut, "/services/"+itoa(s.serviceID), map[string]any{
		"name": "Haircut", "duration": 60, "price": 2500,
	}, http.StatusNoContent)

	raw := s.api.expect(http.MethodGet, "/appointments/"+itoa(id), nil, http.StatusOK)
	var got struct {
		Price float64 `json:"price"`
	}
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.Price != 1000 {
		t.Errorf("price = %v, want the price at booking time (1000)", got.Price)
	}
}

func TestClientStatsCountUpcomingAndCancelled(t *testing.T) {
	s := newSalon(t)
	s.mustBook(s.at(9, 0))
	cancelled := s.mustBook(s.at(10, 0))
	s.api.expect(http.MethodPut, "/appointments/"+itoa(cancelled)+"/status", map[string]any{"status": "cancelled"}, http.StatusNoContent)

	raw := s.api.expect(http.MethodGet, "/clients/"+itoa(s.clientID)+"/stats", nil, http.StatusOK)

	var stats struct {
		Appointments int `json:"appointments"`
		Upcoming     int `json:"upcoming"`
		Cancelled    int `json:"cancelled"`
	}
	if err := json.Unmarshal(raw, &stats); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if stats.Appointments != 2 || stats.Upcoming != 1 || stats.Cancelled != 1 {
		t.Errorf("stats = %+v", stats)
	}
}

func TestDeletingTheClientRemovesTheirAppointments(t *testing.T) {
	s := newSalon(t)
	id := s.mustBook(s.at(10, 0))

	s.api.expect(http.MethodDelete, "/clients/"+itoa(s.clientID), nil, http.StatusNoContent)

	s.api.expect(http.MethodGet, "/appointments/"+itoa(id), nil, http.StatusNotFound)
	if !hasSlot(s.slots(), s.at(10, 0)) {
		t.Error("the slot must be free after the client is deleted")
	}
}
