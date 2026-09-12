package service

import (
	"database/sql"
	"errors"
	"strings"
	"testing"

	"github.com/artemydottech/goclients/internal/models"
)

type stubClientRepo struct {
	called   bool
	existing *models.Client
	saved    models.Client
}

func (r *stubClientRepo) Create(c models.Client) (int64, error) {
	r.called = true
	r.saved = c
	return 3, nil
}

func (r *stubClientRepo) GetAllClients() ([]models.Client, error) { return nil, nil }

func (r *stubClientRepo) GetClientsByCompany(int) ([]models.Client, error) { return nil, nil }

func (r *stubClientRepo) GetClientById(int) (models.Client, error) {
	return models.Client{}, nil
}

func (r *stubClientRepo) GetClientByPhone(int, string) (models.Client, error) {
	if r.existing == nil {
		return models.Client{}, sql.ErrNoRows
	}
	return *r.existing, nil
}

func (r *stubClientRepo) DeleteClientById(int) error { return nil }

func validClient() models.Client {
	return models.Client{CompanyID: 1, Name: "Анна", Phone: "+7 (999) 123-45-67"}
}

func TestNormalizePhoneKeepsOnlyDigits(t *testing.T) {
	cases := map[string]string{
		"+7 (999) 123-45-67": "79991234567",
		"8-999-123-45-67":    "89991234567",
		"79991234567":        "79991234567",
		"":                   "",
	}

	for input, want := range cases {
		if got := NormalizePhone(input); got != want {
			t.Errorf("NormalizePhone(%q) = %q, ожидалось %q", input, got, want)
		}
	}
}

func TestCreateClientStoresTheNormalizedPhone(t *testing.T) {
	repo := &stubClientRepo{}

	if _, err := NewClientService(repo).CreateClient(validClient()); err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}

	if repo.saved.Phone != "79991234567" {
		t.Errorf("сохранён телефон %q, ожидался 79991234567", repo.saved.Phone)
	}
}

func TestCreateClientValidation(t *testing.T) {
	cases := map[string]models.Client{
		"без компании":     {CompanyID: 0, Name: "Анна", Phone: "79991234567"},
		"пустое имя":       {CompanyID: 1, Name: "", Phone: "79991234567"},
		"имя длиннее 200":  {CompanyID: 1, Name: strings.Repeat("а", 201), Phone: "79991234567"},
		"телефон без цифр": {CompanyID: 1, Name: "Анна", Phone: "не телефон"},
		"телефон короткий": {CompanyID: 1, Name: "Анна", Phone: "12345"},
		"телефон длинный":  {CompanyID: 1, Name: "Анна", Phone: "1234567890123456"},
		"почта без собаки": {CompanyID: 1, Name: "Анна", Phone: "79991234567", Email: "anna.ru"},
		"длинный комментарий": {
			CompanyID: 1, Name: "Анна", Phone: "79991234567",
			Comment: strings.Repeat("а", 1001),
		},
	}

	for name, client := range cases {
		t.Run(name, func(t *testing.T) {
			repo := &stubClientRepo{}

			_, err := NewClientService(repo).CreateClient(client)

			var validationErr models.ValidationError
			if !errors.As(err, &validationErr) {
				t.Fatalf("ожидалась ValidationError, получено %v", err)
			}

			if repo.called {
				t.Error("репозиторий не должен вызываться при невалидных данных")
			}
		})
	}
}

func TestCreateClientRejectsDuplicatePhone(t *testing.T) {
	repo := &stubClientRepo{existing: &models.Client{ID: 7, Phone: "79991234567"}}

	_, err := NewClientService(repo).CreateClient(validClient())

	var validationErr models.ValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("ожидалась ValidationError, получено %v", err)
	}

	if repo.called {
		t.Error("повторный телефон не должен доходить до вставки")
	}
}

func TestCreateClientCatchesDuplicateWrittenDifferently(t *testing.T) {
	repo := &stubClientRepo{existing: &models.Client{ID: 7, Phone: "79991234567"}}

	client := validClient()
	client.Phone = "7-999-123-45-67"

	if _, err := NewClientService(repo).CreateClient(client); err == nil {
		t.Fatal("тот же номер в другом написании должен считаться дублем")
	}
}
