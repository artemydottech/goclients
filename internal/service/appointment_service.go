package service

import (
	"database/sql"
	"errors"
	"time"
	"unicode/utf8"

	"github.com/artemydottech/goclients/internal/models"
)

type AppointmentRepo interface {
	Create(a models.Appointment) (int64, error)
	GetAllAppointments() ([]models.Appointment, error)
	GetAppointmentsByEmployee(employeeID int, from, to time.Time) ([]models.Appointment, error)
	GetAppointmentsByClient(clientID int) ([]models.Appointment, error)
	GetAppointmentById(id int) (models.Appointment, error)
	CreateIfFree(a models.Appointment) (int64, bool, error)
	MoveIfFree(id int, a models.Appointment) (bool, error)
	UpdateStatus(id int, status models.AppointmentStatus) error
	DeleteAppointmentById(id int) error
}

type ClientLookup interface {
	GetClientById(id int) (models.Client, error)
}

type Assignments interface {
	EmployeePerformsService(employeeID, serviceID int) (bool, error)
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

// Book заводит запись. Конец интервала считается здесь: если бы его присылал
// клиент, он мог бы занять мастера на пять минут вместо часа.
func (s *AppointmentService) Book(a models.Appointment) (int64, error) {
	prepared, err := s.prepare(a)
	if err != nil {
		return 0, err
	}

	id, busy, err := s.repo.CreateIfFree(prepared)
	if err != nil {
		return 0, err
	}
	if busy {
		return 0, models.Invalid("Это время у сотрудника уже занято!")
	}

	return id, nil
}

// Reschedule переносит запись на другое время и, если указан, к другому мастеру.
// Перенос проходит те же проверки, что и новая запись: иначе через него можно
// было бы обойти график или посадить клиента к мастеру без нужной услуги.
func (s *AppointmentService) Reschedule(id int, startsAt time.Time, employeeID int) (models.Appointment, error) {
	existing, err := s.repo.GetAppointmentById(id)
	if err != nil {
		return models.Appointment{}, err
	}

	if !existing.Status.Blocks() || existing.Status == models.AppointmentCompleted {
		return models.Appointment{}, models.Invalid("Перенести можно только активную запись!")
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

	busy, err := s.repo.MoveIfFree(id, prepared)
	if err != nil {
		return models.Appointment{}, err
	}
	if busy {
		return models.Appointment{}, models.Invalid("Это время у сотрудника уже занято!")
	}

	return prepared, nil
}

func (s *AppointmentService) prepare(a models.Appointment) (models.Appointment, error) {
	if a.StartsAt.IsZero() {
		return a, models.Invalid("Не указано время записи!")
	}

	if !a.StartsAt.After(s.now()) {
		return a, models.Invalid("Записаться можно только на будущее время!")
	}

	if utf8.RuneCountInString(a.Comment) > 1000 {
		return a, models.Invalid("Комментарий слишком длинный! Не превышайте 1000 символов")
	}

	client, err := s.clients.GetClientById(a.ClientID)
	if errors.Is(err, sql.ErrNoRows) {
		return a, models.Invalid("Клиент %d не найден!", a.ClientID)
	}
	if err != nil {
		return a, err
	}

	employee, err := s.employees.GetEmployeeById(a.EmployeeID)
	if errors.Is(err, sql.ErrNoRows) {
		return a, models.Invalid("Сотрудник %d не найден!", a.EmployeeID)
	}
	if err != nil {
		return a, err
	}

	item, err := s.services.GetServiceById(a.ServiceID)
	if errors.Is(err, sql.ErrNoRows) {
		return a, models.Invalid("Услуга %d не найдена!", a.ServiceID)
	}
	if err != nil {
		return a, err
	}

	if client.CompanyID != employee.CompanyID || item.CompanyID != employee.CompanyID {
		return a, models.Invalid("Клиент, сотрудник и услуга должны быть из одной компании!")
	}

	performs, err := s.assignments.EmployeePerformsService(a.EmployeeID, a.ServiceID)
	if err != nil {
		return a, err
	}
	if !performs {
		return a, models.Invalid("Сотрудник %d не оказывает услугу %d!", a.EmployeeID, a.ServiceID)
	}

	a.CompanyID = employee.CompanyID
	a.Price = item.Price
	a.StartsAt = a.StartsAt.UTC()
	a.EndsAt = a.StartsAt.Add(time.Duration(item.Duration) * time.Minute)

	if a.Status == "" {
		a.Status = models.AppointmentPending
	}
	if !a.Status.Valid() {
		return a, models.Invalid("Неизвестный статус записи %s!", a.Status)
	}
	if a.Status != models.AppointmentPending && a.Status != models.AppointmentConfirmed {
		return a, models.Invalid("Новая запись может быть только pending или confirmed!")
	}

	loc, err := companyLocation(s.companies, employee.CompanyID)
	if err != nil {
		return a, err
	}

	local := a.StartsAt.In(loc)
	day := time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, loc)

	opens, closes, breaks, isWorkingDay, err := workingWindow(s.schedule, a.EmployeeID, day)
	if err != nil {
		return a, err
	}
	if !isWorkingDay {
		return a, models.Invalid("У сотрудника в этот день выходной!")
	}
	if a.StartsAt.Before(opens) || a.EndsAt.After(closes) {
		return a, models.Invalid(
			"Запись выходит за рабочее время сотрудника (%s–%s)!",
			opens.Format("15:04"), closes.Format("15:04"),
		)
	}

	if overlapsTimeOff(a.StartsAt, a.EndsAt, breaks) {
		return a, models.Invalid(
			"Запись попадает на перерыв сотрудника (%s–%s)!",
			breaks[0].StartsAt.Format("15:04"), breaks[0].EndsAt.Format("15:04"),
		)
	}

	absences, err := s.timeOff.GetTimeOffInRange(a.EmployeeID, a.StartsAt, a.EndsAt)
	if err != nil {
		return a, err
	}
	if len(absences) > 0 {
		return a, models.Invalid("Сотрудник в это время не принимает (%s)!", absenceReason(absences[0]))
	}

	return a, nil
}

func absenceReason(t models.TimeOff) string {
	if t.Reason == "" {
		return "нерабочий период"
	}
	return t.Reason
}

func (s *AppointmentService) GetAllAppointments() ([]models.Appointment, error) {
	return s.repo.GetAllAppointments()
}

func (s *AppointmentService) GetAppointmentsByEmployee(employeeID int, from, to time.Time) ([]models.Appointment, error) {
	return s.repo.GetAppointmentsByEmployee(employeeID, from, to)
}

func (s *AppointmentService) GetAppointmentsByClient(clientID int) ([]models.Appointment, error) {
	return s.repo.GetAppointmentsByClient(clientID)
}

func (s *AppointmentService) GetAppointmentById(id int) (models.Appointment, error) {
	return s.repo.GetAppointmentById(id)
}

func (s *AppointmentService) SetStatus(id int, status models.AppointmentStatus) error {
	if !status.Valid() {
		return models.Invalid("Неизвестный статус записи %s!", status)
	}

	existing, err := s.repo.GetAppointmentById(id)
	if err != nil {
		return err
	}

	if !existing.Status.CanBecome(status) {
		if existing.Status == models.AppointmentCancelled {
			return models.Invalid("Отменённую запись нельзя вернуть — создайте новую!")
		}
		return models.Invalid("Нельзя сменить статус %s на %s!", existing.Status, status)
	}

	if status == models.AppointmentCompleted && s.now().Before(existing.StartsAt) {
		return models.Invalid("Завершить можно только начавшуюся запись!")
	}

	return s.repo.UpdateStatus(id, status)
}

func (s *AppointmentService) DeleteAppointmentById(id int) error {
	return s.repo.DeleteAppointmentById(id)
}
