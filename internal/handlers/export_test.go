package handlers

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/artemydottech/goclients/internal/models"
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
func TestEmptyCSVAttachmentAndContext(t *testing.T) {
	stub := &exportClientStub{}
	h := NewClientSearchHandler(stub)
	rec := serveWithContext(t, markedContext(), "GET /clients/export", h.Export, http.MethodGet, "/clients/export?company_id=1", "")
	assertStatus(t, rec, 200)
	assertMarked(t, stub.ctx)
	if rec.Header().Get("Content-Type") != "text/csv; charset=utf-8" || rec.Header().Get("Content-Disposition") != `attachment; filename="clients.csv"` || rec.Header().Get("Cache-Control") != "no-store" {
		t.Fatal(rec.Header())
	}
	if rec.Body.String() != "id,company_id,name,phone,email,comment\n" {
		t.Fatal(rec.Body.String())
	}
}
