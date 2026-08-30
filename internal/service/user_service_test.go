package service

import (
	"errors"
	"strings"
	"testing"

	"github.com/artemydottech/goclients/internal/models"
)

type stubUserRepo struct {
	createdName string
}

func (r *stubUserRepo) Create(name string) (int64, error) {
	r.createdName = name
	return 7, nil
}

func (r *stubUserRepo) GetAllUsers() ([]models.User, error) { return nil, nil }

func (r *stubUserRepo) GetUserById(int) (models.User, error) { return models.User{}, nil }

func (r *stubUserRepo) DeleteUserById(int) error { return nil }

func TestRegisterUserRejectsEmptyName(t *testing.T) {
	repo := &stubUserRepo{}

	_, err := NewUserService(repo).RegisterUser("")

	var validationErr models.ValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("ожидалась ValidationError, получено %v", err)
	}

	if repo.createdName != "" {
		t.Error("репозиторий не должен вызываться при невалидном имени")
	}
}

func TestRegisterUserRejectsNameOver100Runes(t *testing.T) {
	_, err := NewUserService(&stubUserRepo{}).RegisterUser(strings.Repeat("a", 101))

	var validationErr models.ValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("ожидалась ValidationError, получено %v", err)
	}
}

// Лимит считается в рунах, а не в байтах: 100 кириллических символов — это
// 200 байт, и они должны проходить.
func TestRegisterUserCountsRunesNotBytes(t *testing.T) {
	name := strings.Repeat("я", 100)
	repo := &stubUserRepo{}

	id, err := NewUserService(repo).RegisterUser(name)
	if err != nil {
		t.Fatalf("ожидался успех, получена ошибка %v", err)
	}

	if id != 7 {
		t.Errorf("ожидался id 7, получен %d", id)
	}

	if repo.createdName != name {
		t.Errorf("в репозиторий ушло %q", repo.createdName)
	}
}
