package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/artemydottech/goclients/internal/models"
)

type fakeEmployeeServ struct {
	gotPage   models.Page
	updatedID int
	updated   models.Employee
	updateErr error
	gotCtx    context.Context
	createID  int64
	createErr error
	created   models.Employee
	listItems []models.Employee
	listErr   error
	item      models.Employee
	getErr    error
	deleteErr error
}

func (f *fakeEmployeeServ) CreateEmployee(_ context.Context, v models.Employee) (int64, error) {
	f.created = v
	return f.createID, f.createErr
}

func (f *fakeEmployeeServ) GetAllEmployees(_ context.Context, page models.Page) ([]models.Employee, error) {
	f.gotPage = page
	return f.listItems, f.listErr
}

func (f *fakeEmployeeServ) GetEmployeeById(ctx context.Context, id int) (models.Employee, error) {
	f.gotCtx = ctx
	return f.item, f.getErr
}

func (f *fakeEmployeeServ) UpdateEmployee(_ context.Context, id int, v models.Employee) error {
	f.updatedID = id
	f.updated = v
	return f.updateErr
}

func (f *fakeEmployeeServ) DeleteEmployeeById(_ context.Context, id int) error {
	return f.deleteErr
}

func TestEmployeeCreate(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		serv       *fakeEmployeeServ
		wantStatus int
	}{
		{"created", `{"company_id":1,"name":"Anna"}`, &fakeEmployeeServ{createID: 9}, http.StatusCreated},
		{"malformed body", `{oops`, &fakeEmployeeServ{}, http.StatusBadRequest},
		{"validation error", `{"company_id":1,"name":"Anna"}`, &fakeEmployeeServ{createErr: models.Invalid("bad input")}, http.StatusBadRequest},
		{"storage failure", `{"company_id":1,"name":"Anna"}`, &fakeEmployeeServ{createErr: errors.New("disk full")}, http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewEmployeeHandler(tt.serv)

			rec := serve(t, "POST /employees", h.CreateEmployee, http.MethodPost, "/employees", tt.body)

			assertStatus(t, rec, tt.wantStatus)
		})
	}
}

func TestEmployeeCreateReturnsID(t *testing.T) {
	serv := &fakeEmployeeServ{createID: 77}
	h := NewEmployeeHandler(serv)

	rec := serve(t, "POST /employees", h.CreateEmployee, http.MethodPost, "/employees", `{"company_id":1,"name":"Anna"}`)

	assertStatus(t, rec, http.StatusCreated)
	var body map[string]int64
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body["id"] != 77 {
		t.Errorf("id = %d, want 77", body["id"])
	}
}

func TestEmployeeGetAll(t *testing.T) {
	tests := []struct {
		name       string
		serv       *fakeEmployeeServ
		wantStatus int
		wantLen    int
	}{
		{"two items", &fakeEmployeeServ{listItems: []models.Employee{{ID: 1}, {ID: 2}}}, http.StatusOK, 2},
		{"empty", &fakeEmployeeServ{}, http.StatusOK, 0},
		{"storage failure", &fakeEmployeeServ{listErr: errors.New("boom")}, http.StatusInternalServerError, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewEmployeeHandler(tt.serv)

			rec := serve(t, "GET /employees", h.GetAllEmployees, http.MethodGet, "/employees", "")

			assertStatus(t, rec, tt.wantStatus)
			if tt.wantStatus != http.StatusOK {
				return
			}
			var items []models.Employee
			if err := json.NewDecoder(rec.Body).Decode(&items); err != nil {
				t.Fatalf("decode: %v", err)
			}
			if len(items) != tt.wantLen {
				t.Errorf("got %d items, want %d", len(items), tt.wantLen)
			}
		})
	}
}

func TestEmployeeGetById(t *testing.T) {
	tests := []struct {
		name       string
		target     string
		serv       *fakeEmployeeServ
		wantStatus int
	}{
		{"found", "/employees/3", &fakeEmployeeServ{item: models.Employee{ID: 3}}, http.StatusOK},
		{"not found", "/employees/3", &fakeEmployeeServ{getErr: models.ErrNotFound}, http.StatusNotFound},
		{"bad id", "/employees/abc", &fakeEmployeeServ{}, http.StatusBadRequest},
		{"storage failure", "/employees/3", &fakeEmployeeServ{getErr: errors.New("boom")}, http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewEmployeeHandler(tt.serv)

			rec := serve(t, "GET /employees/{id}", h.GetEmployeeById, http.MethodGet, tt.target, "")

			assertStatus(t, rec, tt.wantStatus)
		})
	}
}

func TestEmployeeDelete(t *testing.T) {
	tests := []struct {
		name       string
		target     string
		serv       *fakeEmployeeServ
		wantStatus int
	}{
		{"deleted", "/employees/3", &fakeEmployeeServ{}, http.StatusNoContent},
		{"not found", "/employees/3", &fakeEmployeeServ{deleteErr: models.ErrNotFound}, http.StatusNotFound},
		{"bad id", "/employees/x", &fakeEmployeeServ{}, http.StatusBadRequest},
		{"storage failure", "/employees/3", &fakeEmployeeServ{deleteErr: errors.New("boom")}, http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewEmployeeHandler(tt.serv)

			rec := serve(t, "DELETE /employees/{id}", h.DeleteEmployee, http.MethodDelete, tt.target, "")

			assertStatus(t, rec, tt.wantStatus)
		})
	}
}

func TestGetEmployeeByIdPassesRequestContext(t *testing.T) {
	serv := &fakeEmployeeServ{}
	h := NewEmployeeHandler(serv)

	serveWithContext(t, markedContext(), "GET /employees/{id}", h.GetEmployeeById, http.MethodGet, "/employees/1", "")

	assertMarked(t, serv.gotCtx)
}

func TestUpdateEmployee(t *testing.T) {
	body := `{"name":"Anna","surname":"Ivanova"}`
	tests := []struct {
		name       string
		target     string
		body       string
		serv       *fakeEmployeeServ
		wantStatus int
	}{
		{"updated", "/employees/3", body, &fakeEmployeeServ{}, http.StatusNoContent},
		{"bad id", "/employees/x", body, &fakeEmployeeServ{}, http.StatusBadRequest},
		{"malformed body", "/employees/3", `{oops`, &fakeEmployeeServ{}, http.StatusBadRequest},
		{"validation error", "/employees/3", body, &fakeEmployeeServ{updateErr: models.Invalid("employee name cannot be empty")}, http.StatusBadRequest},
		{"not found", "/employees/3", body, &fakeEmployeeServ{updateErr: models.ErrNotFound}, http.StatusNotFound},
		{"storage failure", "/employees/3", body, &fakeEmployeeServ{updateErr: errors.New("boom")}, http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewEmployeeHandler(tt.serv)

			rec := serve(t, "PUT /employees/{id}", h.UpdateEmployee, http.MethodPut, tt.target, tt.body)

			assertStatus(t, rec, tt.wantStatus)
		})
	}
}

func TestUpdateEmployeePassesIDAndBody(t *testing.T) {
	serv := &fakeEmployeeServ{}
	h := NewEmployeeHandler(serv)

	serve(t, "PUT /employees/{id}", h.UpdateEmployee, http.MethodPut, "/employees/8", `{"name":"Anna","surname":"Ivanova","position":"Stylist"}`)

	if serv.updatedID != 8 || serv.updated.Position != "Stylist" {
		t.Errorf("service got id=%d employee=%+v", serv.updatedID, serv.updated)
	}
}

func TestGetAllEmployeesPassesThePage(t *testing.T) {
	serv := &fakeEmployeeServ{}
	h := NewEmployeeHandler(serv)

	rec := serve(t, "GET /employees", h.GetAllEmployees, http.MethodGet, "/employees?limit=5", "")

	assertStatus(t, rec, http.StatusOK)
	if serv.gotPage != (models.Page{Limit: 5}) {
		t.Errorf("service got page %+v", serv.gotPage)
	}
}

func TestGetAllEmployeesRejectsABadPage(t *testing.T) {
	h := NewEmployeeHandler(&fakeEmployeeServ{})

	rec := serve(t, "GET /employees", h.GetAllEmployees, http.MethodGet, "/employees?offset=-3", "")

	assertStatus(t, rec, http.StatusBadRequest)
}

func (f *fakeEmployeeServ) GetEmployeesByCompany(ctx context.Context, id int, page models.Page) ([]models.Employee, error) {
	f.gotCtx = ctx
	f.gotPage = page
	return f.listItems, f.listErr
}

func TestEmployeeDirectoryInvalidCompanies(t *testing.T) {
	h := NewEmployeeHandler(&fakeEmployeeServ{})
	for _, id := range []string{"0", "-1", "abc", "99999999999999999999"} {
		rec := serve(t, "GET /employees", h.GetAllEmployees, http.MethodGet, "/employees?company_id="+id, "")
		assertStatus(t, rec, http.StatusBadRequest)
	}
}
func TestEmployeeDirectoryPassesPageAndContext(t *testing.T) {
	serv := &fakeEmployeeServ{}
	h := NewEmployeeHandler(serv)
	rec := serveWithContext(t, markedContext(), "GET /employees", h.GetAllEmployees, http.MethodGet, "/employees?company_id=1&limit=2&offset=1", "")
	assertStatus(t, rec, http.StatusOK)
	assertMarked(t, serv.gotCtx)
	if serv.gotPage != (models.Page{Limit: 2, Offset: 1}) {
		t.Fatal(serv.gotPage)
	}
}
