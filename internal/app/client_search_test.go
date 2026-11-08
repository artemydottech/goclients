package app

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/artemydottech/goclients/internal/models"
)

func TestClientSearchHTTP(t *testing.T) {
	s := newSalon(t)
	raw := s.api.expect(http.MethodGet, "/clients/search?company_id="+itoa(s.companyID)+"&phone="+url.QueryEscape("+7 (999) 000-00-00"), nil, http.StatusOK)
	var got []models.Client
	if err := json.Unmarshal(raw, &got); err != nil || len(got) != 1 || got[0].ID != s.clientID {
		t.Fatalf("search: %s %v", raw, err)
	}
	other := newCompany(t, s.api)
	raw = s.api.expect(http.MethodGet, "/clients/search?company_id="+itoa(other)+"&q=Ivan", nil, http.StatusOK)
	if string(raw) != "[]\n" {
		t.Fatalf("foreign result: %s", raw)
	}
}
func TestClientSearchRejectsInvalidRequests(t *testing.T) {
	api := newClient(t)
	for _, q := range []string{"", "company_id=0", "company_id=no", "company_id=1&phone=abc", "company_id=1&limit=201", "company_id=1&offset=-1", "company_id=1&q=" + strings.Repeat("x", 201)} {
		api.expect(http.MethodGet, "/clients/search?"+q, nil, http.StatusBadRequest)
	}
}
