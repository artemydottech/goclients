package service

import (
	"context"
	"strings"
	"testing"
)

type commentStub struct {
	called  bool
	comment string
}

func (s *commentStub) UpdateComment(ctx context.Context, id int, comment string) error {
	s.called = true
	s.comment = comment
	return ctx.Err()
}
func TestAppointmentCommentValidation(t *testing.T) {
	for _, tt := range []struct {
		id      int
		comment string
	}{{0, "ok"}, {-1, "ok"}, {1, strings.Repeat("я", 1001)}} {
		repo := &commentStub{}
		err := NewAppointmentCommentService(repo).UpdateComment(context.Background(), tt.id, tt.comment)
		if err == nil || repo.called {
			t.Fatalf("invalid note reached storage: %+v", tt)
		}
	}
	repo := &commentStub{}
	comment := strings.Repeat("я", 1000)
	if err := NewAppointmentCommentService(repo).UpdateComment(context.Background(), 1, comment); err != nil || repo.comment != comment {
		t.Fatalf("valid: %v", err)
	}
}
