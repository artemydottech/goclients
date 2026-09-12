package service

import (
	"database/sql"
	"errors"

	"github.com/artemydottech/goclients/internal/models"
)

type AssignmentRepo interface {
	SetEmployeeServices(employeeID int, serviceIDs []int) error
	GetServicesByEmployee(employeeID int) ([]models.Service, error)
	GetEmployeesByService(serviceID int) ([]models.Employee, error)
	EmployeePerformsService(employeeID, serviceID int) (bool, error)
}

type EmployeeLookup interface {
	GetEmployeeById(id int) (models.Employee, error)
}

type ServiceLookup interface {
	GetServiceById(id int) (models.Service, error)
}

type AssignmentService struct {
	repo      AssignmentRepo
	employees EmployeeLookup
	services  ServiceLookup
}

func NewAssignmentService(repo AssignmentRepo, employees EmployeeLookup, services ServiceLookup) *AssignmentService {
	return &AssignmentService{repo: repo, employees: employees, services: services}
}

// SetEmployeeServices принимает набор услуг сотрудника. Услуга чужой компании
// в наборе — ошибка ввода: мастер одного салона не оказывает услуги другого.
func (s *AssignmentService) SetEmployeeServices(employeeID int, serviceIDs []int) error {
	employee, err := s.employees.GetEmployeeById(employeeID)
	if errors.Is(err, sql.ErrNoRows) {
		return err
	}
	if err != nil {
		return err
	}

	seen := make(map[int]struct{}, len(serviceIDs))
	unique := make([]int, 0, len(serviceIDs))

	for _, serviceID := range serviceIDs {
		if _, duplicate := seen[serviceID]; duplicate {
			continue
		}
		seen[serviceID] = struct{}{}

		item, err := s.services.GetServiceById(serviceID)
		if errors.Is(err, sql.ErrNoRows) {
			return models.Invalid("Услуга %d не найдена!", serviceID)
		}
		if err != nil {
			return err
		}

		if item.CompanyID != employee.CompanyID {
			return models.Invalid("Услуга %d принадлежит другой компании!", serviceID)
		}

		unique = append(unique, serviceID)
	}

	return s.repo.SetEmployeeServices(employeeID, unique)
}

func (s *AssignmentService) GetServicesByEmployee(employeeID int) ([]models.Service, error) {
	return s.repo.GetServicesByEmployee(employeeID)
}

func (s *AssignmentService) GetEmployeesByService(serviceID int) ([]models.Employee, error) {
	return s.repo.GetEmployeesByService(serviceID)
}

func (s *AssignmentService) EmployeePerformsService(employeeID, serviceID int) (bool, error) {
	return s.repo.EmployeePerformsService(employeeID, serviceID)
}
