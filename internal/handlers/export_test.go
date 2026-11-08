package handlers

import (
	"context"
	"errors"
	"github.com/artemydottech/goclients/internal/models"
	"net/http"
	"testing"
)

type exportClientStub struct {
	err error
	ctx context.Context
}

func (s *exportClientStub) SearchClients(ctx context.Context, f models.ClientSearch) ([]models.Client, error) {
	s.ctx = ctx
	return []models.Client{}, s.err
}
func TestExportErrorsRemainJSON(t *testing.T) {
	for _, tt := range []struct {
		err  error
		code int
	}{{models.Invalid("bad search"), 400}, {errors.New("storage"), 500}} {
		h := NewClientSearchHandler(&exportClientStub{err: tt.err})
		rec := serve(t, "GET /clients/export", h.Export, http.MethodGet, "/clients/export?company_id=1", "")
		assertStatus(t, rec, tt.code)
		if rec.Header().Get("Content-Disposition") != "" || rec.Header().Get("Content-Type") != "application/json" {
			t.Fatal(rec.Header())
		}
	}
}
