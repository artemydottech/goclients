package service

import (
	"unicode/utf8"

	"github.com/artemydottech/goclients/internal/models"
)

type EmployeeRepo interface {
	Create(c models.Employee) (int64, error)
	GetAllEmployees() ([]models.Employee, error)
	GetEmployeeById(id int) (models.Employee, error)
	DeleteEmployeeById(id int) error
}

type EmployeeService struct {
	repo EmployeeRepo
}

func NewEmployeeService(repo EmployeeRepo) *EmployeeService {
	return &EmployeeService{repo: repo}
}

func (s *EmployeeService) CreateEmployee(e models.Employee) (int64, error) {
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

	id, err := s.repo.Create(e)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (s *EmployeeService) GetAllEmployees() ([]models.Employee, error) {
	return s.repo.GetAllEmployees()
}

func (s *EmployeeService) GetEmployeeById(id int) (models.Employee, error) {
	return s.repo.GetEmployeeById(id)
}

func (s *EmployeeService) DeleteEmployeeById(id int) error {
	return s.repo.DeleteEmployeeById(id)
}
