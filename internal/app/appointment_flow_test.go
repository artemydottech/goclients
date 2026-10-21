package app

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"
)

func TestCancelledAppointmentFreesTheSlot(t *testing.T) {
	s := newSalon(t)
	id := s.mustBook(s.at(10, 0))

	s.api.expect(http.MethodPut, "/appointments/"+itoa(id)+"/status", map[string]any{"status": "cancelled"}, http.StatusNoContent)

	if !hasSlot(s.slots(), s.at(10, 0)) {
		t.Error("10:00 must be free again after the cancellation")
	}
	s.mustBook(s.at(10, 0))
}

func TestCancelledAppointmentCannotBeRestored(t *testing.T) {
	s := newSalon(t)
	id := s.mustBook(s.at(10, 0))
	s.api.expect(http.MethodPut, "/appointments/"+itoa(id)+"/status", map[string]any{"status": "cancelled"}, http.StatusNoContent)

	s.api.expect(http.MethodPut, "/appointments/"+itoa(id)+"/status", map[string]any{"status": "confirmed"}, http.StatusBadRequest)
}

func TestFutureAppointmentCannotBeCompleted(t *testing.T) {
	s := newSalon(t)
	id := s.mustBook(s.at(10, 0))

	s.api.expect(http.MethodPut, "/appointments/"+itoa(id)+"/status", map[string]any{"status": "completed"}, http.StatusBadRequest)
	s.api.expect(http.MethodPut, "/appointments/"+itoa(id)+"/status", map[string]any{"status": "confirmed"}, http.StatusNoContent)
}

func TestRescheduleMovesTheAppointment(t *testing.T) {
	s := newSalon(t)
	id := s.mustBook(s.at(10, 0))

	raw := s.api.expect(http.MethodPut, "/appointments/"+itoa(id)+"/time",
		map[string]any{"starts_at": s.at(15, 0).Format(time.RFC3339)}, http.StatusOK)

	var moved struct {
		StartsAt time.Time `json:"starts_at"`
		EndsAt   time.Time `json:"ends_at"`
	}
	if err := json.Unmarshal(raw, &moved); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !moved.StartsAt.Equal(s.at(15, 0)) || !moved.EndsAt.Equal(s.at(16, 0)) {
		t.Errorf("moved to %v - %v", moved.StartsAt, moved.EndsAt)
	}

	slots := s.slots()
	if !hasSlot(slots, s.at(10, 0)) || hasSlot(slots, s.at(15, 0)) {
		t.Errorf("slots did not follow the move: %v", slots)
	}
}

func TestRescheduleOntoABusySlotIsRejected(t *testing.T) {
	s := newSalon(t)
	id := s.mustBook(s.at(10, 0))
	s.mustBook(s.at(15, 0))

	s.api.expect(http.MethodPut, "/appointments/"+itoa(id)+"/time",
		map[string]any{"starts_at": s.at(15, 0).Format(time.RFC3339)}, http.StatusBadRequest)

	if hasSlot(s.slots(), s.at(10, 0)) {
		t.Error("a failed move must leave the original slot taken")
	}
}
