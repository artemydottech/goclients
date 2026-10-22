package handlers

import (
	"net/http"
	"testing"

	"github.com/artemydottech/goclients/internal/models"
)

func TestParsePage(t *testing.T) {
	tests := []struct {
		name       string
		target     string
		want       models.Page
		wantStatus int
	}{
		{"no parameters", "/items", models.Page{}, http.StatusOK},
		{"limit", "/items?limit=20", models.Page{Limit: 20}, http.StatusOK},
		{"limit and offset", "/items?limit=20&offset=40", models.Page{Limit: 20, Offset: 40}, http.StatusOK},
		{"offset only", "/items?offset=7", models.Page{Offset: 7}, http.StatusOK},
		{"limit is not a number", "/items?limit=many", models.Page{}, http.StatusBadRequest},
		{"offset is not a number", "/items?offset=far", models.Page{}, http.StatusBadRequest},
		{"negative limit", "/items?limit=-1", models.Page{}, http.StatusBadRequest},
		{"negative offset", "/items?offset=-1", models.Page{}, http.StatusBadRequest},
		{"limit above the maximum", "/items?limit=201", models.Page{}, http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got models.Page
			var ok bool

			rec := serve(t, "GET /items", func(w http.ResponseWriter, r *http.Request) {
				got, ok = parsePage(w, r)
			}, http.MethodGet, tt.target, "")

			if wantOK := tt.wantStatus == http.StatusOK; ok != wantOK {
				t.Fatalf("ok = %v, want %v", ok, wantOK)
			}
			if tt.wantStatus != http.StatusOK {
				assertStatus(t, rec, tt.wantStatus)
				return
			}
			if got != tt.want {
				t.Errorf("page = %+v, want %+v", got, tt.want)
			}
		})
	}
}
