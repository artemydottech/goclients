package app

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestServiceLifecycle(t *testing.T) {
	api := newClient(t)
	companyID := newCompany(t, api)

	id := api.create("/services", map[string]any{
		"company_id": companyID, "name": "Haircut", "duration": 30, "price": 1000,
	})

	api.expect(http.MethodPut, "/services/"+itoa(id), map[string]any{
		"name": "Long haircut", "duration": 60, "price": 1500,
	}, http.StatusNoContent)

	raw := api.expect(http.MethodGet, "/services/"+itoa(id), nil, http.StatusOK)
	var got struct {
		CompanyID int     `json:"company_id"`
		Name      string  `json:"name"`
		Duration  int     `json:"duration"`
		Price     float64 `json:"price"`
	}
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.Name != "Long haircut" || got.Duration != 60 || got.Price != 1500 {
		t.Errorf("got %+v", got)
	}
	if got.CompanyID != companyID {
		t.Errorf("company = %d, want %d", got.CompanyID, companyID)
	}

	api.expect(http.MethodDelete, "/services/"+itoa(id), nil, http.StatusNoContent)
	api.expect(http.MethodGet, "/services/"+itoa(id), nil, http.StatusNotFound)
}

func TestServicesAreFilteredByCompany(t *testing.T) {
	api := newClient(t)
	first := newCompany(t, api)
	second := newCompany(t, api)
	api.create("/services", map[string]any{"company_id": first, "name": "Haircut", "duration": 30})
	api.create("/services", map[string]any{"company_id": second, "name": "Manicure", "duration": 45})

	raw := api.expect(http.MethodGet, "/services?company_id="+itoa(first), nil, http.StatusOK)

	var got []struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(got) != 1 || got[0].Name != "Haircut" {
		t.Errorf("got %+v", got)
	}
}

func TestServiceValidation(t *testing.T) {
	api := newClient(t)
	companyID := newCompany(t, api)

	api.expect(http.MethodPost, "/services", map[string]any{"company_id": companyID, "name": "Haircut", "duration": 0}, http.StatusBadRequest)
	api.expect(http.MethodPost, "/services", map[string]any{"name": "Haircut", "duration": 30}, http.StatusBadRequest)
	api.expect(http.MethodPut, "/services/404", map[string]any{"name": "Haircut", "duration": 30}, http.StatusNotFound)
}
