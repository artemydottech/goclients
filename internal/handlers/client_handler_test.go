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
	gotPage   models.Page
	updatedID int
	updated   models.Client
	updateErr error
	gotCtx    context.Context
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

func (f *fakeClientServ) CreateClient(ctx context.Context, v models.Client) (int64, error) {
	f.gotCtx = ctx
	f.created = v
	return f.createID, f.createErr
}

func (f *fakeClientServ) GetAllClients(_ context.Context, page models.Page) ([]models.Client, error) {
	f.gotPage = page
	return f.listItems, f.listErr
}

func (f *fakeClientServ) GetClientsByCompany(_ context.Context, companyID int, page models.Page) ([]models.Client, error) {
	f.gotPage = page
	f.companyID = companyID
	return f.listItems, f.listErr
}

func (f *fakeClientServ) GetClientById(_ context.Context, id int) (models.Client, error) {
	return f.item, f.getErr
}

func (f *fakeClientServ) UpdateClient(_ context.Context, id int, v models.Client) error {
	f.updatedID = id
	f.updated = v
	return f.updateErr
}

func (f *fakeClientServ) DeleteClientById(_ context.Context, id int) error {
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

func TestCreateClientPassesRequestContext(t *testing.T) {
	serv := &fakeClientServ{}
	h := NewClientHandler(serv)

	serveWithContext(t, markedContext(), "POST /clients", h.CreateClient, http.MethodPost, "/clients", `{"name":"Ivan","phone":"+79990000000"}`)

	assertMarked(t, serv.gotCtx)
}

func TestUpdateClient(t *testing.T) {
	body := `{"name":"Ivan","phone":"+79990000000"}`
	tests := []struct {
		name       string
		target     string
		body       string
		serv       *fakeClientServ
		wantStatus int
	}{
		{"updated", "/clients/3", body, &fakeClientServ{}, http.StatusNoContent},
		{"bad id", "/clients/x", body, &fakeClientServ{}, http.StatusBadRequest},
		{"malformed body", "/clients/3", `{oops`, &fakeClientServ{}, http.StatusBadRequest},
		{"validation error", "/clients/3", body, &fakeClientServ{updateErr: models.Invalid("client with phone already exists")}, http.StatusBadRequest},
		{"not found", "/clients/3", body, &fakeClientServ{updateErr: models.ErrNotFound}, http.StatusNotFound},
		{"storage failure", "/clients/3", body, &fakeClientServ{updateErr: errors.New("boom")}, http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewClientHandler(tt.serv)

			rec := serve(t, "PUT /clients/{id}", h.UpdateClient, http.MethodPut, tt.target, tt.body)

			assertStatus(t, rec, tt.wantStatus)
		})
	}
}

func TestUpdateClientPassesIDAndBody(t *testing.T) {
	serv := &fakeClientServ{}
	h := NewClientHandler(serv)

	serve(t, "PUT /clients/{id}", h.UpdateClient, http.MethodPut, "/clients/9", `{"name":"Ivan","phone":"+79990000000","comment":"regular"}`)

	if serv.updatedID != 9 || serv.updated.Comment != "regular" {
		t.Errorf("service got id=%d client=%+v", serv.updatedID, serv.updated)
	}
}

func TestGetAllClientsPassesThePage(t *testing.T) {
	serv := &fakeClientServ{}
	h := NewClientHandler(serv)

	rec := serve(t, "GET /clients", h.GetAllClients, http.MethodGet, "/clients?limit=25&offset=50", "")

	assertStatus(t, rec, http.StatusOK)
	if serv.gotPage != (models.Page{Limit: 25, Offset: 50}) {
		t.Errorf("service got page %+v", serv.gotPage)
	}
}

func TestGetAllClientsPassesThePageWithTheCompanyFilter(t *testing.T) {
	serv := &fakeClientServ{}
	h := NewClientHandler(serv)

	rec := serve(t, "GET /clients", h.GetAllClients, http.MethodGet, "/clients?company_id=7&offset=4", "")

	assertStatus(t, rec, http.StatusOK)
	if serv.companyID != 7 || serv.gotPage != (models.Page{Offset: 4}) {
		t.Errorf("service got company=%d page=%+v", serv.companyID, serv.gotPage)
	}
}

func TestGetAllClientsRejectsABadPage(t *testing.T) {
	h := NewClientHandler(&fakeClientServ{})

	rec := serve(t, "GET /clients", h.GetAllClients, http.MethodGet, "/clients?limit=500", "")

	assertStatus(t, rec, http.StatusBadRequest)
}
