package service

import (
	"context"

	"github.com/artemydottech/goclients/internal/models"
)

type AppointmentHistoryRepo interface {
	GetAppointmentHistory(context.Context, int, models.Page) ([]models.AppointmentEvent, error)
}

type HistoryAppointmentLookup interface {
	GetAppointmentById(context.Context, int) (models.Appointment, error)
}

type AppointmentHistoryService struct {
	repo         AppointmentHistoryRepo
	appointments HistoryAppointmentLookup
}

func NewAppointmentHistoryService(r AppointmentHistoryRepo, a HistoryAppointmentLookup) *AppointmentHistoryService {
	return &AppointmentHistoryService{repo: r, appointments: a}
}

func (s *AppointmentHistoryService) List(ctx context.Context, id int, page models.Page) ([]models.AppointmentEvent, error) {
	if id <= 0 {
		return nil, models.Invalid("invalid appointment id")
	}
	if err := page.Validate(); err != nil {
		return nil, err
	}
	if _, err := s.appointments.GetAppointmentById(ctx, id); err != nil {
		return nil, err
	}
	return s.repo.GetAppointmentHistory(ctx, id, page)
}
