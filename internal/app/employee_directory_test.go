package app

import (
	"encoding/json"
	"github.com/artemydottech/goclients/internal/models"
	"net/http"
	"testing"
)

func TestCompanyEmployeeDirectory(t *testing.T) {
	api := newClient(t)
	a := newCompany(t, api)
	b := newCompany(t, api)
	for _, id := range []int{a, b, a} {
		api.create("/employees", map[string]any{"company_id": id, "name": "Anna", "surname": "Ivanova"})
	}
	raw := api.expect(http.MethodGet, "/employees?company_id="+itoa(a)+"&limit=1&offset=1", nil, http.StatusOK)
	var got []models.Employee
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].CompanyID != a {
		t.Fatalf("directory: %+v", got)
	}
	raw = api.expect(http.MethodGet, "/employees?company_id=99999", nil, http.StatusOK)
	if string(raw) != "[]\n" {
		t.Fatalf("empty directory: %s", raw)
	}
}
