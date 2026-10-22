package app

import (
	"encoding/json"
	"net/http"
	"strconv"
	"testing"
)

func names(t *testing.T, raw []byte) []string {
	t.Helper()

	var items []struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(raw, &items); err != nil {
		t.Fatalf("decode: %v", err)
	}

	out := make([]string, 0, len(items))
	for _, item := range items {
		out = append(out, item.Name)
	}

	return out
}

func sameNames(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

func TestCompaniesArePaged(t *testing.T) {
	api := newClient(t)
	for i := 1; i <= 5; i++ {
		api.create("/companies", map[string]any{"name": "Company " + strconv.Itoa(i)})
	}

	all := names(t, api.expect(http.MethodGet, "/companies", nil, http.StatusOK))
	if len(all) != 5 {
		t.Fatalf("without a page every company is returned, got %v", all)
	}

	page := names(t, api.expect(http.MethodGet, "/companies?limit=2&offset=1", nil, http.StatusOK))
	if want := all[1:3]; !sameNames(page, want) {
		t.Errorf("page = %v, want %v", page, want)
	}

	tail := names(t, api.expect(http.MethodGet, "/companies?offset=4", nil, http.StatusOK))
	if want := all[4:]; !sameNames(tail, want) {
		t.Errorf("tail = %v, want %v", tail, want)
	}
}

func TestServicesArePagedInsideACompany(t *testing.T) {
	api := newClient(t)
	mine := newCompany(t, api)
	other := newCompany(t, api)
	for i := 1; i <= 4; i++ {
		api.create("/services", map[string]any{"company_id": mine, "name": "Mine " + strconv.Itoa(i), "duration": 30})
		api.create("/services", map[string]any{"company_id": other, "name": "Other " + strconv.Itoa(i), "duration": 30})
	}

	got := names(t, api.expect(http.MethodGet, "/services?company_id="+itoa(mine)+"&limit=2&offset=1", nil, http.StatusOK))

	if want := []string{"Mine 2", "Mine 3"}; !sameNames(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestBadPagingParametersAreRejected(t *testing.T) {
	api := newClient(t)

	for _, path := range []string{
		"/companies?limit=abc",
		"/companies?limit=-1",
		"/companies?limit=201",
		"/employees?offset=-1",
		"/services?limit=1000",
		"/clients?offset=x",
	} {
		api.expect(http.MethodGet, path, nil, http.StatusBadRequest)
	}
}
