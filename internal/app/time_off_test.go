package app

import (
	"net/http"
	"testing"
	"time"
)

func (s salon) timeOff(from, to time.Time) (int, []byte) {
	return s.api.do(http.MethodPost, "/employees/"+itoa(s.employeeID)+"/time-off", map[string]any{
		"starts_at": from.Format(time.RFC3339),
		"ends_at":   to.Format(time.RFC3339),
		"reason":    "vacation",
	})
}

func TestTimeOffRemovesTheSlotsAndBlocksBooking(t *testing.T) {
	s := newSalon(t)

	status, raw := s.timeOff(s.at(0, 0), s.at(0, 0).AddDate(0, 0, 1))
	if status != http.StatusCreated {
		t.Fatalf("time off = %d: %s", status, raw)
	}

	if slots := s.slots(); len(slots) != 0 {
		t.Errorf("expected no slots during time off, got %v", slots)
	}
	if status, _ := s.book(s.at(10, 0)); status != http.StatusBadRequest {
		t.Errorf("booking during time off = %d, want 400", status)
	}
}

func TestDeletingTimeOffReturnsTheSlots(t *testing.T) {
	s := newSalon(t)
	status, raw := s.timeOff(s.at(0, 0), s.at(0, 0).AddDate(0, 0, 1))
	if status != http.StatusCreated {
		t.Fatalf("time off = %d: %s", status, raw)
	}
	id := decodeID(t, raw)

	s.api.expect(http.MethodDelete, "/time-off/"+itoa(id), nil, http.StatusNoContent)

	if !hasSlot(s.slots(), s.at(10, 0)) {
		t.Error("10:00 must be free after the time off is removed")
	}
}

func TestTimeOffOverActiveAppointmentsIsRejected(t *testing.T) {
	s := newSalon(t)
	s.mustBook(s.at(10, 0))

	status, _ := s.timeOff(s.at(0, 0), s.at(0, 0).AddDate(0, 0, 1))

	if status != http.StatusBadRequest {
		t.Errorf("time off over an appointment = %d, want 400", status)
	}
}

func TestTimeOffMustEndAfterItStarts(t *testing.T) {
	s := newSalon(t)

	status, _ := s.timeOff(s.at(12, 0), s.at(10, 0))

	if status != http.StatusBadRequest {
		t.Errorf("reversed period = %d, want 400", status)
	}
}
