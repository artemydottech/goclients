package service

import (
	"errors"
	"strings"
	"testing"

	"github.com/artemydottech/goclients/internal/models"
)

type stubCompanyRepo struct {
	called bool
}

func (r *stubCompanyRepo) Create(models.Company) (int64, error) {
	r.called = true
	return 3, nil
}

func (r *stubCompanyRepo) GetAllCompanies() ([]models.Company, error) { return nil, nil }

func (r *stubCompanyRepo) GetCompanyById(int) (models.Company, error) {
	return models.Company{}, nil
}

func (r *stubCompanyRepo) DeleteCompanyById(int) error { return nil }

func TestCreateCompanyValidation(t *testing.T) {
	cases := map[string]models.Company{
		"пустое название":       {Name: ""},
		"название длиннее 200":  {Name: strings.Repeat("a", 201)},
		"адрес длиннее 500":     {Name: "Ромашка", Address: strings.Repeat("a", 501)},
		"сайт длиннее 500":      {Name: "Ромашка", Site: strings.Repeat("a", 501)},
		"логотип длиннее 500":   {Name: "Ромашка", Logo: strings.Repeat("a", 501)},
		"неизвестная соц. сеть": {Name: "Ромашка", Socials: models.Socials{"myspace": "@x"}},
	}

	for name, company := range cases {
		t.Run(name, func(t *testing.T) {
			repo := &stubCompanyRepo{}

			_, err := NewCompanyService(repo).CreateCompany(company)

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

func TestCreateCompanyAcceptsKnownSocials(t *testing.T) {
	company := models.Company{
		Name: "Ромашка",
		Socials: models.Socials{
			models.SocialVK:       "vk.com/romashka",
			models.SocialTelegram: "@romashka",
		},
	}

	id, err := NewCompanyService(&stubCompanyRepo{}).CreateCompany(company)
	if err != nil {
		t.Fatalf("ожидался успех, получена ошибка %v", err)
	}

	if id != 3 {
		t.Errorf("ожидался id 3, получен %d", id)
	}
}
