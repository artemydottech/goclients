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
	now         func() time.Time
}

func NewAppointmentService(
	repo AppointmentRepo,
	clients ClientLookup,
	employees EmployeeLookup,
	services ServiceLookup,
	assignments Assignments,
) *AppointmentService {
	return &AppointmentService{
		repo:        repo,
		clients:     clients,
		employees:   employees,
		services:    services,
		assignments: assignments,
		now:         time.Now,
	}
}

// Book заводит запись. Конец интервала считается здесь: если бы его присылал
// клиент, он мог бы занять мастера на пять минут вместо часа.
func (s *AppointmentService) Book(a models.Appointment) (int64, error) {
	if a.StartsAt.IsZero() {
		return 0, models.Invalid("Не указано время записи!")
	}

	if !a.StartsAt.After(s.now()) {
		return 0, models.Invalid("Записаться можно только на будущее время!")
	}

	if utf8.RuneCountInString(a.Comment) > 1000 {
		return 0, models.Invalid("Комментарий слишком длинный! Не превышайте 1000 символов")
	}

	client, err := s.clients.GetClientById(a.ClientID)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, models.Invalid("Клиент %d не найден!", a.ClientID)
	}
	if err != nil {
		return 0, err
	}

	employee, err := s.employees.GetEmployeeById(a.EmployeeID)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, models.Invalid("Сотрудник %d не найден!", a.EmployeeID)
	}
	if err != nil {
		return 0, err
	}

	item, err := s.services.GetServiceById(a.ServiceID)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, models.Invalid("Услуга %d не найдена!", a.ServiceID)
	}
	if err != nil {
		return 0, err
	}

	if client.CompanyID != employee.CompanyID || item.CompanyID != employee.CompanyID {
		return 0, models.Invalid("Клиент, сотрудник и услуга должны быть из одной компании!")
	}

	performs, err := s.assignments.EmployeePerformsService(a.EmployeeID, a.ServiceID)
	if err != nil {
		return 0, err
	}
	if !performs {
		return 0, models.Invalid("Сотрудник %d не оказывает услугу %d!", a.EmployeeID, a.ServiceID)
	}

	a.CompanyID = employee.CompanyID
	a.StartsAt = a.StartsAt.UTC()
	a.EndsAt = a.StartsAt.Add(time.Duration(item.Duration) * time.Minute)

	if a.Status == "" {
		a.Status = models.AppointmentPending
	}
	if !a.Status.Valid() {
		return 0, models.Invalid("Неизвестный статус записи %s!", a.Status)
	}

	id, busy, err := s.repo.CreateIfFree(a)
	if err != nil {
		return 0, err
	}
	if busy {
		return 0, models.Invalid("Это время у сотрудника уже занято!")
	}

	return id, nil
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

	return s.repo.UpdateStatus(id, status)
}

func (s *AppointmentService) DeleteAppointmentById(id int) error {
	return s.repo.DeleteAppointmentById(id)
}
