package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWriteErrorUsesJSONEnvelope(t *testing.T) {
	rec := httptest.NewRecorder()

	writeError(rec, http.StatusNotFound, "user not found")

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
	if got := rec.Header().Get("Content-Type"); got != "application/json" {
		t.Errorf("content type = %q", got)
	}

	var body errorResponse
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body.Error != "user not found" {
		t.Errorf("error = %q", body.Error)
	}
}

func TestWriteStatusJSONEncodesValue(t *testing.T) {
	rec := httptest.NewRecorder()

	writeStatusJSON(rec, http.StatusCreated, map[string]int{"id": 7})

	if rec.Code != http.StatusCreated {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusCreated)
	}

	var body map[string]int
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body["id"] != 7 {
		t.Errorf("body = %v", body)
	}
}

func TestDecodeJSONRejectsMalformedBody(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("{oops"))
	var dst struct{ Name string }

	if decodeJSON(rec, req, &dst) {
		t.Fatal("decodeJSON accepted a malformed body")
	}
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rec.Code)
	}
}

func TestDecodeJSONFillsDestination(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"name":"Anna"}`))
	var dst struct{ Name string }

	if !decodeJSON(rec, req, &dst) {
		t.Fatalf("decodeJSON failed: %s", rec.Body.String())
	}
	if dst.Name != "Anna" {
		t.Errorf("Name = %q, want Anna", dst.Name)
	}
}

func TestDecodeJSONReportsTooLargeBody(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"name":"`+strings.Repeat("a", 100)+`"}`))
	req.Body = http.MaxBytesReader(rec, req.Body, 10)
	var dst struct{ Name string }

	if decodeJSON(rec, req, &dst) {
		t.Fatal("decodeJSON accepted an oversized body")
	}
	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("status = %d, want 413", rec.Code)
	}
}

func TestPathID(t *testing.T) {
	tests := []struct {
		name   string
		target string
		wantID int
		wantOK bool
	}{
		{"number", "/items/12", 12, true},
		{"zero", "/items/0", 0, true},
		{"letters", "/items/abc", 0, false},
		{"float", "/items/1.5", 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotID int
			var gotOK bool
			rec := serve(t, "GET /items/{id}", func(w http.ResponseWriter, r *http.Request) {
				gotID, gotOK = pathID(w, r)
			}, http.MethodGet, tt.target, "")

			if gotID != tt.wantID || gotOK != tt.wantOK {
				t.Fatalf("pathID = (%d, %v), want (%d, %v)", gotID, gotOK, tt.wantID, tt.wantOK)
			}
			if !tt.wantOK && rec.Code != http.StatusBadRequest {
				t.Errorf("status = %d, want 400", rec.Code)
			}
		})
	}
}
