package app

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/artemydottech/goclients/internal/models"
)

func TestAppointmentHistoryHTTPAndIdempotentCreation(t *testing.T) {
	s := newSalon(t)
	_, id, _ := keyedBooking(t, s, "history-key", "")
	keyedBooking(t, s, "history-key", "")
	path := "/appointments/" + itoa(id) + "/history"
	raw := s.api.expect(http.MethodGet, path, nil, 200)
	var events []models.AppointmentEvent
	if err := json.Unmarshal(raw, &events); err != nil || len(events) != 1 || events[0].Kind != "created" {
		t.Fatalf("duplicate create: %s %v", raw, err)
	}
	s.api.expect(http.MethodPut, "/appointments/"+itoa(id)+"/comment", map[string]any{"comment": "New note"}, 204)
	raw = s.api.expect(http.MethodGet, path+"?limit=1", nil, 200)
	if err := json.Unmarshal(raw, &events); err != nil || len(events) != 1 || events[0].Kind != "comment" || events[0].Before == nil || events[0].After.Comment != "New note" {
		t.Fatalf("note event: %s %v", raw, err)
	}
	s.api.expect(http.MethodGet, path+"?limit=201", nil, 400)
	s.api.expect(http.MethodGet, "/appointments/999999/history", nil, 404)
}
