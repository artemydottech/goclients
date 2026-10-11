package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/artemydottech/goclients/internal/models"
)

type fakeCompanyServ struct {
	createID  int64
	createErr error
	created   models.Company
	listItems []models.Company
	listErr   error
	item      models.Company
	getErr    error
	deleteErr error
}

func (f *fakeCompanyServ) CreateCompany(v models.Company) (int64, error) {
	f.created = v
	return f.createID, f.createErr
}

func (f *fakeCompanyServ) GetAllCompanies() ([]models.Company, error) {
	return f.listItems, f.listErr
}

func (f *fakeCompanyServ) GetCompanyById(id int) (models.Company, error) {
	return f.item, f.getErr
}

func (f *fakeCompanyServ) DeleteCompanyById(id int) error {
	return f.deleteErr
}

func TestCompanyCreate(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		serv       *fakeCompanyServ
		wantStatus int
	}{
		{"created", `{"name":"Romashka"}`, &fakeCompanyServ{createID: 9}, http.StatusCreated},
		{"malformed body", `{oops`, &fakeCompanyServ{}, http.StatusBadRequest},
		{"validation error", `{"name":"Romashka"}`, &fakeCompanyServ{createErr: models.Invalid("bad input")}, http.StatusBadRequest},
		{"storage failure", `{"name":"Romashka"}`, &fakeCompanyServ{createErr: errors.New("disk full")}, http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewCompanyHandler(tt.serv)

			rec := serve(t, "POST /companies", h.CreateCompany, http.MethodPost, "/companies", tt.body)

			assertStatus(t, rec, tt.wantStatus)
		})
	}
}

func TestCompanyCreateReturnsID(t *testing.T) {
	serv := &fakeCompanyServ{createID: 77}
	h := NewCompanyHandler(serv)

	rec := serve(t, "POST /companies", h.CreateCompany, http.MethodPost, "/companies", `{"name":"Romashka"}`)

	assertStatus(t, rec, http.StatusCreated)
	var body map[string]int64
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body["id"] != 77 {
		t.Errorf("id = %d, want 77", body["id"])
	}
}

func TestCompanyGetAll(t *testing.T) {
	tests := []struct {
		name       string
		serv       *fakeCompanyServ
		wantStatus int
		wantLen    int
	}{
		{"two items", &fakeCompanyServ{listItems: []models.Company{{ID: 1}, {ID: 2}}}, http.StatusOK, 2},
		{"empty", &fakeCompanyServ{}, http.StatusOK, 0},
		{"storage failure", &fakeCompanyServ{listErr: errors.New("boom")}, http.StatusInternalServerError, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewCompanyHandler(tt.serv)

			rec := serve(t, "GET /companies", h.GetAllCompanies, http.MethodGet, "/companies", "")

			assertStatus(t, rec, tt.wantStatus)
			if tt.wantStatus != http.StatusOK {
				return
			}
			var items []models.Company
			if err := json.NewDecoder(rec.Body).Decode(&items); err != nil {
				t.Fatalf("decode: %v", err)
			}
			if len(items) != tt.wantLen {
				t.Errorf("got %d items, want %d", len(items), tt.wantLen)
			}
		})
	}
}

func TestCompanyGetById(t *testing.T) {
	tests := []struct {
		name       string
		target     string
		serv       *fakeCompanyServ
		wantStatus int
	}{
		{"found", "/companies/3", &fakeCompanyServ{item: models.Company{ID: 3}}, http.StatusOK},
		{"not found", "/companies/3", &fakeCompanyServ{getErr: models.ErrNotFound}, http.StatusNotFound},
		{"bad id", "/companies/abc", &fakeCompanyServ{}, http.StatusBadRequest},
		{"storage failure", "/companies/3", &fakeCompanyServ{getErr: errors.New("boom")}, http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewCompanyHandler(tt.serv)

			rec := serve(t, "GET /companies/{id}", h.GetCompanyById, http.MethodGet, tt.target, "")

			assertStatus(t, rec, tt.wantStatus)
		})
	}
}

func TestCompanyDelete(t *testing.T) {
	tests := []struct {
		name       string
		target     string
		serv       *fakeCompanyServ
		wantStatus int
	}{
		{"deleted", "/companies/3", &fakeCompanyServ{}, http.StatusNoContent},
		{"not found", "/companies/3", &fakeCompanyServ{deleteErr: models.ErrNotFound}, http.StatusNotFound},
		{"bad id", "/companies/x", &fakeCompanyServ{}, http.StatusBadRequest},
		{"storage failure", "/companies/3", &fakeCompanyServ{deleteErr: errors.New("boom")}, http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewCompanyHandler(tt.serv)

			rec := serve(t, "DELETE /companies/{id}", h.DeleteCompany, http.MethodDelete, tt.target, "")

			assertStatus(t, rec, tt.wantStatus)
		})
	}
}
