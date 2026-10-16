package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/artemydottech/goclients/internal/models"
)

type fakeCatalogServ struct {
	createID  int64
	createErr error
	created   models.Service
	listItems []models.Service
	listErr   error
	companyID int
	item      models.Service
	getErr    error
	deleteErr error
}

func (f *fakeCatalogServ) CreateService(_ context.Context, v models.Service) (int64, error) {
	f.created = v
	return f.createID, f.createErr
}

func (f *fakeCatalogServ) GetAllServices(context.Context) ([]models.Service, error) {
	return f.listItems, f.listErr
}

func (f *fakeCatalogServ) GetServicesByCompany(_ context.Context, companyID int) ([]models.Service, error) {
	f.companyID = companyID
	return f.listItems, f.listErr
}

func (f *fakeCatalogServ) GetServiceById(_ context.Context, id int) (models.Service, error) {
	return f.item, f.getErr
}

func (f *fakeCatalogServ) DeleteServiceById(_ context.Context, id int) error {
	return f.deleteErr
}

func TestServiceCreate(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		serv       *fakeCatalogServ
		wantStatus int
	}{
		{"created", `{"company_id":1,"name":"Haircut","duration":30}`, &fakeCatalogServ{createID: 9}, http.StatusCreated},
		{"malformed body", `{oops`, &fakeCatalogServ{}, http.StatusBadRequest},
		{"validation error", `{"company_id":1,"name":"Haircut","duration":30}`, &fakeCatalogServ{createErr: models.Invalid("bad input")}, http.StatusBadRequest},
		{"storage failure", `{"company_id":1,"name":"Haircut","duration":30}`, &fakeCatalogServ{createErr: errors.New("disk full")}, http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewServiceHandler(tt.serv)

			rec := serve(t, "POST /services", h.CreateService, http.MethodPost, "/services", tt.body)

			assertStatus(t, rec, tt.wantStatus)
		})
	}
}

func TestServiceCreateReturnsID(t *testing.T) {
	serv := &fakeCatalogServ{createID: 77}
	h := NewServiceHandler(serv)

	rec := serve(t, "POST /services", h.CreateService, http.MethodPost, "/services", `{"company_id":1,"name":"Haircut","duration":30}`)

	assertStatus(t, rec, http.StatusCreated)
	var body map[string]int64
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body["id"] != 77 {
		t.Errorf("id = %d, want 77", body["id"])
	}
}

func TestServiceGetAll(t *testing.T) {
	tests := []struct {
		name       string
		serv       *fakeCatalogServ
		wantStatus int
		wantLen    int
	}{
		{"two items", &fakeCatalogServ{listItems: []models.Service{{ID: 1}, {ID: 2}}}, http.StatusOK, 2},
		{"empty", &fakeCatalogServ{}, http.StatusOK, 0},
		{"storage failure", &fakeCatalogServ{listErr: errors.New("boom")}, http.StatusInternalServerError, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewServiceHandler(tt.serv)

			rec := serve(t, "GET /services", h.GetAllServices, http.MethodGet, "/services", "")

			assertStatus(t, rec, tt.wantStatus)
			if tt.wantStatus != http.StatusOK {
				return
			}
			var items []models.Service
			if err := json.NewDecoder(rec.Body).Decode(&items); err != nil {
				t.Fatalf("decode: %v", err)
			}
			if len(items) != tt.wantLen {
				t.Errorf("got %d items, want %d", len(items), tt.wantLen)
			}
		})
	}
}

func TestServiceGetAllFiltersByCompany(t *testing.T) {
	serv := &fakeCatalogServ{listItems: []models.Service{{ID: 1}}}
	h := NewServiceHandler(serv)

	rec := serve(t, "GET /services", h.GetAllServices, http.MethodGet, "/services?company_id=4", "")

	assertStatus(t, rec, http.StatusOK)
	if serv.companyID != 4 {
		t.Errorf("companyID = %d, want 4", serv.companyID)
	}
}

func TestServiceGetAllRejectsBadCompanyID(t *testing.T) {
	h := NewServiceHandler(&fakeCatalogServ{})

	rec := serve(t, "GET /services", h.GetAllServices, http.MethodGet, "/services?company_id=abc", "")

	assertStatus(t, rec, http.StatusBadRequest)
}

func TestServiceGetById(t *testing.T) {
	tests := []struct {
		name       string
		target     string
		serv       *fakeCatalogServ
		wantStatus int
	}{
		{"found", "/services/3", &fakeCatalogServ{item: models.Service{ID: 3}}, http.StatusOK},
		{"not found", "/services/3", &fakeCatalogServ{getErr: models.ErrNotFound}, http.StatusNotFound},
		{"bad id", "/services/abc", &fakeCatalogServ{}, http.StatusBadRequest},
		{"storage failure", "/services/3", &fakeCatalogServ{getErr: errors.New("boom")}, http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewServiceHandler(tt.serv)

			rec := serve(t, "GET /services/{id}", h.GetServiceById, http.MethodGet, tt.target, "")

			assertStatus(t, rec, tt.wantStatus)
		})
	}
}

func TestServiceDelete(t *testing.T) {
	tests := []struct {
		name       string
		target     string
		serv       *fakeCatalogServ
		wantStatus int
	}{
		{"deleted", "/services/3", &fakeCatalogServ{}, http.StatusNoContent},
		{"not found", "/services/3", &fakeCatalogServ{deleteErr: models.ErrNotFound}, http.StatusNotFound},
		{"bad id", "/services/x", &fakeCatalogServ{}, http.StatusBadRequest},
		{"storage failure", "/services/3", &fakeCatalogServ{deleteErr: errors.New("boom")}, http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewServiceHandler(tt.serv)

			rec := serve(t, "DELETE /services/{id}", h.DeleteService, http.MethodDelete, tt.target, "")

			assertStatus(t, rec, tt.wantStatus)
		})
	}
}
