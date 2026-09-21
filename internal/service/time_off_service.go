package service

import (
	"time"
	"unicode/utf8"

	"github.com/artemydottech/goclients/internal/models"
)

type TimeOffRepo interface {
	Create(t models.TimeOff) (int64, error)
	GetTimeOffByEmployee(employeeID int) ([]models.TimeOff, error)
	GetTimeOffInRange(employeeID int, from, to time.Time) ([]models.TimeOff, error)
	DeleteTimeOffById(id int) error
}

type TimeOffCalendar interface {
	GetTimeOffInRange(employeeID int, from, to time.Time) ([]models.TimeOff, error)
}

type TimeOffService struct {
	repo      TimeOffRepo
	employees EmployeeLookup
	calendar  Calendar
}

func NewTimeOffService(repo TimeOffRepo, employees EmployeeLookup, calendar Calendar) *TimeOffService {
	return &TimeOffService{repo: repo, employees: employees, calendar: calendar}
}

func (s *TimeOffService) CreateTimeOff(employeeID int, t models.TimeOff) (int64, error) {
	if _, err := s.employees.GetEmployeeById(employeeID); err != nil {
		return 0, err
	}

	if t.StartsAt.IsZero() || t.EndsAt.IsZero() {
		return 0, models.Invalid("period start and end are required")
	}

	if !t.EndsAt.After(t.StartsAt) {
		return 0, models.Invalid("period end must be after start")
	}

	if utf8.RuneCountInString(t.Reason) > 200 {
		return 0, models.Invalid("reason is too long, keep it under 200 characters")
	}

	booked, err := s.calendar.GetAppointmentsByEmployee(employeeID, t.StartsAt, t.EndsAt)
	if err != nil {
		return 0, err
	}

	active := 0
	for _, appointment := range booked {
		if appointment.Status.Active() {
			active++
		}
	}
	if active > 0 {
		return 0, models.Invalid("the employee has %d active appointments in this period, reschedule or cancel them first", active)
	}

	t.EmployeeID = employeeID
	t.StartsAt = t.StartsAt.UTC()
	t.EndsAt = t.EndsAt.UTC()

	return s.repo.Create(t)
}

func (s *TimeOffService) GetTimeOffByEmployee(employeeID int) ([]models.TimeOff, error) {
	return s.repo.GetTimeOffByEmployee(employeeID)
}

func (s *TimeOffService) DeleteTimeOffById(id int) error {
	return s.repo.DeleteTimeOffById(id)
}

func overlapsTimeOff(from, to time.Time, periods []models.TimeOff) bool {
	for _, period := range periods {
		if from.Before(period.EndsAt) && to.After(period.StartsAt) {
			return true
		}
	}

	return false
}
