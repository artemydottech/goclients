package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/artemydottech/goclients/internal/models"
)

type stubEmployeeRepo struct {
	updated models.Employee
	err     error
	called  bool
}

func (r *stubEmployeeRepo) Create(context.Context, models.Employee) (int64, error) {
	r.called = true
	return 5, nil
}

func (r *stubEmployeeRepo) GetAllEmployees(context.Context, models.Page) ([]models.Employee, error) {
	return nil, nil
}

func (r *stubEmployeeRepo) GetEmployeeById(context.Context, int) (models.Employee, error) {
	return models.Employee{}, nil
}

func (r *stubEmployeeRepo) UpdateEmployee(_ context.Context, e models.Employee) error {
	r.called = true
	r.updated = e
	return r.err
}

func (r *stubEmployeeRepo) DeleteEmployeeById(context.Context, int) error { return nil }

func TestCreateEmployeeValidation(t *testing.T) {
	cases := map[string]models.Employee{
		"blank required field 0": {Name: " \t\n", Surname: "Dubov"},
		"blank required field 1": {Name: "Daniil", Surname: " \t\n"},
		"пустое имя":             {Name: "", Surname: "Dubov"},
		"пустая фамилия":         {Name: "Daniil", Surname: ""},
		"имя длиннее 200":        {Name: strings.Repeat("a", 201), Surname: "Dubov"},
		"фамилия длиннее 200":    {Name: "Daniil", Surname: strings.Repeat("a", 201)},
		"должность длиннее 500": {
			Name: "Daniil", Surname: "Dubov", Position: strings.Repeat("a", 501),
		},
		"аватар длиннее 500": {
			Name: "Daniil", Surname: "Dubov", Avatar: strings.Repeat("a", 501),
		},
	}

	for name, employee := range cases {
		t.Run(name, func(t *testing.T) {
			repo := &stubEmployeeRepo{}

			employee.CompanyID = 1
			_, err := NewEmployeeService(repo).CreateEmployee(context.Background(), employee)

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

func TestCreateEmployeeCountsRunesInPosition(t *testing.T) {
	employee := models.Employee{
		CompanyID: 1,
		Name:      "Daniil",
		Surname:   "Dubov",
		Position:  strings.Repeat("я", 500),
	}

	if _, err := NewEmployeeService(&stubEmployeeRepo{}).CreateEmployee(context.Background(), employee); err != nil {
		t.Fatalf("ожидался успех, получена ошибка %v", err)
	}
}

func TestCreateEmployeeAcceptsValidInput(t *testing.T) {
	employee := models.Employee{
		CompanyID: 1,
		Name:      "Daniil",
		Surname:   "Dubov",
		Position:  "Разработчик",
	}

	id, err := NewEmployeeService(&stubEmployeeRepo{}).CreateEmployee(context.Background(), employee)
	if err != nil {
		t.Fatalf("ожидался успех, получена ошибка %v", err)
	}

	if id != 5 {
		t.Errorf("ожидался id 5, получен %d", id)
	}
}

func TestUpdateEmployeeValidates(t *testing.T) {
	cases := map[string]models.Employee{
		"blank required field 0": {Name: " \t\n", Surname: "Dubov"},
		"blank required field 1": {Name: "Daniil", Surname: " \t\n"},
		"empty name":             {Name: "", Surname: "Dubov"},
		"empty surname":          {Name: "Daniil", Surname: ""},
		"name too long":          {Name: strings.Repeat("a", 201), Surname: "Dubov"},
		"position too long":      {Name: "Daniil", Surname: "Dubov", Position: strings.Repeat("a", 501)},
	}

	for name, employee := range cases {
		t.Run(name, func(t *testing.T) {
			repo := &stubEmployeeRepo{}

			err := NewEmployeeService(repo).UpdateEmployee(context.Background(), 1, employee)

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

func TestUpdateEmployeeUsesThePathID(t *testing.T) {
	repo := &stubEmployeeRepo{}

	err := NewEmployeeService(repo).UpdateEmployee(context.Background(), 4, models.Employee{ID: 99, Name: "Daniil", Surname: "Dubov"})
	if err != nil {
		t.Fatalf("UpdateEmployee: %v", err)
	}

	if repo.updated.ID != 4 {
		t.Errorf("ID = %d, want the path id 4", repo.updated.ID)
	}
}

func TestUpdateEmployeePassesRepositoryErrors(t *testing.T) {
	repo := &stubEmployeeRepo{err: models.ErrNotFound}

	err := NewEmployeeService(repo).UpdateEmployee(context.Background(), 1, models.Employee{Name: "Daniil", Surname: "Dubov"})

	if !errors.Is(err, models.ErrNotFound) {
		t.Fatalf("got %v, want models.ErrNotFound", err)
	}
}

func (r *stubEmployeeRepo) GetEmployeesByCompany(context.Context, int, models.Page) ([]models.Employee, error) {
	r.called = true
	return nil, r.err
}

func TestEmployeeDirectoryRejectsInvalidScopeAndPage(t *testing.T) {
	for _, tt := range []struct {
		id   int
		page models.Page
	}{{0, models.Page{}}, {-1, models.Page{}}, {1, models.Page{Limit: 201}}, {1, models.Page{Offset: -1}}} {
		repo := &stubEmployeeRepo{}
		if _, err := NewEmployeeService(repo).GetEmployeesByCompany(context.Background(), tt.id, tt.page); err == nil || repo.called {
			t.Fatalf("invalid input reached storage: %+v", tt)
		}
	}
	repo := &stubEmployeeRepo{err: errors.New("storage")}
	if _, err := NewEmployeeService(repo).GetEmployeesByCompany(context.Background(), 1, models.Page{}); !errors.Is(err, repo.err) {
		t.Fatalf("error: %v", err)
	}
}
