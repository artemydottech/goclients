package app

import (
	"encoding/json"
	"net/http"
	"testing"
)

func newCompany(t *testing.T, api client) int {
	t.Helper()

	return api.create("/companies", map[string]any{"name": "Romashka"})
}

func TestEmployeeLifecycle(t *testing.T) {
	api := newClient(t)
	companyID := newCompany(t, api)

	id := api.create("/employees", map[string]any{"company_id": companyID, "name": "Anna", "surname": "Ivanova"})

	api.expect(http.MethodPut, "/employees/"+itoa(id), map[string]any{
		"company_id": 999, "name": "Maria", "surname": "Petrova", "position": "Stylist",
	}, http.StatusNoContent)

	raw := api.expect(http.MethodGet, "/employees/"+itoa(id), nil, http.StatusOK)
	var got struct {
		CompanyID int    `json:"company_id"`
		Name      string `json:"name"`
		Position  string `json:"position"`
	}
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.Name != "Maria" || got.Position != "Stylist" {
		t.Errorf("got %+v", got)
	}
	if got.CompanyID != companyID {
		t.Errorf("company = %d, want %d: an update must not move the employee", got.CompanyID, companyID)
	}

	api.expect(http.MethodDelete, "/employees/"+itoa(id), nil, http.StatusNoContent)
	api.expect(http.MethodGet, "/employees/"+itoa(id), nil, http.StatusNotFound)
}

func TestEmployeesGoAwayWithTheirCompany(t *testing.T) {
	api := newClient(t)
	companyID := newCompany(t, api)
	id := api.create("/employees", map[string]any{"company_id": companyID, "name": "Anna", "surname": "Ivanova"})

	api.expect(http.MethodDelete, "/companies/"+itoa(companyID), nil, http.StatusNoContent)

	api.expect(http.MethodGet, "/employees/"+itoa(id), nil, http.StatusNotFound)
}

func TestEmployeeValidation(t *testing.T) {
	api := newClient(t)

	api.expect(http.MethodPost, "/employees", map[string]any{"name": "Anna"}, http.StatusBadRequest)
	api.expect(http.MethodPut, "/employees/404", map[string]any{"name": "Anna", "surname": "Ivanova"}, http.StatusNotFound)
}
