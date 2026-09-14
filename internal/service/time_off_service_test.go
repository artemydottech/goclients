package service

import (
	"database/sql"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/artemydottech/goclients/internal/models"
)

type stubTimeOffRepo struct {
	saved models.TimeOff
	calls int
}

func (r *stubTimeOffRepo) Create(t models.TimeOff) (int64, error) {
	r.calls++
	r.saved = t
	return 4, nil
}

func (r *stubTimeOffRepo) GetTimeOffByEmployee(int) ([]models.TimeOff, error) { return nil, nil }

func (r *stubTimeOffRepo) GetTimeOffInRange(int, time.Time, time.Time) ([]models.TimeOff, error) {
	return nil, nil
}

func (r *stubTimeOffRepo) DeleteTimeOffById(int) error { return nil }

func vacation() models.TimeOff {
	return models.TimeOff{
		StartsAt: slotsDate,
		EndsAt:   slotsDate.AddDate(0, 0, 7),
		Reason:   "отпуск",
	}
}

func newTimeOffService(repo *stubTimeOffRepo, booked []models.Appointment) *TimeOffService {
	return NewTimeOffService(
		repo,
		stubEmployeeLookup{employee: models.Employee{ID: 2, CompanyID: 10}},
		stubCalendar{booked: booked},
	)
}

func TestCreateTimeOffStampsTheEmployee(t *testing.T) {
	repo := &stubTimeOffRepo{}

	if _, err := newTimeOffService(repo, nil).CreateTimeOff(2, vacation()); err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}
	if repo.saved.EmployeeID != 2 {
		t.Errorf("employee_id %d, ожидался 2", repo.saved.EmployeeID)
	}
}

func TestCreateTimeOffValidation(t *testing.T) {
	cases := map[string]models.TimeOff{
		"без начала":          {EndsAt: slotsDate},
		"без конца":           {StartsAt: slotsDate},
		"конец раньше начала": {StartsAt: slotsDate, EndsAt: slotsDate.Add(-time.Hour)},
		"нулевая длина":       {StartsAt: slotsDate, EndsAt: slotsDate},
		"длинная причина": {
			StartsAt: slotsDate, EndsAt: slotsDate.Add(time.Hour),
			Reason: strings.Repeat("а", 201),
		},
	}

	for name, period := range cases {
		t.Run(name, func(t *testing.T) {
			repo := &stubTimeOffRepo{}

			_, err := newTimeOffService(repo, nil).CreateTimeOff(2, period)

			var validationErr models.ValidationError
			if !errors.As(err, &validationErr) {
				t.Fatalf("ожидалась ValidationError, получено %v", err)
			}
			if repo.calls != 0 {
				t.Error("невалидный период не должен сохраняться")
			}
		})
	}
}

func TestCreateTimeOffRefusesToStrandActiveAppointments(t *testing.T) {
	repo := &stubTimeOffRepo{}
	booked := []models.Appointment{{
		StartsAt: slotsDate.Add(12 * time.Hour),
		EndsAt:   slotsDate.Add(13 * time.Hour),
		Status:   models.AppointmentConfirmed,
	}}

	_, err := newTimeOffService(repo, booked).CreateTimeOff(2, vacation())

	var validationErr models.ValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("ожидалась ValidationError, получено %v", err)
	}
	if repo.calls != 0 {
		t.Error("отпуск поверх активных записей не должен сохраняться")
	}
}

func TestCreateTimeOffIgnoresCancelledAndCompletedAppointments(t *testing.T) {
	repo := &stubTimeOffRepo{}
	booked := []models.Appointment{
		{StartsAt: slotsDate.Add(12 * time.Hour), EndsAt: slotsDate.Add(13 * time.Hour), Status: models.AppointmentCancelled},
		{StartsAt: slotsDate.Add(14 * time.Hour), EndsAt: slotsDate.Add(15 * time.Hour), Status: models.AppointmentCompleted},
	}

	if _, err := newTimeOffService(repo, booked).CreateTimeOff(2, vacation()); err != nil {
		t.Fatalf("отменённые и завершённые записи не мешают отпуску, получено %v", err)
	}
}

func TestCreateTimeOffReportsMissingEmployee(t *testing.T) {
	repo := &stubTimeOffRepo{}
	svc := NewTimeOffService(repo, stubEmployeeLookup{err: sql.ErrNoRows}, stubCalendar{})

	if _, err := svc.CreateTimeOff(2, vacation()); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("ожидалась sql.ErrNoRows, получено %v", err)
	}
}

func TestCreateTimeOffIgnoresNoShows(t *testing.T) {
	repo := &stubTimeOffRepo{}
	booked := []models.Appointment{{
		StartsAt: slotsDate.Add(12 * time.Hour),
		EndsAt:   slotsDate.Add(13 * time.Hour),
		Status:   models.AppointmentNoShow,
	}}

	if _, err := newTimeOffService(repo, booked).CreateTimeOff(2, vacation()); err != nil {
		t.Fatalf("неявка не мешает отпуску, получено %v", err)
	}
}
