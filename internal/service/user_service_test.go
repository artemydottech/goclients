package service

import (
	"errors"
	"strings"
	"testing"

	"github.com/artemydottech/goclients/internal/models"
)

type stubUserRepo struct {
	created models.User
}

func (r *stubUserRepo) Create(u models.User) (int64, error) {
	r.created = u
	return 7, nil
}

func (r *stubUserRepo) GetAllUsers() ([]models.User, error) { return nil, nil }

func (r *stubUserRepo) GetUserById(int) (models.User, error) { return models.User{}, nil }

func (r *stubUserRepo) DeleteUserById(int) error { return nil }

func TestRegisterUserRejectsEmptyName(t *testing.T) {
	repo := &stubUserRepo{}

	_, err := NewUserService(repo).RegisterUser(models.User{})

	var validationErr models.ValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("ожидалась ValidationError, получено %v", err)
	}

	if repo.created.Name != "" {
		t.Error("репозиторий не должен вызываться при невалидном имени")
	}
}

func TestRegisterUserRejectsNameOver100Runes(t *testing.T) {
	_, err := NewUserService(&stubUserRepo{}).RegisterUser(models.User{Name: strings.Repeat("a", 101)})

	var validationErr models.ValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("ожидалась ValidationError, получено %v", err)
	}
}

func TestRegisterUserCountsRunesNotBytes(t *testing.T) {
	name := strings.Repeat("я", 100)
	repo := &stubUserRepo{}

	id, err := NewUserService(repo).RegisterUser(models.User{Name: name})
	if err != nil {
		t.Fatalf("ожидался успех, получена ошибка %v", err)
	}

	if id != 7 {
		t.Errorf("ожидался id 7, получен %d", id)
	}

	if repo.created.Name != name {
		t.Errorf("в репозиторий ушло %q", repo.created.Name)
	}
}

func TestRegisterUserKeepsTheOptionalFields(t *testing.T) {
	repo := &stubUserRepo{}
	user := models.User{Name: "Артемий", Surname: "Зверев", Username: "artemy", Avatar: "a.png"}

	if _, err := NewUserService(repo).RegisterUser(user); err != nil {
		t.Fatalf("ожидался успех, получена ошибка %v", err)
	}

	if repo.created != user {
		t.Errorf("в репозиторий ушло %+v", repo.created)
	}
}

func TestRegisterUserRejectsLongUsername(t *testing.T) {
	user := models.User{Name: "Артемий", Username: strings.Repeat("a", 51)}

	_, err := NewUserService(&stubUserRepo{}).RegisterUser(user)

	var validationErr models.ValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("ожидалась ValidationError, получено %v", err)
	}
}
