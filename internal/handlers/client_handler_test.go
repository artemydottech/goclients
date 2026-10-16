package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/artemydottech/goclients/internal/models"
)

type fakeClientServ struct {
	createID  int64
	createErr error
	created   models.Client
	listItems []models.Client
	listErr   error
	companyID int
	item      models.Client
	getErr    error
	deleteErr error
}

func (f *fakeClientServ) CreateClient(_ context.Context, v models.Client) (int64, error) {
	f.created = v
	return f.createID, f.createErr
}

func (f *fakeClientServ) GetAllClients(context.Context) ([]models.Client, error) {
	return f.listItems, f.listErr
}

func (f *fakeClientServ) GetClientsByCompany(_ context.Context, companyID int) ([]models.Client, error) {
	f.companyID = companyID
	return f.listItems, f.listErr
}

func (f *fakeClientServ) GetClientById(_ context.Context, id int) (models.Client, error) {
	return f.item, f.getErr
}

func (f *fakeClientServ) DeleteClientById(id int) error {
	return f.deleteErr
}

func TestClientCreate(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		serv       *fakeClientServ
		wantStatus int
	}{
		{"created", `{"company_id":1,"name":"Ivan","phone":"+79990000000"}`, &fakeClientServ{createID: 9}, http.StatusCreated},
		{"malformed body", `{oops`, &fakeClientServ{}, http.StatusBadRequest},
		{"validation error", `{"company_id":1,"name":"Ivan","phone":"+79990000000"}`, &fakeClientServ{createErr: models.Invalid("bad input")}, http.StatusBadRequest},
		{"storage failure", `{"company_id":1,"name":"Ivan","phone":"+79990000000"}`, &fakeClientServ{createErr: errors.New("disk full")}, http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewClientHandler(tt.serv)

			rec := serve(t, "POST /clients", h.CreateClient, http.MethodPost, "/clients", tt.body)

			assertStatus(t, rec, tt.wantStatus)
		})
	}
}

func TestClientCreateReturnsID(t *testing.T) {
	serv := &fakeClientServ{createID: 77}
	h := NewClientHandler(serv)

	rec := serve(t, "POST /clients", h.CreateClient, http.MethodPost, "/clients", `{"company_id":1,"name":"Ivan","phone":"+79990000000"}`)

	assertStatus(t, rec, http.StatusCreated)
	var body map[string]int64
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body["id"] != 77 {
		t.Errorf("id = %d, want 77", body["id"])
	}
}

func TestClientGetAll(t *testing.T) {
	tests := []struct {
		name       string
		serv       *fakeClientServ
		wantStatus int
		wantLen    int
	}{
		{"two items", &fakeClientServ{listItems: []models.Client{{ID: 1}, {ID: 2}}}, http.StatusOK, 2},
		{"empty", &fakeClientServ{}, http.StatusOK, 0},
		{"storage failure", &fakeClientServ{listErr: errors.New("boom")}, http.StatusInternalServerError, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewClientHandler(tt.serv)

			rec := serve(t, "GET /clients", h.GetAllClients, http.MethodGet, "/clients", "")

			assertStatus(t, rec, tt.wantStatus)
			if tt.wantStatus != http.StatusOK {
				return
			}
			var items []models.Client
			if err := json.NewDecoder(rec.Body).Decode(&items); err != nil {
				t.Fatalf("decode: %v", err)
			}
			if len(items) != tt.wantLen {
				t.Errorf("got %d items, want %d", len(items), tt.wantLen)
			}
		})
	}
}

func TestClientGetAllFiltersByCompany(t *testing.T) {
	serv := &fakeClientServ{listItems: []models.Client{{ID: 1}}}
	h := NewClientHandler(serv)

	rec := serve(t, "GET /clients", h.GetAllClients, http.MethodGet, "/clients?company_id=4", "")

	assertStatus(t, rec, http.StatusOK)
	if serv.companyID != 4 {
		t.Errorf("companyID = %d, want 4", serv.companyID)
	}
}

func TestClientGetAllRejectsBadCompanyID(t *testing.T) {
	h := NewClientHandler(&fakeClientServ{})

	rec := serve(t, "GET /clients", h.GetAllClients, http.MethodGet, "/clients?company_id=abc", "")

	assertStatus(t, rec, http.StatusBadRequest)
}

func TestClientGetById(t *testing.T) {
	tests := []struct {
		name       string
		target     string
		serv       *fakeClientServ
		wantStatus int
	}{
		{"found", "/clients/3", &fakeClientServ{item: models.Client{ID: 3}}, http.StatusOK},
		{"not found", "/clients/3", &fakeClientServ{getErr: models.ErrNotFound}, http.StatusNotFound},
		{"bad id", "/clients/abc", &fakeClientServ{}, http.StatusBadRequest},
		{"storage failure", "/clients/3", &fakeClientServ{getErr: errors.New("boom")}, http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewClientHandler(tt.serv)

			rec := serve(t, "GET /clients/{id}", h.GetClientById, http.MethodGet, tt.target, "")

			assertStatus(t, rec, tt.wantStatus)
		})
	}
}

func TestClientDelete(t *testing.T) {
	tests := []struct {
		name       string
		target     string
		serv       *fakeClientServ
		wantStatus int
	}{
		{"deleted", "/clients/3", &fakeClientServ{}, http.StatusNoContent},
		{"not found", "/clients/3", &fakeClientServ{deleteErr: models.ErrNotFound}, http.StatusNotFound},
		{"bad id", "/clients/x", &fakeClientServ{}, http.StatusBadRequest},
		{"storage failure", "/clients/3", &fakeClientServ{deleteErr: errors.New("boom")}, http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewClientHandler(tt.serv)

			rec := serve(t, "DELETE /clients/{id}", h.DeleteClient, http.MethodDelete, tt.target, "")

			assertStatus(t, rec, tt.wantStatus)
		})
	}
}
