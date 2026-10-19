package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/artemydottech/goclients/internal/models"
)

type fakeScheduleServ struct {
	gotCtx   context.Context
	setErr   error
	setDays  []models.WorkingDay
	days     []models.WorkingDay
	getErr   error
	setEmpID int
}

func (f *fakeScheduleServ) SetEmployeeSchedule(_ context.Context, employeeID int, days []models.WorkingDay) error {
	f.setEmpID = employeeID
	f.setDays = days
	return f.setErr
}

func (f *fakeScheduleServ) GetEmployeeSchedule(ctx context.Context, employeeID int) ([]models.WorkingDay, error) {
	f.gotCtx = ctx
	return f.days, f.getErr
}

type fakeSlotsServ struct {
	slots      []time.Time
	err        error
	employeeID int
	serviceID  int
	date       time.Time
	step       int
}

func (f *fakeSlotsServ) FreeSlots(_ context.Context, employeeID, serviceID int, date time.Time, step int) ([]time.Time, error) {
	f.employeeID, f.serviceID, f.date, f.step = employeeID, serviceID, date, step
	return f.slots, f.err
}

func TestSetEmployeeSchedule(t *testing.T) {
	body := `{"days":[{"weekday":1,"starts_at":"09:00","ends_at":"18:00"}]}`
	tests := []struct {
		name       string
		target     string
		body       string
		serv       *fakeScheduleServ
		wantStatus int
	}{
		{"saved", "/employees/2/schedule", body, &fakeScheduleServ{}, http.StatusNoContent},
		{"bad id", "/employees/x/schedule", body, &fakeScheduleServ{}, http.StatusBadRequest},
		{"malformed body", "/employees/2/schedule", `{oops`, &fakeScheduleServ{}, http.StatusBadRequest},
		{"validation error", "/employees/2/schedule", body, &fakeScheduleServ{setErr: models.Invalid("bad day")}, http.StatusBadRequest},
		{"employee not found", "/employees/2/schedule", body, &fakeScheduleServ{setErr: models.ErrNotFound}, http.StatusNotFound},
		{"storage failure", "/employees/2/schedule", body, &fakeScheduleServ{setErr: errors.New("boom")}, http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewScheduleHandler(tt.serv, &fakeSlotsServ{})

			rec := serve(t, "PUT /employees/{id}/schedule", h.SetEmployeeSchedule, http.MethodPut, tt.target, tt.body)

			assertStatus(t, rec, tt.wantStatus)
		})
	}
}

func TestSetEmployeeSchedulePassesDays(t *testing.T) {
	serv := &fakeScheduleServ{}
	h := NewScheduleHandler(serv, &fakeSlotsServ{})

	serve(t, "PUT /employees/{id}/schedule", h.SetEmployeeSchedule, http.MethodPut, "/employees/9/schedule",
		`{"days":[{"weekday":2,"starts_at":"10:00","ends_at":"19:00","break_starts_at":"13:00","break_ends_at":"14:00"}]}`)

	if serv.setEmpID != 9 {
		t.Errorf("employee = %d, want 9", serv.setEmpID)
	}
	if len(serv.setDays) != 1 || serv.setDays[0].BreakStartsAt != "13:00" {
		t.Errorf("days = %+v", serv.setDays)
	}
}

func TestGetEmployeeSchedule(t *testing.T) {
	serv := &fakeScheduleServ{days: []models.WorkingDay{{Weekday: 1, StartsAt: "09:00", EndsAt: "18:00"}}}
	h := NewScheduleHandler(serv, &fakeSlotsServ{})

	rec := serve(t, "GET /employees/{id}/schedule", h.GetEmployeeSchedule, http.MethodGet, "/employees/2/schedule", "")

	assertStatus(t, rec, http.StatusOK)
	var days []models.WorkingDay
	if err := json.NewDecoder(rec.Body).Decode(&days); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(days) != 1 || days[0].StartsAt != "09:00" {
		t.Errorf("days = %+v", days)
	}
}

func TestGetEmployeeScheduleErrors(t *testing.T) {
	tests := []struct {
		name       string
		target     string
		serv       *fakeScheduleServ
		wantStatus int
	}{
		{"bad id", "/employees/x/schedule", &fakeScheduleServ{}, http.StatusBadRequest},
		{"storage failure", "/employees/2/schedule", &fakeScheduleServ{getErr: errors.New("boom")}, http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewScheduleHandler(tt.serv, &fakeSlotsServ{})

			rec := serve(t, "GET /employees/{id}/schedule", h.GetEmployeeSchedule, http.MethodGet, tt.target, "")

			assertStatus(t, rec, tt.wantStatus)
		})
	}
}

func TestGetEmployeeSchedulePassesRequestContext(t *testing.T) {
	serv := &fakeScheduleServ{}
	h := NewScheduleHandler(serv, &fakeSlotsServ{})

	serveWithContext(t, markedContext(), "GET /employees/{id}/schedule", h.GetEmployeeSchedule, http.MethodGet, "/employees/1/schedule", "")

	assertMarked(t, serv.gotCtx)
}
