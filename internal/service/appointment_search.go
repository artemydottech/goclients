package service

import (
	"context"

	"github.com/artemydottech/goclients/internal/models"
)

type AppointmentSearchRepo interface {
	SearchAppointments(context.Context, models.AppointmentFilter) ([]models.Appointment, error)
}
type AppointmentSearchService struct{ repo AppointmentSearchRepo }

func NewAppointmentSearchService(r AppointmentSearchRepo) *AppointmentSearchService {
	return &AppointmentSearchService{repo: r}
}
func (s *AppointmentSearchService) SearchAppointments(ctx context.Context, f models.AppointmentFilter) ([]models.Appointment, error) {
	if err := f.Validate(); err != nil {
		return nil, err
	}
	return s.repo.SearchAppointments(ctx, f)
}
