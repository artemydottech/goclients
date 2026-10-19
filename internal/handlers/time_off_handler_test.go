package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/artemydottech/goclients/internal/models"
)

type fakeTimeOffServ struct {
	gotCtx     context.Context
	createID   int64
	createErr  error
	employeeID int
	created    models.TimeOff
	periods    []models.TimeOff
	listErr    error
	deleteErr  error
}

func (f *fakeTimeOffServ) CreateTimeOff(_ context.Context, employeeID int, t models.TimeOff) (int64, error) {
	f.employeeID = employeeID
	f.created = t
	return f.createID, f.createErr
}

func (f *fakeTimeOffServ) GetTimeOffByEmployee(ctx context.Context, employeeID int) ([]models.TimeOff, error) {
	f.gotCtx = ctx
	return f.periods, f.listErr
}

func (f *fakeTimeOffServ) DeleteTimeOffById(_ context.Context, id int) error {
	return f.deleteErr
}

const timeOffBody = `{"starts_at":"2026-10-20T00:00:00Z","ends_at":"2026-10-25T00:00:00Z","reason":"vacation"}`

func TestCreateTimeOff(t *testing.T) {
	tests := []struct {
		name       string
		target     string
		body       string
		serv       *fakeTimeOffServ
		wantStatus int
	}{
		{"created", "/employees/2/time-off", timeOffBody, &fakeTimeOffServ{createID: 8}, http.StatusCreated},
		{"bad id", "/employees/x/time-off", timeOffBody, &fakeTimeOffServ{}, http.StatusBadRequest},
		{"malformed body", "/employees/2/time-off", `{oops`, &fakeTimeOffServ{}, http.StatusBadRequest},
		{"bad date", "/employees/2/time-off", `{"starts_at":"tomorrow"}`, &fakeTimeOffServ{}, http.StatusBadRequest},
		{"validation error", "/employees/2/time-off", timeOffBody, &fakeTimeOffServ{createErr: models.Invalid("period overlaps an appointment")}, http.StatusBadRequest},
		{"employee not found", "/employees/2/time-off", timeOffBody, &fakeTimeOffServ{createErr: models.ErrNotFound}, http.StatusNotFound},
		{"storage failure", "/employees/2/time-off", timeOffBody, &fakeTimeOffServ{createErr: errors.New("boom")}, http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewTimeOffHandler(tt.serv)

			rec := serve(t, "POST /employees/{id}/time-off", h.CreateTimeOff, http.MethodPost, tt.target, tt.body)

			assertStatus(t, rec, tt.wantStatus)
		})
	}
}

func TestCreateTimeOffPassesEmployeeAndPeriod(t *testing.T) {
	serv := &fakeTimeOffServ{createID: 3}
	h := NewTimeOffHandler(serv)

	rec := serve(t, "POST /employees/{id}/time-off", h.CreateTimeOff, http.MethodPost, "/employees/6/time-off", timeOffBody)

	assertStatus(t, rec, http.StatusCreated)
	if serv.employeeID != 6 {
		t.Errorf("employee = %d, want 6", serv.employeeID)
	}
	if serv.created.Reason != "vacation" || serv.created.StartsAt.Day() != 20 {
		t.Errorf("period = %+v", serv.created)
	}

	var body map[string]int64
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body["id"] != 3 {
		t.Errorf("id = %d, want 3", body["id"])
	}
}

func TestGetTimeOff(t *testing.T) {
	tests := []struct {
		name       string
		target     string
		serv       *fakeTimeOffServ
		wantStatus int
	}{
		{"found", "/employees/2/time-off", &fakeTimeOffServ{periods: []models.TimeOff{{ID: 1}}}, http.StatusOK},
		{"bad id", "/employees/x/time-off", &fakeTimeOffServ{}, http.StatusBadRequest},
		{"storage failure", "/employees/2/time-off", &fakeTimeOffServ{listErr: errors.New("boom")}, http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewTimeOffHandler(tt.serv)

			rec := serve(t, "GET /employees/{id}/time-off", h.GetTimeOff, http.MethodGet, tt.target, "")

			assertStatus(t, rec, tt.wantStatus)
		})
	}
}

func TestDeleteTimeOff(t *testing.T) {
	tests := []struct {
		name       string
		target     string
		serv       *fakeTimeOffServ
		wantStatus int
	}{
		{"deleted", "/time-off/4", &fakeTimeOffServ{}, http.StatusNoContent},
		{"not found", "/time-off/4", &fakeTimeOffServ{deleteErr: models.ErrNotFound}, http.StatusNotFound},
		{"bad id", "/time-off/x", &fakeTimeOffServ{}, http.StatusBadRequest},
		{"storage failure", "/time-off/4", &fakeTimeOffServ{deleteErr: errors.New("boom")}, http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewTimeOffHandler(tt.serv)

			rec := serve(t, "DELETE /time-off/{id}", h.DeleteTimeOff, http.MethodDelete, tt.target, "")

			assertStatus(t, rec, tt.wantStatus)
		})
	}
}

func TestGetTimeOffPassesRequestContext(t *testing.T) {
	serv := &fakeTimeOffServ{}
	h := NewTimeOffHandler(serv)

	serveWithContext(t, markedContext(), "GET /employees/{id}/time-off", h.GetTimeOff, http.MethodGet, "/employees/1/time-off", "")

	assertMarked(t, serv.gotCtx)
}
