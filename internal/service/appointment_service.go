package service

import (
	"context"
	"errors"
	"time"
	"unicode/utf8"

	"github.com/artemydottech/goclients/internal/models"
)

type AppointmentRepo interface {
	Create(ctx context.Context, a models.Appointment) (int64, error)
	GetAllAppointments(ctx context.Context) ([]models.Appointment, error)
	GetAppointmentsByEmployee(ctx context.Context, employeeID int, from, to time.Time) ([]models.Appointment, error)
	GetAppointmentsByClient(ctx context.Context, clientID int) ([]models.Appointment, error)
	GetAppointmentById(ctx context.Context, id int) (models.Appointment, error)
	CreateIfFree(ctx context.Context, a models.Appointment) (int64, bool, error)
	MoveIfFree(ctx context.Context, id int, a models.Appointment) (bool, error)
	UpdateStatus(ctx context.Context, id int, status models.AppointmentStatus) error
	GetClientStats(ctx context.Context, clientID int, now time.Time) (models.ClientStats, error)
	DeleteAppointmentById(ctx context.Context, id int) error
}

type ClientLookup interface {
	GetClientById(ctx context.Context, id int) (models.Client, error)
}

type Assignments interface {
	EmployeePerformsService(ctx context.Context, employeeID, serviceID int) (bool, error)
}

type AppointmentService struct {
	repo        AppointmentRepo
	clients     ClientLookup
	employees   EmployeeLookup
	services    ServiceLookup
	assignments Assignments
	schedule    Schedule
	companies   CompanyLookup
	timeOff     TimeOffCalendar
	now         func() time.Time
}

func NewAppointmentService(
	repo AppointmentRepo,
	clients ClientLookup,
	employees EmployeeLookup,
	services ServiceLookup,
	assignments Assignments,
	schedule Schedule,
	companies CompanyLookup,
	timeOff TimeOffCalendar,
) *AppointmentService {
	return &AppointmentService{
		repo:        repo,
		clients:     clients,
		employees:   employees,
		services:    services,
		assignments: assignments,
		schedule:    schedule,
		companies:   companies,
		timeOff:     timeOff,
		now:         time.Now,
	}
}

func (s *AppointmentService) Book(a models.Appointment) (int64, error) {
	prepared, err := s.prepare(a)
	if err != nil {
		return 0, err
	}

	id, busy, err := s.repo.CreateIfFree(context.TODO(), prepared)
	if err != nil {
		return 0, err
	}
	if busy {
		return 0, models.Invalid("this time is already taken for the employee")
	}

	return id, nil
}

func (s *AppointmentService) Reschedule(id int, startsAt time.Time, employeeID int) (models.Appointment, error) {
	existing, err := s.repo.GetAppointmentById(context.TODO(), id)
	if err != nil {
		return models.Appointment{}, err
	}

	if !existing.Status.Active() {
		return models.Appointment{}, models.Invalid("only an active appointment can be rescheduled")
	}

	candidate := existing
	candidate.StartsAt = startsAt
	if employeeID != 0 {
		candidate.EmployeeID = employeeID
	}

	prepared, err := s.prepare(candidate)
	if err != nil {
		return models.Appointment{}, err
	}

	prepared.Price = existing.Price

	busy, err := s.repo.MoveIfFree(context.TODO(), id, prepared)
	if err != nil {
		return models.Appointment{}, err
	}
	if busy {
		return models.Appointment{}, models.Invalid("this time is already taken for the employee")
	}

	return prepared, nil
}

func (s *AppointmentService) prepare(a models.Appointment) (models.Appointment, error) {
	if a.StartsAt.IsZero() {
		return a, models.Invalid("appointment time is not set")
	}

	if !a.StartsAt.After(s.now()) {
		return a, models.Invalid("appointments can only be booked in the future")
	}

	if utf8.RuneCountInString(a.Comment) > 1000 {
		return a, models.Invalid("comment is too long, keep it under 1000 characters")
	}

	client, err := s.clients.GetClientById(context.TODO(), a.ClientID)
	if errors.Is(err, models.ErrNotFound) {
		return a, models.Invalid("client %d not found", a.ClientID)
	}
	if err != nil {
		return a, err
	}

	employee, err := s.employees.GetEmployeeById(context.TODO(), a.EmployeeID)
	if errors.Is(err, models.ErrNotFound) {
		return a, models.Invalid("employee %d not found", a.EmployeeID)
	}
	if err != nil {
		return a, err
	}

	item, err := s.services.GetServiceById(context.TODO(), a.ServiceID)
	if errors.Is(err, models.ErrNotFound) {
		return a, models.Invalid("service %d not found", a.ServiceID)
	}
	if err != nil {
		return a, err
	}

	if client.CompanyID != employee.CompanyID || item.CompanyID != employee.CompanyID {
		return a, models.Invalid("client, employee and service must belong to the same company")
	}

	performs, err := s.assignments.EmployeePerformsService(context.TODO(), a.EmployeeID, a.ServiceID)
	if err != nil {
		return a, err
	}
	if !performs {
		return a, models.Invalid("employee %d does not provide service %d", a.EmployeeID, a.ServiceID)
	}

	a.CompanyID = employee.CompanyID
	a.Price = item.Price
	a.StartsAt = a.StartsAt.UTC()
	a.EndsAt = a.StartsAt.Add(time.Duration(item.Duration) * time.Minute)

	if a.Status == "" {
		a.Status = models.AppointmentPending
	}
	if !a.Status.Valid() {
		return a, models.Invalid("unknown appointment status %s", a.Status)
	}
	if a.Status != models.AppointmentPending && a.Status != models.AppointmentConfirmed {
		return a, models.Invalid("a new appointment can only be pending or confirmed")
	}

	loc, err := companyLocation(context.TODO(), s.companies, employee.CompanyID)
	if err != nil {
		return a, err
	}

	local := a.StartsAt.In(loc)
	day := time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, loc)

	opens, closes, breaks, isWorkingDay, err := workingWindow(context.TODO(), s.schedule, a.EmployeeID, day)
	if err != nil {
		return a, err
	}
	if !isWorkingDay {
		return a, models.Invalid("the employee has a day off")
	}
	if a.StartsAt.Before(opens) || a.EndsAt.After(closes) {
		return a, models.Invalid(
			"appointment is outside the employee working hours (%s-%s)",
			opens.Format("15:04"), closes.Format("15:04"),
		)
	}

	if overlapsTimeOff(a.StartsAt, a.EndsAt, breaks) {
		return a, models.Invalid(
			"appointment falls on the employee break (%s-%s)",
			breaks[0].StartsAt.Format("15:04"), breaks[0].EndsAt.Format("15:04"),
		)
	}

	absences, err := s.timeOff.GetTimeOffInRange(context.TODO(), a.EmployeeID, a.StartsAt, a.EndsAt)
	if err != nil {
		return a, err
	}
	if len(absences) > 0 {
		return a, models.Invalid("the employee is unavailable at this time (%s)", absenceReason(absences[0]))
	}

	return a, nil
}

func absenceReason(t models.TimeOff) string {
	if t.Reason == "" {
		return "time off"
	}
	return t.Reason
}

func (s *AppointmentService) GetAllAppointments() ([]models.Appointment, error) {
	return s.repo.GetAllAppointments(context.TODO())
}

func (s *AppointmentService) GetAppointmentsByEmployee(employeeID int, from, to time.Time) ([]models.Appointment, error) {
	return s.repo.GetAppointmentsByEmployee(context.TODO(), employeeID, from, to)
}

func (s *AppointmentService) GetAppointmentsByClient(clientID int) ([]models.Appointment, error) {
	return s.repo.GetAppointmentsByClient(context.TODO(), clientID)
}

func (s *AppointmentService) GetAppointmentById(id int) (models.Appointment, error) {
	return s.repo.GetAppointmentById(context.TODO(), id)
}

func (s *AppointmentService) SetStatus(id int, status models.AppointmentStatus) error {
	if !status.Valid() {
		return models.Invalid("unknown appointment status %s", status)
	}

	existing, err := s.repo.GetAppointmentById(context.TODO(), id)
	if err != nil {
		return err
	}

	if !existing.Status.CanBecome(status) {
		if existing.Status == models.AppointmentCancelled {
			return models.Invalid("a cancelled appointment cannot be restored, create a new one")
		}
		return models.Invalid("cannot change status from %s to %s", existing.Status, status)
	}

	if status.Happened() && s.now().Before(existing.StartsAt) {
		return models.Invalid("status %s can only be set after the appointment has started", status)
	}

	return s.repo.UpdateStatus(context.TODO(), id, status)
}

func (s *AppointmentService) GetClientStats(clientID int) (models.ClientStats, error) {
	if _, err := s.clients.GetClientById(context.TODO(), clientID); err != nil {
		return models.ClientStats{}, err
	}

	return s.repo.GetClientStats(context.TODO(), clientID, s.now())
}

func (s *AppointmentService) DeleteAppointmentById(id int) error {
	return s.repo.DeleteAppointmentById(context.TODO(), id)
}
