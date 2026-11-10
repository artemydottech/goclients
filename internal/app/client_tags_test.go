package app

import (
	"net/http"
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
