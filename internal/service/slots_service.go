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
	companies   CompanyLookup
	timeOff     TimeOffCalendar
	now         func() time.Time
}

type CompanyLookup interface {
	GetCompanyById(id int) (models.Company, error)
}

func NewSlotsService(
	schedule Schedule,
	calendar Calendar,
	employees EmployeeLookup,
	services ServiceLookup,
	assignments Assignments,
	companies CompanyLookup,
	timeOff TimeOffCalendar,
) *SlotsService {
	return &SlotsService{
		schedule:    schedule,
		calendar:    calendar,
		employees:   employees,
		services:    services,
		assignments: assignments,
		companies:   companies,
		timeOff:     timeOff,
		now:         time.Now,
	}
}

// FreeSlots перечисляет время, на которое можно записаться к мастеру в
// указанный день. Дата приходит без зоны и понимается в поясе компании.
func (s *SlotsService) FreeSlots(employeeID, serviceID int, date time.Time, step int) ([]time.Time, error) {
	if step <= 0 {
		step = defaultSlotStep
	}

	employee, err := s.employees.GetEmployeeById(employeeID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, models.Invalid("Сотрудник %d не найден!", employeeID)
	}
	if err != nil {
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

	loc, err := companyLocation(s.companies, employee.CompanyID)
	if err != nil {
		return nil, err
	}

	day := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, loc)

	opens, closes, breaks, isWorkingDay, err := workingWindow(s.schedule, employeeID, day)
	if err != nil {
		return nil, err
	}
	if !isWorkingDay {
		return []time.Time{}, nil
	}

	booked, err := s.calendar.GetAppointmentsByEmployee(employeeID, day, day.AddDate(0, 0, 1))
	if err != nil {
		return nil, err
	}

	absences, err := s.timeOff.GetTimeOffInRange(employeeID, day, day.AddDate(0, 0, 1))
	if err != nil {
		return nil, err
	}
	absences = append(absences, breaks...)

	duration := time.Duration(item.Duration) * time.Minute
	now := s.now()
	slots := []time.Time{}

	stepDuration := time.Duration(step) * time.Minute

	for start := opens; !start.Add(duration).After(closes); start = start.Add(stepDuration) {
		if !start.After(now) {
			continue
		}

		end := start.Add(duration)
		if !overlapsAny(start, end, booked) && !overlapsTimeOff(start, end, absences) {
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

// workingWindow переводит рабочий день мастера из графика в моменты времени
// для конкретной даты. day — полночь нужного дня.
// Перерыв возвращается интервалом того же вида, что и отпуск: для слотов и
// записи это одно и то же — время, когда мастер не принимает.
func workingWindow(schedule Schedule, employeeID int, day time.Time) (time.Time, time.Time, []models.TimeOff, bool, error) {
	working, isWorkingDay, err := schedule.WorkingDay(employeeID, int(day.Weekday()))
	if err != nil || !isWorkingDay {
		return time.Time{}, time.Time{}, nil, false, err
	}

	atMinute := func(value string) (time.Time, error) {
		minute, err := models.ParseDayTime(value)
		if err != nil {
			return time.Time{}, err
		}
		return time.Date(day.Year(), day.Month(), day.Day(), 0, minute, 0, 0, day.Location()), nil
	}

	opens, err := atMinute(working.StartsAt)
	if err != nil {
		return time.Time{}, time.Time{}, nil, false, err
	}

	closes, err := atMinute(working.EndsAt)
	if err != nil {
		return time.Time{}, time.Time{}, nil, false, err
	}

	if !working.HasBreak() {
		return opens, closes, nil, true, nil
	}

	breakStarts, err := atMinute(working.BreakStartsAt)
	if err != nil {
		return time.Time{}, time.Time{}, nil, false, err
	}

	breakEnds, err := atMinute(working.BreakEndsAt)
	if err != nil {
		return time.Time{}, time.Time{}, nil, false, err
	}

	breaks := []models.TimeOff{{EmployeeID: employeeID, StartsAt: breakStarts, EndsAt: breakEnds, Reason: "перерыв"}}

	return opens, closes, breaks, true, nil
}

func companyLocation(companies CompanyLookup, companyID int) (*time.Location, error) {
	company, err := companies.GetCompanyById(companyID)
	if err != nil {
		return nil, err
	}

	return company.Location()
}
