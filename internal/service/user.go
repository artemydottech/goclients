package service

import (
	"errors"
	"unicode/utf8"

	"github.com/artemydottech/goclients/internal/models"
)

type UserRepo interface {
	Create(name string) (int64, error)
	GetAllUsers() ([]models.User, error)
	GetUserById(id int) (models.User, error)
}

type UserService struct {
	repo UserRepo
}

func NewUserService (repo UserRepo) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) RegisterUser(name string) (int64, error) {
	if name == "" {
		return 0, errors.New("Имя не может быть пустым!")
	}

	if utf8.RuneCountInString(name) > 100 {
		return 0, errors.New("Имя слишком длинное! Не превышайте 100 символов")
	}

	id, err := s.repo.Create(name)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (s* UserService) GetAllUsers() ([]models.User, error) {
	return s.repo.GetAllUsers()
}

func (s* UserService) GetUserById(id int) (models.User, error) {
	return s.repo.GetUserById(id)
}