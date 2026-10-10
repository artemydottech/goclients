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
