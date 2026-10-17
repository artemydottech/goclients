package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/artemydottech/goclients/internal/models"
)

type fakeAssignmentServ struct {
	setErr      error
	setEmployee int
	setServices []int
	services    []models.Service
	employees   []models.Employee
	listErr     error
}

func (f *fakeAssignmentServ) SetEmployeeServices(_ context.Context, employeeID int, serviceIDs []int) error {
	f.setEmployee = employeeID
	f.setServices = serviceIDs
	return f.setErr
}

func (f *fakeAssignmentServ) GetServicesByEmployee(_ context.Context, employeeID int) ([]models.Service, error) {
	return f.services, f.listErr
}

func (f *fakeAssignmentServ) GetEmployeesByService(_ context.Context, serviceID int) ([]models.Employee, error) {
	return f.employees, f.listErr
}

func TestSetEmployeeServices(t *testing.T) {
	tests := []struct {
		name       string
		target     string
		body       string
		serv       *fakeAssignmentServ
		wantStatus int
	}{
		{"saved", "/employees/2/services", `{"service_ids":[1,2]}`, &fakeAssignmentServ{}, http.StatusNoContent},
		{"bad id", "/employees/x/services", `{"service_ids":[1]}`, &fakeAssignmentServ{}, http.StatusBadRequest},
		{"malformed body", "/employees/2/services", `{oops`, &fakeAssignmentServ{}, http.StatusBadRequest},
		{"validation error", "/employees/2/services", `{"service_ids":[1]}`, &fakeAssignmentServ{setErr: models.Invalid("service from another company")}, http.StatusBadRequest},
		{"employee not found", "/employees/2/services", `{"service_ids":[1]}`, &fakeAssignmentServ{setErr: models.ErrNotFound}, http.StatusNotFound},
		{"storage failure", "/employees/2/services", `{"service_ids":[1]}`, &fakeAssignmentServ{setErr: errors.New("boom")}, http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewAssignmentHandler(tt.serv)

			rec := serve(t, "PUT /employees/{id}/services", h.SetEmployeeServices, http.MethodPut, tt.target, tt.body)

			assertStatus(t, rec, tt.wantStatus)
		})
	}
}

func TestSetEmployeeServicesPassesArguments(t *testing.T) {
	serv := &fakeAssignmentServ{}
	h := NewAssignmentHandler(serv)

	serve(t, "PUT /employees/{id}/services", h.SetEmployeeServices, http.MethodPut, "/employees/7/services", `{"service_ids":[3,4]}`)

	if serv.setEmployee != 7 {
		t.Errorf("employee = %d, want 7", serv.setEmployee)
	}
	if len(serv.setServices) != 2 || serv.setServices[0] != 3 || serv.setServices[1] != 4 {
		t.Errorf("services = %v, want [3 4]", serv.setServices)
	}
}

func TestGetEmployeeServices(t *testing.T) {
	h := NewAssignmentHandler(&fakeAssignmentServ{services: []models.Service{{ID: 1}, {ID: 2}}})

	rec := serve(t, "GET /employees/{id}/services", h.GetEmployeeServices, http.MethodGet, "/employees/1/services", "")

	assertStatus(t, rec, http.StatusOK)
	var got []models.Service
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("got %d services, want 2", len(got))
	}
}

func TestGetEmployeeServicesErrors(t *testing.T) {
	tests := []struct {
		name       string
		target     string
		serv       *fakeAssignmentServ
		wantStatus int
	}{
		{"bad id", "/employees/x/services", &fakeAssignmentServ{}, http.StatusBadRequest},
		{"storage failure", "/employees/1/services", &fakeAssignmentServ{listErr: errors.New("boom")}, http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewAssignmentHandler(tt.serv)

			rec := serve(t, "GET /employees/{id}/services", h.GetEmployeeServices, http.MethodGet, tt.target, "")

			assertStatus(t, rec, tt.wantStatus)
		})
	}
}

func TestGetServiceEmployees(t *testing.T) {
	tests := []struct {
		name       string
		target     string
		serv       *fakeAssignmentServ
		wantStatus int
	}{
		{"found", "/services/1/employees", &fakeAssignmentServ{employees: []models.Employee{{ID: 5}}}, http.StatusOK},
		{"bad id", "/services/x/employees", &fakeAssignmentServ{}, http.StatusBadRequest},
		{"storage failure", "/services/1/employees", &fakeAssignmentServ{listErr: errors.New("boom")}, http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewAssignmentHandler(tt.serv)

			rec := serve(t, "GET /services/{id}/employees", h.GetServiceEmployees, http.MethodGet, tt.target, "")

			assertStatus(t, rec, tt.wantStatus)
		})
	}
}
