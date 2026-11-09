package app

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"
)

func keyedBooking(t *testing.T, s salon, key string, comment string) (int, int, string) {
	t.Helper()
	raw, err := json.Marshal(map[string]any{"client_id": s.clientID, "employee_id": s.employeeID, "service_id": s.serviceID, "starts_at": s.at(10, 0).Format(time.RFC3339), "comment": comment})
	if err != nil {
		t.Fatal(err)
	}
	req, err := http.NewRequest(http.MethodPost, s.api.base+"/appointments", bytes.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Idempotency-Key", key)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	id := 0
	if resp.StatusCode == http.StatusCreated {
		id = decodeID(t, body)
	}
	return resp.StatusCode, id, resp.Header.Get("Idempotent-Replayed")
}
func TestIdempotentBookingHTTP(t *testing.T) {
	s := newSalon(t)
	code, id, replay := keyedBooking(t, s, "first-request", "")
	if code != 201 || id == 0 || replay != "" {
		t.Fatalf("create: %d %d %q", code, id, replay)
	}
	code, retry, replay := keyedBooking(t, s, "first-request", "")
	if code != 201 || retry != id || replay != "true" {
		t.Fatalf("retry: %d %d %q", code, retry, replay)
	}
	code, _, _ = keyedBooking(t, s, "first-request", "changed")
	if code != 409 {
		t.Fatalf("conflict: %d", code)
	}
	for _, key := range []string{"", "with space"} {
		code, _, _ = keyedBooking(t, s, key, "")
		if code != 400 {
			t.Fatalf("key %q: %d", key, code)
		}
	}
}
