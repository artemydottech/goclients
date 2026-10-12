package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/artemydottech/goclients/internal/models"
)

func TestGetFreeSlotsValidatesQuery(t *testing.T) {
	tests := []struct {
		name   string
		target string
	}{
		{"no employee", "/slots?service_id=1&date=2026-10-20"},
		{"employee is not a number", "/slots?employee_id=x&service_id=1&date=2026-10-20"},
		{"no service", "/slots?employee_id=1&date=2026-10-20"},
		{"no date", "/slots?employee_id=1&service_id=1"},
		{"date in wrong format", "/slots?employee_id=1&service_id=1&date=20.10.2026"},
		{"step is not a number", "/slots?employee_id=1&service_id=1&date=2026-10-20&step=x"},
		{"step is zero", "/slots?employee_id=1&service_id=1&date=2026-10-20&step=0"},
		{"step is negative", "/slots?employee_id=1&service_id=1&date=2026-10-20&step=-5"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewScheduleHandler(&fakeScheduleServ{}, &fakeSlotsServ{})

			rec := serve(t, "GET /slots", h.GetFreeSlots, http.MethodGet, tt.target, "")

			assertStatus(t, rec, http.StatusBadRequest)
		})
	}
}

func TestGetFreeSlotsPassesQueryToService(t *testing.T) {
	slots := &fakeSlotsServ{slots: []time.Time{time.Date(2026, 10, 20, 10, 0, 0, 0, time.UTC)}}
	h := NewScheduleHandler(&fakeScheduleServ{}, slots)

	rec := serve(t, "GET /slots", h.GetFreeSlots, http.MethodGet, "/slots?employee_id=3&service_id=4&date=2026-10-20&step=30", "")

	assertStatus(t, rec, http.StatusOK)
	if slots.employeeID != 3 || slots.serviceID != 4 || slots.step != 30 {
		t.Errorf("service got employee=%d service=%d step=%d", slots.employeeID, slots.serviceID, slots.step)
	}
	if slots.date.Format("2006-01-02") != "2026-10-20" {
		t.Errorf("date = %v", slots.date)
	}

	var got []time.Time
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(got) != 1 {
		t.Errorf("got %d slots, want 1", len(got))
	}
}

func TestGetFreeSlotsDefaultsStepToZero(t *testing.T) {
	slots := &fakeSlotsServ{}
	h := NewScheduleHandler(&fakeScheduleServ{}, slots)

	serve(t, "GET /slots", h.GetFreeSlots, http.MethodGet, "/slots?employee_id=1&service_id=1&date=2026-10-20", "")

	if slots.step != 0 {
		t.Errorf("step = %d, want 0 so the service picks its default", slots.step)
	}
}

func TestGetFreeSlotsServiceErrors(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
	}{
		{"validation error", models.Invalid("unknown timezone"), http.StatusBadRequest},
		{"storage failure", errors.New("boom"), http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewScheduleHandler(&fakeScheduleServ{}, &fakeSlotsServ{err: tt.err})

			rec := serve(t, "GET /slots", h.GetFreeSlots, http.MethodGet, "/slots?employee_id=1&service_id=1&date=2026-10-20", "")

			assertStatus(t, rec, tt.wantStatus)
		})
	}
}
