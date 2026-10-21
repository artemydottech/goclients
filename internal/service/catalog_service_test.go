package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/artemydottech/goclients/internal/models"
)

type stubServiceRepo struct {
	updated models.Service
	err     error
	called  bool
}

func (r *stubServiceRepo) Create(context.Context, models.Service) (int64, error) {
	r.called = true
	return 9, nil
}

func (r *stubServiceRepo) GetAllServices(context.Context, models.Page) ([]models.Service, error) {
	return nil, nil
}

func (r *stubServiceRepo) GetServicesByCompany(context.Context, int, models.Page) ([]models.Service, error) {
	return nil, nil
}

func (r *stubServiceRepo) GetServiceById(context.Context, int) (models.Service, error) {
	return models.Service{}, nil
}

func (r *stubServiceRepo) UpdateService(_ context.Context, item models.Service) error {
	r.called = true
	r.updated = item
	return r.err
}

func (r *stubServiceRepo) DeleteServiceById(context.Context, int) error { return nil }

func validService() models.Service {
	return models.Service{CompanyID: 1, Name: "Стрижка", Duration: 60, Price: 1500}
}

func TestCreateServiceValidation(t *testing.T) {
	long := strings.Repeat("а", 201)

	cases := map[string]models.Service{
		"без компании":         {CompanyID: 0, Name: "Стрижка", Duration: 60},
		"пустое название":      {CompanyID: 1, Name: "", Duration: 60},
		"название длиннее 200": {CompanyID: 1, Name: long, Duration: 60},
		"описание длиннее 500": {
			CompanyID: 1, Name: "Стрижка", Duration: 60,
			Description: strings.Repeat("а", 501),
		},
		"нулевая длительность":       {CompanyID: 1, Name: "Стрижка", Duration: 0},
		"отрицательная длительность": {CompanyID: 1, Name: "Стрижка", Duration: -30},
		"длительность больше суток":  {CompanyID: 1, Name: "Стрижка", Duration: 24*60 + 1},
		"отрицательная цена":         {CompanyID: 1, Name: "Стрижка", Duration: 60, Price: -1},
	}

	for name, item := range cases {
		t.Run(name, func(t *testing.T) {
			repo := &stubServiceRepo{}

			_, err := NewCatalogService(repo).CreateService(context.Background(), item)

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

func TestCreateServiceAcceptsValidInput(t *testing.T) {
	repo := &stubServiceRepo{}

	id, err := NewCatalogService(repo).CreateService(context.Background(), validService())
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}

	if id != 9 {
		t.Errorf("получен id %d, ожидался 9", id)
	}

	if !repo.called {
		t.Error("репозиторий не был вызван")
	}
}

func TestCreateServiceAllowsFreeService(t *testing.T) {
	item := validService()
	item.Price = 0

	if _, err := NewCatalogService(&stubServiceRepo{}).CreateService(context.Background(), item); err != nil {
		t.Fatalf("бесплатная услуга должна приниматься, получено %v", err)
	}
}

func TestCreateServiceAllowsExactlyOneDay(t *testing.T) {
	item := validService()
	item.Duration = 24 * 60

	if _, err := NewCatalogService(&stubServiceRepo{}).CreateService(context.Background(), item); err != nil {
		t.Fatalf("сутки — допустимая длительность, получено %v", err)
	}
}

func TestUpdateServiceValidates(t *testing.T) {
	cases := map[string]models.Service{
		"empty name":          {Name: "", Duration: 30},
		"name too long":       {Name: strings.Repeat("a", 201), Duration: 30},
		"description long":    {Name: "Haircut", Description: strings.Repeat("a", 501), Duration: 30},
		"zero duration":       {Name: "Haircut", Duration: 0},
		"duration over a day": {Name: "Haircut", Duration: 24*60 + 1},
		"negative price":      {Name: "Haircut", Duration: 30, Price: -1},
	}

	for name, item := range cases {
		t.Run(name, func(t *testing.T) {
			repo := &stubServiceRepo{}

			err := NewCatalogService(repo).UpdateService(context.Background(), 1, item)

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

func TestUpdateServiceAllowsABodyWithoutCompany(t *testing.T) {
	repo := &stubServiceRepo{}

	err := NewCatalogService(repo).UpdateService(context.Background(), 6, models.Service{ID: 99, Name: "Haircut", Duration: 30})
	if err != nil {
		t.Fatalf("UpdateService: %v", err)
	}

	if repo.updated.ID != 6 {
		t.Errorf("ID = %d, want the path id 6", repo.updated.ID)
	}
}

func TestUpdateServicePassesRepositoryErrors(t *testing.T) {
	repo := &stubServiceRepo{err: models.ErrNotFound}

	err := NewCatalogService(repo).UpdateService(context.Background(), 1, models.Service{Name: "Haircut", Duration: 30})

	if !errors.Is(err, models.ErrNotFound) {
		t.Fatalf("got %v, want models.ErrNotFound", err)
	}
}
