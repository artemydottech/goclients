package app

import (
	"encoding/csv"
	"net/http"
	"strings"
	"testing"
)

func TestClientCSVExportIsolationAndEscaping(t *testing.T) {
	s := newSalon(t)
	s.api.create("/clients", map[string]any{"company_id": s.companyID, "name": "=HYPERLINK(\"bad\")", "phone": "79991111111", "comment": "line1\nline2"})
	q := "?company_id=" + itoa(s.companyID)
	raw := s.api.expect(http.MethodGet, "/clients/export"+q, nil, http.StatusOK)
	rows, err := csv.NewReader(strings.NewReader(string(raw))).ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 3 || rows[2][2] != "'=HYPERLINK(\"bad\")" || rows[2][5] != "line1\nline2" {
		t.Fatalf("csv: %s", raw)
	}
	raw = s.api.expect(http.MethodGet, "/clients/export?company_id="+itoa(newCompany(t, s.api)), nil, http.StatusOK)
	rows, err = csv.NewReader(strings.NewReader(string(raw))).ReadAll()
	if err != nil || len(rows) != 1 {
		t.Fatalf("foreign: %s %v", raw, err)
	}
}
