package service

import (
	"context"
	"unicode/utf8"

	"github.com/artemydottech/goclients/internal/models"
)

type AppointmentCommentRepo interface {
	UpdateComment(context.Context, int, string) error
}
type AppointmentCommentService struct{ repo AppointmentCommentRepo }

func NewAppointmentCommentService(r AppointmentCommentRepo) *AppointmentCommentService {
	return &AppointmentCommentService{repo: r}
}
func (s *AppointmentCommentService) UpdateComment(ctx context.Context, id int, comment string) error {
	if id <= 0 {
		return models.Invalid("invalid appointment id")
	}
	if utf8.RuneCountInString(comment) > 1000 {
		return models.Invalid("comment cannot exceed 1000 characters")
	}
	return s.repo.UpdateComment(ctx, id, comment)
}
