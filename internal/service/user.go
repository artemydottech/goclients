package service

import (
	"errors"
	"unicode/utf8"
)

type UserRepo interface {
	Create(name string) (int64, error)
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
		return 0, errors.New("Имя слишком длинное! Не превышай 100 символов")
	}

	id, err := s.repo.Create(name)
	if err != nil {
		return 0, err
	}

	return id, nil
}