package app

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestClientLifecycle(t *testing.T) {
	api := newClient(t)
	companyID := newCompany(t, api)

	id := api.create("/clients", map[string]any{"company_id": companyID, "name": "Ivan", "phone": "+7 (999) 000-00-00"})

	api.expect(http.MethodPut, "/clients/"+itoa(id), map[string]any{
		"name": "Ivan Petrov", "phone": "8 999 111 22 33", "comment": "regular",
	}, http.StatusNoContent)

	raw := api.expect(http.MethodGet, "/clients/"+itoa(id), nil, http.StatusOK)
	var got struct {
		CompanyID int    `json:"company_id"`
		Name      string `json:"name"`
		Phone     string `json:"phone"`
		Comment   string `json:"comment"`
	}
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.Name != "Ivan Petrov" || got.Phone != "89991112233" || got.Comment != "regular" {
		t.Errorf("got %+v", got)
	}
	if got.CompanyID != companyID {
		t.Errorf("company = %d, want %d", got.CompanyID, companyID)
	}

	api.expect(http.MethodDelete, "/clients/"+itoa(id), nil, http.StatusNoContent)
	api.expect(http.MethodGet, "/clients/"+itoa(id), nil, http.StatusNotFound)
}

func TestClientPhoneIsUniqueInsideACompany(t *testing.T) {
	api := newClient(t)
	first := newCompany(t, api)
	second := newCompany(t, api)

	api.create("/clients", map[string]any{"company_id": first, "name": "Anna", "phone": "79990000001"})
	other := api.create("/clients", map[string]any{"company_id": first, "name": "Ivan", "phone": "79990000002"})

	api.expect(http.MethodPost, "/clients", map[string]any{"company_id": first, "name": "Clone", "phone": "+7 999 000 00 01"}, http.StatusBadRequest)
	api.expect(http.MethodPut, "/clients/"+itoa(other), map[string]any{"name": "Ivan", "phone": "79990000001"}, http.StatusBadRequest)

	api.create("/clients", map[string]any{"company_id": second, "name": "Anna", "phone": "79990000001"})
}

func TestClientValidation(t *testing.T) {
	api := newClient(t)
	companyID := newCompany(t, api)

	api.expect(http.MethodPost, "/clients", map[string]any{"company_id": companyID, "name": "Ivan", "phone": "12"}, http.StatusBadRequest)
	api.expect(http.MethodPut, "/clients/404", map[string]any{"name": "Ivan", "phone": "79990000000"}, http.StatusNotFound)
}
