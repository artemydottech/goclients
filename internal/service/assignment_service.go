package service

import (
	"context"
	"errors"

	"github.com/artemydottech/goclients/internal/models"
)

type AssignmentRepo interface {
	SetEmployeeServices(ctx context.Context, employeeID int, serviceIDs []int) error
	GetServicesByEmployee(ctx context.Context, employeeID int) ([]models.Service, error)
	GetEmployeesByService(ctx context.Context, serviceID int) ([]models.Employee, error)
	EmployeePerformsService(ctx context.Context, employeeID, serviceID int) (bool, error)
}

type EmployeeLookup interface {
	GetEmployeeById(ctx context.Context, id int) (models.Employee, error)
}

type ServiceLookup interface {
	GetServiceById(ctx context.Context, id int) (models.Service, error)
}

type AssignmentService struct {
	repo      AssignmentRepo
	employees EmployeeLookup
	services  ServiceLookup
}

func NewAssignmentService(repo AssignmentRepo, employees EmployeeLookup, services ServiceLookup) *AssignmentService {
	return &AssignmentService{repo: repo, employees: employees, services: services}
}

func (s *AssignmentService) SetEmployeeServices(ctx context.Context, employeeID int, serviceIDs []int) error {
	employee, err := s.employees.GetEmployeeById(ctx, employeeID)
	if errors.Is(err, models.ErrNotFound) {
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

		item, err := s.services.GetServiceById(ctx, serviceID)
		if errors.Is(err, models.ErrNotFound) {
			return models.Invalid("service %d not found", serviceID)
		}
		if err != nil {
			return err
		}

		if item.CompanyID != employee.CompanyID {
			return models.Invalid("service %d belongs to another company", serviceID)
		}

		unique = append(unique, serviceID)
	}

	return s.repo.SetEmployeeServices(ctx, employeeID, unique)
}

func (s *AssignmentService) GetServicesByEmployee(ctx context.Context, employeeID int) ([]models.Service, error) {
	return s.repo.GetServicesByEmployee(ctx, employeeID)
}

func (s *AssignmentService) GetEmployeesByService(ctx context.Context, serviceID int) ([]models.Employee, error) {
	return s.repo.GetEmployeesByService(ctx, serviceID)
}

func (s *AssignmentService) EmployeePerformsService(employeeID, serviceID int) (bool, error) {
	return s.repo.EmployeePerformsService(context.TODO(), employeeID, serviceID)
}
