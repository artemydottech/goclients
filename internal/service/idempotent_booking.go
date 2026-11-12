package service

import (
	"context"
	"errors"

	"github.com/artemydottech/goclients/internal/models"
)

type BookingRequestRepo interface {
	FindBookingRequest(context.Context, int, string) (models.BookingRequest, error)
	ReserveBooking(context.Context, string, string, models.Appointment) (models.BookingResult, error)
}

type IdempotentBookingService struct {
	bookings *AppointmentService
	repo     BookingRequestRepo
}

func NewIdempotentBookingService(b *AppointmentService, r BookingRequestRepo) *IdempotentBookingService {
	return &IdempotentBookingService{bookings: b, repo: r}
}

func (s *IdempotentBookingService) Book(ctx context.Context, key string, a models.Appointment) (models.BookingResult, error) {
	if key == "" {
		id, err := s.bookings.Book(ctx, a)
		return models.BookingResult{ID: id}, err
	}
	if err := models.ValidateBookingKey(key); err != nil {
		return models.BookingResult{}, err
	}
	if a.ClientID <= 0 {
		return models.BookingResult{}, models.Invalid("client_id must be positive")
	}
	hash, err := models.BookingFingerprint(a)
	if err != nil {
		return models.BookingResult{}, models.Invalid("invalid booking request")
	}
	previous, err := s.repo.FindBookingRequest(ctx, a.ClientID, key)
	if err == nil {
		if previous.Fingerprint != hash {
			return models.BookingResult{}, models.ErrConflict
		}
		return models.BookingResult{ID: previous.AppointmentID, Replayed: true}, nil
	}
	if !errors.Is(err, models.ErrNotFound) {
		return models.BookingResult{}, err
	}
	prepared, err := s.bookings.prepare(ctx, a)
	if err != nil {
		return models.BookingResult{}, err
	}
	return s.repo.ReserveBooking(ctx, key, hash, prepared)
}
