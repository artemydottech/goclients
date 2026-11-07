package app

import (
	"encoding/json"
	"github.com/artemydottech/goclients/internal/models"
	"net/http"
	"strings"
	"testing"
)

func TestAppointmentCommentHTTP(t *testing.T) {
	s := newSalon(t)
	id := s.mustBook(s.at(10, 0))
	path := "/appointments/" + itoa(id) + "/comment"
	for _, comment := range []string{"Пожелания клиента", ""} {
		s.api.expect(http.MethodPut, path, map[string]any{"comment": comment}, http.StatusNoContent)
		raw := s.api.expect(http.MethodGet, "/appointments/"+itoa(id), nil, http.StatusOK)
		var got models.Appointment
		if err := json.Unmarshal(raw, &got); err != nil || got.Comment != comment || got.Price != 1000 || got.Status != models.AppointmentPending {
			t.Fatalf("note: %s %v", raw, err)
		}
	}
	for _, body := range []any{map[string]any{}, map[string]any{"comment": nil}, map[string]any{"comment": strings.Repeat("я", 1001)}} {
		s.api.expect(http.MethodPut, path, body, http.StatusBadRequest)
	}
	s.api.expect(http.MethodPut, "/appointments/999999/comment", map[string]any{"comment": "x"}, http.StatusNotFound)
}
