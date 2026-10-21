package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/artemydottech/goclients/internal/models"
)

type stubCompanyRepo struct {
	called  bool
	updated models.Company
	err     error
}

func (r *stubCompanyRepo) Create(context.Context, models.Company) (int64, error) {
	r.called = true
	return 3, nil
}

func (r *stubCompanyRepo) GetAllCompanies(context.Context, models.Page) ([]models.Company, error) {
	return nil, nil
}

func (r *stubCompanyRepo) GetCompanyById(context.Context, int) (models.Company, error) {
	return models.Company{}, nil
}

func (r *stubCompanyRepo) UpdateCompany(_ context.Context, c models.Company) error {
	r.called = true
	r.updated = c
	return r.err
}

func (r *stubCompanyRepo) DeleteCompanyById(context.Context, int) error { return nil }

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

			_, err := NewCompanyService(repo).CreateCompany(context.Background(), company)

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

func TestCreateCompanyCountsRunesInSite(t *testing.T) {
	company := models.Company{Name: "Ромашка", Site: strings.Repeat("я", 500)}

	if _, err := NewCompanyService(&stubCompanyRepo{}).CreateCompany(context.Background(), company); err != nil {
		t.Fatalf("ожидался успех, получена ошибка %v", err)
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

	id, err := NewCompanyService(&stubCompanyRepo{}).CreateCompany(context.Background(), company)
	if err != nil {
		t.Fatalf("ожидался успех, получена ошибка %v", err)
	}

	if id != 3 {
		t.Errorf("ожидался id 3, получен %d", id)
	}
}

func TestCreateCompanyRejectsUnknownTimezone(t *testing.T) {
	_, err := NewCompanyService(&stubCompanyRepo{}).CreateCompany(context.Background(), models.Company{
		Name:     "Ромашка",
		Timezone: "Mars/Olympus",
	})

	var validationErr models.ValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("ожидалась ValidationError, получено %v", err)
	}
}

func TestCreateCompanyAcceptsIANATimezone(t *testing.T) {
	_, err := NewCompanyService(&stubCompanyRepo{}).CreateCompany(context.Background(), models.Company{
		Name:     "Ромашка",
		Timezone: "Asia/Yekaterinburg",
	})
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}
}

func TestUpdateCompanyValidates(t *testing.T) {
	cases := map[string]models.Company{
		"empty name":      {Name: ""},
		"name too long":   {Name: strings.Repeat("a", 201)},
		"unknown network": {Name: "Romashka", Socials: models.Socials{"myspace": "@x"}},
		"unknown zone":    {Name: "Romashka", Timezone: "Mars/Olympus"},
	}

	for name, company := range cases {
		t.Run(name, func(t *testing.T) {
			repo := &stubCompanyRepo{}

			err := NewCompanyService(repo).UpdateCompany(context.Background(), 1, company)

			var validationErr models.ValidationError
			if !errors.As(err, &validationErr) {
				t.Fatalf("expected ValidationError, got %v", err)
			}
			if repo.called {
				t.Error("repository must not be called for invalid input")
			}
		})
	}
}

func TestUpdateCompanyUsesThePathIDAndDefaultsTimezone(t *testing.T) {
	repo := &stubCompanyRepo{}

	err := NewCompanyService(repo).UpdateCompany(context.Background(), 7, models.Company{ID: 99, Name: "Romashka"})
	if err != nil {
		t.Fatalf("UpdateCompany: %v", err)
	}

	if repo.updated.ID != 7 {
		t.Errorf("ID = %d, want the path id 7", repo.updated.ID)
	}
	if repo.updated.Timezone != "UTC" {
		t.Errorf("Timezone = %q, want UTC", repo.updated.Timezone)
	}
}

func TestUpdateCompanyPassesRepositoryErrors(t *testing.T) {
	repo := &stubCompanyRepo{err: models.ErrNotFound}

	err := NewCompanyService(repo).UpdateCompany(context.Background(), 1, models.Company{Name: "Romashka"})

	if !errors.Is(err, models.ErrNotFound) {
		t.Fatalf("got %v, want models.ErrNotFound", err)
	}
}
