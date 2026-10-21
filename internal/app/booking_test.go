package app

import (
	"net/http"
	"testing"
)

func TestSlotsSkipTheLunchBreak(t *testing.T) {
	s := newSalon(t)

	slots := s.slots()

	if !hasSlot(slots, s.at(9, 0)) || !hasSlot(slots, s.at(12, 0)) || !hasSlot(slots, s.at(17, 0)) {
		t.Errorf("expected 09:00, 12:00 and 17:00 to be free, got %v", slots)
	}
	if hasSlot(slots, s.at(13, 0)) {
		t.Errorf("13:00 falls into the break, got %v", slots)
	}
	if hasSlot(slots, s.at(18, 0)) {
		t.Errorf("18:00 is after closing, got %v", slots)
	}
}

func TestBookingTakesTheSlot(t *testing.T) {
	s := newSalon(t)

	s.mustBook(s.at(10, 0))

	slots := s.slots()
	if hasSlot(slots, s.at(10, 0)) {
		t.Errorf("10:00 must be taken, got %v", slots)
	}
	if !hasSlot(slots, s.at(9, 0)) || !hasSlot(slots, s.at(11, 0)) {
		t.Errorf("neighbouring slots must stay free, got %v", slots)
	}
}

func TestDoubleBookingIsRejected(t *testing.T) {
	s := newSalon(t)
	s.mustBook(s.at(10, 0))

	status, _ := s.book(s.at(10, 0))
	if status != http.StatusBadRequest {
		t.Errorf("same slot = %d, want 400", status)
	}

	status, _ = s.book(s.at(10, 30))
	if status != http.StatusBadRequest {
		t.Errorf("overlapping slot = %d, want 400", status)
	}

	status, _ = s.book(s.at(11, 0))
	if status != http.StatusCreated {
		t.Errorf("touching slot = %d, want 201", status)
	}
}

func TestBookingDuringTheBreakIsRejected(t *testing.T) {
	s := newSalon(t)

	status, _ := s.book(s.at(13, 0))

	if status != http.StatusBadRequest {
		t.Errorf("booking during the break = %d, want 400", status)
	}
}

func TestBookingAfterClosingIsRejected(t *testing.T) {
	s := newSalon(t)

	status, _ := s.book(s.at(17, 30))

	if status != http.StatusBadRequest {
		t.Errorf("booking past closing time = %d, want 400", status)
	}
}
