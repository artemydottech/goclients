package app

import (
	"encoding/json"
	"github.com/artemydottech/goclients/internal/models"
	"net/http"
	"testing"
)

func TestCompanyReportHTTP(t *testing.T) {
	api := newClient(t)
	id := newCompany(t, api)
	period := "?from=2026-11-01T00:00:00Z&to=2026-12-01T00:00:00Z"
	raw := api.expect(http.MethodGet, "/companies/"+itoa(id)+"/report"+period, nil, http.StatusOK)
	var got models.CompanyReport
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatal(err)
	}
	if got.CompanyID != id || got.Summary.Appointments != 0 || got.Services == nil || got.Employees == nil {
		t.Fatalf("empty: %s", raw)
	}
	api.expect(http.MethodGet, "/companies/99999/report"+period, nil, http.StatusNotFound)
	for _, q := range []string{"", "?from=invalid", "?from=2026-11-01T00:00:00Z&to=2026-10-01T00:00:00Z", "?from=2020-01-01T00:00:00Z&to=2026-01-01T00:00:00Z"} {
		api.expect(http.MethodGet, "/companies/"+itoa(id)+"/report"+q, nil, http.StatusBadRequest)
	}
}
