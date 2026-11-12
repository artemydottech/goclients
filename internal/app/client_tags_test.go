package app

import (
	"net/http"
	"net/url"
	"strings"
	"testing"
)

func TestClientTagsHTTP(t *testing.T) {
	s := newSalon(t)
	path := "/clients/" + itoa(s.clientID) + "/tags"
	s.api.expect(http.MethodPut, path, map[string]any{"tags": []string{" VIP ", "Новый", "vip"}}, 204)
	raw := s.api.expect(http.MethodGet, path, nil, 200)
	if string(raw) != "[\"vip\",\"новый\"]\n" {
		t.Fatalf("tags: %s", raw)
	}
	for _, body := range []any{map[string]any{}, map[string]any{"tags": nil}, map[string]any{"tags": []string{" "}}, map[string]any{"tags": []string{strings.Repeat("я", 33)}}} {
		s.api.expect(http.MethodPut, path, body, 400)
	}
	s.api.expect(http.MethodPut, path, map[string]any{"tags": []string{}}, 204)
	raw = s.api.expect(http.MethodGet, path, nil, 200)
	if string(raw) != "[]\n" {
		t.Fatal(string(raw))
	}
	s.api.expect(http.MethodGet, "/clients/999999/tags", nil, 404)
}

func TestClientTagSearchAndExportStayInCompany(t *testing.T) {
	s := newSalon(t)
	other := newCompany(t, s.api)
	path := "/clients/" + itoa(s.clientID) + "/tags"
	s.api.expect(http.MethodPut, path, map[string]any{"tags": []string{"Постоянный"}}, 204)
	q := url.Values{"company_id": {itoa(s.companyID)}, "tag": {" ПОСТОЯННЫЙ "}, "q": {"Ivan"}}
	raw := s.api.expect(http.MethodGet, "/clients/search?"+q.Encode(), nil, 200)
	if !strings.Contains(string(raw), "Ivan") {
		t.Fatal(string(raw))
	}
	q.Set("company_id", itoa(other))
	raw = s.api.expect(http.MethodGet, "/clients/search?"+q.Encode(), nil, 200)
	if string(raw) != "[]\n" {
		t.Fatal(string(raw))
	}
	q.Set("company_id", itoa(s.companyID))
	raw = s.api.expect(http.MethodGet, "/clients/export?"+q.Encode(), nil, 200)
	if !strings.Contains(string(raw), "Ivan") {
		t.Fatal(string(raw))
	}
}
