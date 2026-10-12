package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/artemydottech/goclients/internal/models"
)

func TestSetStatus(t *testing.T) {
	tests := []struct {
		name       string
		target     string
		body       string
		serv       *fakeAppointmentServ
		wantStatus int
	}{
		{"changed", "/appointments/3/status", `{"status":"confirmed"}`, &fakeAppointmentServ{}, http.StatusNoContent},
		{"bad id", "/appointments/x/status", `{"status":"confirmed"}`, &fakeAppointmentServ{}, http.StatusBadRequest},
		{"malformed body", "/appointments/3/status", `{oops`, &fakeAppointmentServ{}, http.StatusBadRequest},
		{"forbidden transition", "/appointments/3/status", `{"status":"pending"}`, &fakeAppointmentServ{statusErr: models.Invalid("cannot move from cancelled")}, http.StatusBadRequest},
		{"not found", "/appointments/3/status", `{"status":"confirmed"}`, &fakeAppointmentServ{statusErr: models.ErrNotFound}, http.StatusNotFound},
		{"storage failure", "/appointments/3/status", `{"status":"confirmed"}`, &fakeAppointmentServ{statusErr: errors.New("boom")}, http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewAppointmentHandler(tt.serv)

			rec := serve(t, "PUT /appointments/{id}/status", h.SetStatus, http.MethodPut, tt.target, tt.body)

			assertStatus(t, rec, tt.wantStatus)
		})
	}
}

func TestSetStatusPassesStatus(t *testing.T) {
	serv := &fakeAppointmentServ{}
	h := NewAppointmentHandler(serv)

	serve(t, "PUT /appointments/{id}/status", h.SetStatus, http.MethodPut, "/appointments/3/status", `{"status":"no_show"}`)

	if serv.status != models.AppointmentNoShow {
		t.Errorf("status = %q, want no_show", serv.status)
	}
}

func TestReschedule(t *testing.T) {
	body := `{"starts_at":"2026-10-21T11:00:00Z","employee_id":4}`
	tests := []struct {
		name       string
		target     string
		body       string
		serv       *fakeAppointmentServ
		wantStatus int
	}{
		{"moved", "/appointments/3/time", body, &fakeAppointmentServ{rescheduled: models.Appointment{ID: 3}}, http.StatusOK},
		{"bad id", "/appointments/x/time", body, &fakeAppointmentServ{}, http.StatusBadRequest},
		{"malformed body", "/appointments/3/time", `{oops`, &fakeAppointmentServ{}, http.StatusBadRequest},
		{"slot is taken", "/appointments/3/time", body, &fakeAppointmentServ{rescheduleErr: models.Invalid("slot is taken")}, http.StatusBadRequest},
		{"not found", "/appointments/3/time", body, &fakeAppointmentServ{rescheduleErr: models.ErrNotFound}, http.StatusNotFound},
		{"storage failure", "/appointments/3/time", body, &fakeAppointmentServ{rescheduleErr: errors.New("boom")}, http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewAppointmentHandler(tt.serv)

			rec := serve(t, "PUT /appointments/{id}/time", h.Reschedule, http.MethodPut, tt.target, tt.body)

			assertStatus(t, rec, tt.wantStatus)
		})
	}
}

func TestReschedulePassesNewTimeAndEmployee(t *testing.T) {
	serv := &fakeAppointmentServ{rescheduled: models.Appointment{ID: 3, EmployeeID: 4}}
	h := NewAppointmentHandler(serv)

	rec := serve(t, "PUT /appointments/{id}/time", h.Reschedule, http.MethodPut, "/appointments/3/time",
		`{"starts_at":"2026-10-21T11:00:00Z","employee_id":4}`)

	assertStatus(t, rec, http.StatusOK)
	if !serv.rescheduleAt.Equal(time.Date(2026, 10, 21, 11, 0, 0, 0, time.UTC)) {
		t.Errorf("starts_at = %v", serv.rescheduleAt)
	}
	if serv.rescheduleWith != 4 {
		t.Errorf("employee = %d, want 4", serv.rescheduleWith)
	}

	var got models.Appointment
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.EmployeeID != 4 {
		t.Errorf("response employee = %d, want 4", got.EmployeeID)
	}
}

func TestGetClientStats(t *testing.T) {
	tests := []struct {
		name       string
		target     string
		serv       *fakeAppointmentServ
		wantStatus int
	}{
		{"found", "/clients/2/stats", &fakeAppointmentServ{stats: models.ClientStats{ClientID: 2, CompletedVisits: 3}}, http.StatusOK},
		{"not found", "/clients/2/stats", &fakeAppointmentServ{statsErr: models.ErrNotFound}, http.StatusNotFound},
		{"bad id", "/clients/x/stats", &fakeAppointmentServ{}, http.StatusBadRequest},
		{"storage failure", "/clients/2/stats", &fakeAppointmentServ{statsErr: errors.New("boom")}, http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewAppointmentHandler(tt.serv)

			rec := serve(t, "GET /clients/{id}/stats", h.GetClientStats, http.MethodGet, tt.target, "")

			assertStatus(t, rec, tt.wantStatus)
		})
	}
}

func TestParseRangeDefaultsToNextWeek(t *testing.T) {
	from, to, err := parseRange("", "")
	if err != nil {
		t.Fatalf("parseRange: %v", err)
	}

	if got := to.Sub(from); got != 7*24*time.Hour {
		t.Errorf("default window = %v, want 7 days", got)
	}
	if time.Since(from) > time.Minute {
		t.Errorf("default start %v is not close to now", from)
	}
}

func TestParseRangeFromOnlyExtendsAWeek(t *testing.T) {
	from, to, err := parseRange("2026-10-20T08:00:00+05:00", "")
	if err != nil {
		t.Fatalf("parseRange: %v", err)
	}

	if from.Location() != time.UTC {
		t.Errorf("from must be in UTC, got %v", from.Location())
	}
	if !to.Equal(from.AddDate(0, 0, 7)) {
		t.Errorf("to = %v, want a week after %v", to, from)
	}
}

func TestParseRangeErrors(t *testing.T) {
	tests := []struct {
		name string
		from string
		to   string
	}{
		{"bad from", "soon", ""},
		{"bad to", "", "later"},
		{"empty period", "2026-10-20T00:00:00Z", "2026-10-20T00:00:00Z"},
		{"reversed period", "2026-10-20T00:00:00Z", "2026-10-19T00:00:00Z"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, _, err := parseRange(tt.from, tt.to); err == nil {
				t.Fatal("expected an error")
			}
		})
	}
}
