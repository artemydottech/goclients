package service

import (
	"context"
	"unicode/utf8"

	"github.com/artemydottech/goclients/internal/models"
)

type EmployeeRepo interface {
	Create(ctx context.Context, c models.Employee) (int64, error)
	GetAllEmployees(ctx context.Context) ([]models.Employee, error)
	GetEmployeeById(ctx context.Context, id int) (models.Employee, error)
	DeleteEmployeeById(ctx context.Context, id int) error
}

type EmployeeService struct {
	repo EmployeeRepo
}

func NewEmployeeService(repo EmployeeRepo) *EmployeeService {
	return &EmployeeService{repo: repo}
}

func (s *EmployeeService) CreateEmployee(ctx context.Context, e models.Employee) (int64, error) {
	if e.Name == "" {
		return 0, models.Invalid("employee name cannot be empty")
	}

	if e.Surname == "" {
		return 0, models.Invalid("employee surname cannot be empty")
	}

	if utf8.RuneCountInString(e.Name) > 200 {
		return 0, models.Invalid("name is too long, keep it under 200 characters")
	}

	if utf8.RuneCountInString(e.Surname) > 200 {
		return 0, models.Invalid("surname is too long, keep it under 200 characters")
	}

	if e.Position != "" && utf8.RuneCountInString(e.Position) > 500 {
		return 0, models.Invalid("position is too long, keep it under 500 characters")
	}

	if e.Avatar != "" && utf8.RuneCountInString(e.Avatar) > 500 {
		return 0, models.Invalid("avatar URL is too long, keep it under 500 characters")
	}

	id, err := s.repo.Create(ctx, e)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (s *EmployeeService) GetAllEmployees(ctx context.Context) ([]models.Employee, error) {
	return s.repo.GetAllEmployees(ctx)
}

func (s *EmployeeService) GetEmployeeById(ctx context.Context, id int) (models.Employee, error) {
	return s.repo.GetEmployeeById(ctx, id)
}

func (s *EmployeeService) DeleteEmployeeById(ctx context.Context, id int) error {
	return s.repo.DeleteEmployeeById(ctx, id)
}
