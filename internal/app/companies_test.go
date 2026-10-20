package app

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestCompanyLifecycle(t *testing.T) {
	api := newClient(t)

	id := api.create("/companies", map[string]any{"name": "Romashka", "timezone": "Asia/Yekaterinburg"})

	raw := api.expect(http.MethodGet, "/companies/"+itoa(id), nil, http.StatusOK)
	var got struct {
		Name     string `json:"name"`
		Timezone string `json:"timezone"`
	}
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.Name != "Romashka" || got.Timezone != "Asia/Yekaterinburg" {
		t.Errorf("got %+v", got)
	}

	api.expect(http.MethodPut, "/companies/"+itoa(id), map[string]any{"name": "Lilia"}, http.StatusNoContent)

	raw = api.expect(http.MethodGet, "/companies/"+itoa(id), nil, http.StatusOK)
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.Name != "Lilia" || got.Timezone != "UTC" {
		t.Errorf("after update got %+v", got)
	}

	api.expect(http.MethodDelete, "/companies/"+itoa(id), nil, http.StatusNoContent)
	api.expect(http.MethodGet, "/companies/"+itoa(id), nil, http.StatusNotFound)
}

func TestCompanyValidationAndMissingRecords(t *testing.T) {
	api := newClient(t)

	raw := api.expect(http.MethodPost, "/companies", map[string]any{"name": ""}, http.StatusBadRequest)
	var body struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(raw, &body); err != nil || body.Error == "" {
		t.Errorf("expected a JSON error body, got %s", raw)
	}

	api.expect(http.MethodPut, "/companies/404", map[string]any{"name": "Ghost"}, http.StatusNotFound)
	api.expect(http.MethodPut, "/companies/abc", map[string]any{"name": "Ghost"}, http.StatusBadRequest)
	api.expect(http.MethodDelete, "/companies/404", nil, http.StatusNotFound)
}
