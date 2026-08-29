package service

import (
	"errors"
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
		return 0, errors.New("Имя сотрудника не может быть пустым!")
	}

	if e.Surname == "" {
		return 0, errors.New("Фамилия сотрудника не может быть пустой!")
	}

	if utf8.RuneCountInString(e.Name) > 200 {
		return 0, errors.New("Имя слишком длинное! Не превышайте 200 символов")
	}

	if utf8.RuneCountInString(e.Surname) > 200 {
		return 0, errors.New("Фамилия слишком длинная! Не превышайте 200 символов")
	}

	if e.Position != "" && len(e.Position) > 500 {
		return 0, errors.New("Должность слишком длинная! Не превышайте 500 символов")
	}

	if e.Avatar != "" && len(e.Avatar) > 500 {
		return 0, errors.New("Ссылка на аватар слишком длинная! Не превышайте 500 символов")
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
