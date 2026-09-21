package service

import (
	"unicode/utf8"

	"github.com/artemydottech/goclients/internal/models"
)

type UserRepo interface {
	Create(u models.User) (int64, error)
	GetAllUsers() ([]models.User, error)
	GetUserById(id int) (models.User, error)
	DeleteUserById(id int) error
}

type UserService struct {
	repo UserRepo
}

func NewUserService(repo UserRepo) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) RegisterUser(u models.User) (int64, error) {
	if u.Name == "" {
		return 0, models.Invalid("name cannot be empty")
	}

	if utf8.RuneCountInString(u.Name) > 100 {
		return 0, models.Invalid("name is too long, keep it under 100 characters")
	}

	if utf8.RuneCountInString(u.Surname) > 100 {
		return 0, models.Invalid("surname is too long, keep it under 100 characters")
	}

	if utf8.RuneCountInString(u.Username) > 50 {
		return 0, models.Invalid("username is too long, keep it under 50 characters")
	}

	id, err := s.repo.Create(u)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (s *UserService) GetAllUsers() ([]models.User, error) {
	return s.repo.GetAllUsers()
}

func (s *UserService) GetUserById(id int) (models.User, error) {
	return s.repo.GetUserById(id)
}

func (s *UserService) DeleteUserById(id int) error {
	return s.repo.DeleteUserById(id)
}
