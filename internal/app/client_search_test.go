package app

import (
	"encoding/json"
	"github.com/artemydottech/goclients/internal/models"
	"net/http"
	"net/url"
	"testing"
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
