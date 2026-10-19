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

type fakeAppointmentServ struct {
	gotCtx    context.Context
	bookID    int64
	bookErr   error
	booked    models.Appointment
	all       []models.Appointment
	allErr    error
	item      models.Appointment
	getErr    error
	statusErr error
	status    models.AppointmentStatus
	deleteErr error

	rescheduled    models.Appointment
	rescheduleErr  error
	rescheduleAt   time.Time
	rescheduleWith int

	stats    models.ClientStats
	statsErr error

	byClientID   int
	byEmployeeID int
	from, to     time.Time
}

func (f *fakeAppointmentServ) Book(ctx context.Context, a models.Appointment) (int64, error) {
	f.gotCtx = ctx
	f.booked = a
	return f.bookID, f.bookErr
}

func (f *fakeAppointmentServ) GetAllAppointments(context.Context) ([]models.Appointment, error) {
	return f.all, f.allErr
}

func (f *fakeAppointmentServ) GetAppointmentsByEmployee(_ context.Context, employeeID int, from, to time.Time) ([]models.Appointment, error) {
	f.byEmployeeID, f.from, f.to = employeeID, from, to
	return f.all, f.allErr
}

func (f *fakeAppointmentServ) GetAppointmentsByClient(_ context.Context, clientID int) ([]models.Appointment, error) {
	f.byClientID = clientID
	return f.all, f.allErr
}

func (f *fakeAppointmentServ) GetAppointmentById(_ context.Context, id int) (models.Appointment, error) {
	return f.item, f.getErr
}

func (f *fakeAppointmentServ) SetStatus(_ context.Context, id int, status models.AppointmentStatus) error {
	f.status = status
	return f.statusErr
}

func (f *fakeAppointmentServ) Reschedule(_ context.Context, id int, startsAt time.Time, employeeID int) (models.Appointment, error) {
	f.rescheduleAt, f.rescheduleWith = startsAt, employeeID
	return f.rescheduled, f.rescheduleErr
}

func (f *fakeAppointmentServ) GetClientStats(_ context.Context, clientID int) (models.ClientStats, error) {
	return f.stats, f.statsErr
}

func (f *fakeAppointmentServ) DeleteAppointmentById(_ context.Context, id int) error {
	return f.deleteErr
}

const appointmentBody = `{"client_id":1,"employee_id":2,"service_id":3,"starts_at":"2026-10-20T10:00:00Z"}`

func TestCreateAppointment(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		serv       *fakeAppointmentServ
		wantStatus int
	}{
		{"created", appointmentBody, &fakeAppointmentServ{bookID: 11}, http.StatusCreated},
		{"malformed body", `{oops`, &fakeAppointmentServ{}, http.StatusBadRequest},
		{"validation error", appointmentBody, &fakeAppointmentServ{bookErr: models.Invalid("slot is taken")}, http.StatusBadRequest},
		{"storage failure", appointmentBody, &fakeAppointmentServ{bookErr: errors.New("boom")}, http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewAppointmentHandler(tt.serv)

			rec := serve(t, "POST /appointments", h.CreateAppointment, http.MethodPost, "/appointments", tt.body)

			assertStatus(t, rec, tt.wantStatus)
		})
	}
}

func TestCreateAppointmentPassesInput(t *testing.T) {
	serv := &fakeAppointmentServ{bookID: 11}
	h := NewAppointmentHandler(serv)

	rec := serve(t, "POST /appointments", h.CreateAppointment, http.MethodPost, "/appointments", appointmentBody)

	assertStatus(t, rec, http.StatusCreated)
	if serv.booked.ClientID != 1 || serv.booked.EmployeeID != 2 || serv.booked.ServiceID != 3 {
		t.Errorf("booked = %+v", serv.booked)
	}

	var body map[string]int64
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body["id"] != 11 {
		t.Errorf("id = %d, want 11", body["id"])
	}
}

func TestGetAllAppointmentsPicksTheRightQuery(t *testing.T) {
	serv := &fakeAppointmentServ{all: []models.Appointment{{ID: 1}}}
	h := NewAppointmentHandler(serv)

	rec := serve(t, "GET /appointments", h.GetAllAppointments, http.MethodGet, "/appointments?client_id=5", "")

	assertStatus(t, rec, http.StatusOK)
	if serv.byClientID != 5 {
		t.Errorf("client filter = %d, want 5", serv.byClientID)
	}

	rec = serve(t, "GET /appointments", h.GetAllAppointments, http.MethodGet, "/appointments?employee_id=6&from=2026-10-20T00:00:00Z&to=2026-10-22T00:00:00Z", "")

	assertStatus(t, rec, http.StatusOK)
	if serv.byEmployeeID != 6 {
		t.Errorf("employee filter = %d, want 6", serv.byEmployeeID)
	}
	if serv.to.Sub(serv.from) != 48*time.Hour {
		t.Errorf("range = %v - %v, want 48h", serv.from, serv.to)
	}
}

func TestGetAllAppointmentsErrors(t *testing.T) {
	tests := []struct {
		name       string
		target     string
		serv       *fakeAppointmentServ
		wantStatus int
	}{
		{"all", "/appointments", &fakeAppointmentServ{}, http.StatusOK},
		{"bad client id", "/appointments?client_id=x", &fakeAppointmentServ{}, http.StatusBadRequest},
		{"bad employee id", "/appointments?employee_id=x", &fakeAppointmentServ{}, http.StatusBadRequest},
		{"bad from", "/appointments?employee_id=1&from=yesterday", &fakeAppointmentServ{}, http.StatusBadRequest},
		{"bad to", "/appointments?employee_id=1&to=tomorrow", &fakeAppointmentServ{}, http.StatusBadRequest},
		{"to before from", "/appointments?employee_id=1&from=2026-10-20T00:00:00Z&to=2026-10-19T00:00:00Z", &fakeAppointmentServ{}, http.StatusBadRequest},
		{"storage failure", "/appointments", &fakeAppointmentServ{allErr: errors.New("boom")}, http.StatusInternalServerError},
		{"storage failure by client", "/appointments?client_id=1", &fakeAppointmentServ{allErr: errors.New("boom")}, http.StatusInternalServerError},
		{"storage failure by employee", "/appointments?employee_id=1", &fakeAppointmentServ{allErr: errors.New("boom")}, http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewAppointmentHandler(tt.serv)

			rec := serve(t, "GET /appointments", h.GetAllAppointments, http.MethodGet, tt.target, "")

			assertStatus(t, rec, tt.wantStatus)
		})
	}
}

func TestGetAppointmentById(t *testing.T) {
	tests := []struct {
		name       string
		target     string
		serv       *fakeAppointmentServ
		wantStatus int
	}{
		{"found", "/appointments/3", &fakeAppointmentServ{item: models.Appointment{ID: 3}}, http.StatusOK},
		{"not found", "/appointments/3", &fakeAppointmentServ{getErr: models.ErrNotFound}, http.StatusNotFound},
		{"bad id", "/appointments/x", &fakeAppointmentServ{}, http.StatusBadRequest},
		{"storage failure", "/appointments/3", &fakeAppointmentServ{getErr: errors.New("boom")}, http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewAppointmentHandler(tt.serv)

			rec := serve(t, "GET /appointments/{id}", h.GetAppointmentById, http.MethodGet, tt.target, "")

			assertStatus(t, rec, tt.wantStatus)
		})
	}
}

func TestDeleteAppointment(t *testing.T) {
	tests := []struct {
		name       string
		target     string
		serv       *fakeAppointmentServ
		wantStatus int
	}{
		{"deleted", "/appointments/3", &fakeAppointmentServ{}, http.StatusNoContent},
		{"not found", "/appointments/3", &fakeAppointmentServ{deleteErr: models.ErrNotFound}, http.StatusNotFound},
		{"bad id", "/appointments/x", &fakeAppointmentServ{}, http.StatusBadRequest},
		{"storage failure", "/appointments/3", &fakeAppointmentServ{deleteErr: errors.New("boom")}, http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewAppointmentHandler(tt.serv)

			rec := serve(t, "DELETE /appointments/{id}", h.DeleteAppointment, http.MethodDelete, tt.target, "")

			assertStatus(t, rec, tt.wantStatus)
		})
	}
}

func TestCreateAppointmentPassesRequestContext(t *testing.T) {
	serv := &fakeAppointmentServ{}
	h := NewAppointmentHandler(serv)

	serveWithContext(t, markedContext(), "POST /appointments", h.CreateAppointment, http.MethodPost, "/appointments", `{"client_id":1}`)

	assertMarked(t, serv.gotCtx)
}
