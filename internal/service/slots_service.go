package service

import (
	"database/sql"
	"errors"
	"time"

	"github.com/artemydottech/goclients/internal/models"
)

// defaultSlotStep — шаг сетки записи. Услуга на 50 минут всё равно
// предлагается с круглых значений, как в привычных календарях.
const defaultSlotStep = 15

type Schedule interface {
	WorkingDay(employeeID, weekday int) (models.WorkingDay, bool, error)
}

type Calendar interface {
	GetAppointmentsByEmployee(employeeID int, from, to time.Time) ([]models.Appointment, error)
}

type SlotsService struct {
	schedule    Schedule
	calendar    Calendar
	employees   EmployeeLookup
	services    ServiceLookup
	assignments Assignments
	now         func() time.Time
}

func NewSlotsService(
	schedule Schedule,
	calendar Calendar,
	employees EmployeeLookup,
	services ServiceLookup,
	assignments Assignments,
) *SlotsService {
	return &SlotsService{
		schedule:    schedule,
		calendar:    calendar,
		employees:   employees,
		services:    services,
		assignments: assignments,
		now:         time.Now,
	}
}

// FreeSlots перечисляет время, на которое можно записаться к мастеру в
// указанный день. Дата приходит без зоны, поэтому день считается в UTC.
func (s *SlotsService) FreeSlots(employeeID, serviceID int, date time.Time, step int) ([]time.Time, error) {
	if step <= 0 {
		step = defaultSlotStep
	}

	if _, err := s.employees.GetEmployeeById(employeeID); errors.Is(err, sql.ErrNoRows) {
		return nil, models.Invalid("Сотрудник %d не найден!", employeeID)
	} else if err != nil {
		return nil, err
	}

	item, err := s.services.GetServiceById(serviceID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, models.Invalid("Услуга %d не найдена!", serviceID)
	}
	if err != nil {
		return nil, err
	}

	performs, err := s.assignments.EmployeePerformsService(employeeID, serviceID)
	if err != nil {
		return nil, err
	}
	if !performs {
		return nil, models.Invalid("Сотрудник %d не оказывает услугу %d!", employeeID, serviceID)
	}

	day := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)

	working, isWorkingDay, err := s.schedule.WorkingDay(employeeID, int(day.Weekday()))
	if err != nil {
		return nil, err
	}
	if !isWorkingDay {
		return []time.Time{}, nil
	}

	startMinute, err := models.ParseDayTime(working.StartsAt)
	if err != nil {
		return nil, err
	}

	endMinute, err := models.ParseDayTime(working.EndsAt)
	if err != nil {
		return nil, err
	}

	booked, err := s.calendar.GetAppointmentsByEmployee(employeeID, day, day.AddDate(0, 0, 1))
	if err != nil {
		return nil, err
	}

	duration := time.Duration(item.Duration) * time.Minute
	now := s.now().UTC()
	slots := []time.Time{}

	for minute := startMinute; minute+item.Duration <= endMinute; minute += step {
		start := day.Add(time.Duration(minute) * time.Minute)
		if !start.After(now) {
			continue
		}

		if !overlapsAny(start, start.Add(duration), booked) {
			slots = append(slots, start)
		}
	}

	return slots, nil
}

func overlapsAny(from, to time.Time, booked []models.Appointment) bool {
	for _, appointment := range booked {
		if !appointment.Status.Blocks() {
			continue
		}

		if from.Before(appointment.EndsAt) && to.After(appointment.StartsAt) {
			return true
		}
	}

	return false
}
