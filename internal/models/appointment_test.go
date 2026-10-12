package models

import "testing"

func TestAppointmentStatusValid(t *testing.T) {
	valid := []AppointmentStatus{AppointmentPending, AppointmentConfirmed, AppointmentCancelled, AppointmentCompleted, AppointmentNoShow}
	for _, s := range valid {
		if !s.Valid() {
			t.Errorf("%q must be valid", s)
		}
	}

	for _, s := range []AppointmentStatus{"", "done", "PENDING"} {
		if s.Valid() {
			t.Errorf("%q must not be valid", s)
		}
	}
}

func TestAppointmentStatusCanBecome(t *testing.T) {
	tests := []struct {
		from AppointmentStatus
		to   AppointmentStatus
		want bool
	}{
		{AppointmentPending, AppointmentConfirmed, true},
		{AppointmentPending, AppointmentCancelled, true},
		{AppointmentPending, AppointmentCompleted, true},
		{AppointmentPending, AppointmentNoShow, true},
		{AppointmentConfirmed, AppointmentCancelled, true},
		{AppointmentConfirmed, AppointmentCompleted, true},
		{AppointmentConfirmed, AppointmentNoShow, true},
		{AppointmentConfirmed, AppointmentPending, false},
		{AppointmentCancelled, AppointmentConfirmed, false},
		{AppointmentCancelled, AppointmentPending, false},
		{AppointmentCompleted, AppointmentCancelled, false},
		{AppointmentNoShow, AppointmentCompleted, false},
		{AppointmentCancelled, AppointmentCancelled, true},
		{AppointmentCompleted, AppointmentCompleted, true},
	}

	for _, tt := range tests {
		t.Run(string(tt.from)+" to "+string(tt.to), func(t *testing.T) {
			if got := tt.from.CanBecome(tt.to); got != tt.want {
				t.Errorf("CanBecome = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAppointmentStatusGroups(t *testing.T) {
	tests := []struct {
		status   AppointmentStatus
		blocks   bool
		active   bool
		happened bool
	}{
		{AppointmentPending, true, true, false},
		{AppointmentConfirmed, true, true, false},
		{AppointmentCancelled, false, false, false},
		{AppointmentCompleted, true, false, true},
		{AppointmentNoShow, true, false, true},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			if got := tt.status.Blocks(); got != tt.blocks {
				t.Errorf("Blocks = %v, want %v", got, tt.blocks)
			}
			if got := tt.status.Active(); got != tt.active {
				t.Errorf("Active = %v, want %v", got, tt.active)
			}
			if got := tt.status.Happened(); got != tt.happened {
				t.Errorf("Happened = %v, want %v", got, tt.happened)
			}
		})
	}
}
