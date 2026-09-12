package service

import (
	"errors"
	"strings"
	"testing"

	"github.com/artemydottech/goclients/internal/models"
)

type stubServiceRepo struct {
	called bool
}

func (r *stubServiceRepo) Create(models.Service) (int64, error) {
	r.called = true
	return 9, nil
}

func (r *stubServiceRepo) GetAllServices() ([]models.Service, error) { return nil, nil }

func (r *stubServiceRepo) GetServicesByCompany(int) ([]models.Service, error) { return nil, nil }

func (r *stubServiceRepo) GetServiceById(int) (models.Service, error) {
	return models.Service{}, nil
}

func (r *stubServiceRepo) DeleteServiceById(int) error { return nil }

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

			_, err := NewCatalogService(repo).CreateService(item)

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

	id, err := NewCatalogService(repo).CreateService(validService())
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

	if _, err := NewCatalogService(&stubServiceRepo{}).CreateService(item); err != nil {
		t.Fatalf("бесплатная услуга должна приниматься, получено %v", err)
	}
}

func TestCreateServiceAllowsExactlyOneDay(t *testing.T) {
	item := validService()
	item.Duration = 24 * 60

	if _, err := NewCatalogService(&stubServiceRepo{}).CreateService(item); err != nil {
		t.Fatalf("сутки — допустимая длительность, получено %v", err)
	}
}
