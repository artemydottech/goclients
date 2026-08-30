package service

import (
	"errors"
	"strings"
	"testing"

	"github.com/artemydottech/goclients/internal/models"
)

type stubEmployeeRepo struct {
	called bool
}

func (r *stubEmployeeRepo) Create(models.Employee) (int64, error) {
	r.called = true
	return 5, nil
}

func (r *stubEmployeeRepo) GetAllEmployees() ([]models.Employee, error) { return nil, nil }

func (r *stubEmployeeRepo) GetEmployeeById(int) (models.Employee, error) {
	return models.Employee{}, nil
}

func (r *stubEmployeeRepo) DeleteEmployeeById(int) error { return nil }

func TestCreateEmployeeValidation(t *testing.T) {
	cases := map[string]models.Employee{
		"пустое имя":          {Name: "", Surname: "Зверев"},
		"пустая фамилия":      {Name: "Артемий", Surname: ""},
		"имя длиннее 200":     {Name: strings.Repeat("a", 201), Surname: "Зверев"},
		"фамилия длиннее 200": {Name: "Артемий", Surname: strings.Repeat("a", 201)},
		"должность длиннее 500": {
			Name: "Артемий", Surname: "Зверев", Position: strings.Repeat("a", 501),
		},
		"аватар длиннее 500": {
			Name: "Артемий", Surname: "Зверев", Avatar: strings.Repeat("a", 501),
		},
	}

	for name, employee := range cases {
		t.Run(name, func(t *testing.T) {
			repo := &stubEmployeeRepo{}

			_, err := NewEmployeeService(repo).CreateEmployee(employee)

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

// Лимиты свободных полей тоже считаются в рунах: 500 кириллических символов
// занимают 1000 байт и обязаны проходить.
func TestCreateEmployeeCountsRunesInPosition(t *testing.T) {
	employee := models.Employee{
		Name:     "Артемий",
		Surname:  "Зверев",
		Position: strings.Repeat("я", 500),
	}

	if _, err := NewEmployeeService(&stubEmployeeRepo{}).CreateEmployee(employee); err != nil {
		t.Fatalf("ожидался успех, получена ошибка %v", err)
	}
}

func TestCreateEmployeeAcceptsValidInput(t *testing.T) {
	employee := models.Employee{
		CompanyID: 1,
		Name:      "Артемий",
		Surname:   "Зверев",
		Position:  "Разработчик",
	}

	id, err := NewEmployeeService(&stubEmployeeRepo{}).CreateEmployee(employee)
	if err != nil {
		t.Fatalf("ожидался успех, получена ошибка %v", err)
	}

	if id != 5 {
		t.Errorf("ожидался id 5, получен %d", id)
	}
}
