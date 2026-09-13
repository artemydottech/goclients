package service

import (
	"errors"
	"testing"

	"github.com/artemydottech/goclients/internal/models"
)

type stubScheduleRepo struct {
	saved  []models.WorkingDay
	called bool
}

func (r *stubScheduleRepo) SetEmployeeSchedule(_ int, days []models.WorkingDay) error {
	r.called = true
	r.saved = days
	return nil
}

func (r *stubScheduleRepo) GetEmployeeSchedule(int) ([]models.WorkingDay, error) {
	return nil, nil
}

func (r *stubScheduleRepo) GetWorkingDay(int, int) (models.WorkingDay, error) {
	return models.WorkingDay{}, nil
}

func newScheduleService(repo *stubScheduleRepo) *ScheduleService {
	return NewScheduleService(repo, stubEmployeeLookup{employee: models.Employee{ID: 1, CompanyID: 10}})
}

func TestSetEmployeeScheduleStampsTheEmployee(t *testing.T) {
	repo := &stubScheduleRepo{}

	days := []models.WorkingDay{{Weekday: 1, StartsAt: "10:00", EndsAt: "20:00"}}
	if err := newScheduleService(repo).SetEmployeeSchedule(7, days); err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}

	if repo.saved[0].EmployeeID != 7 {
		t.Errorf("employee_id %d, ожидался 7", repo.saved[0].EmployeeID)
	}
}

func TestSetEmployeeScheduleValidation(t *testing.T) {
	cases := map[string][]models.WorkingDay{
		"день вне недели":     {{Weekday: 7, StartsAt: "10:00", EndsAt: "20:00"}},
		"отрицательный день":  {{Weekday: -1, StartsAt: "10:00", EndsAt: "20:00"}},
		"конец раньше начала": {{Weekday: 1, StartsAt: "20:00", EndsAt: "10:00"}},
		"конец равен началу":  {{Weekday: 1, StartsAt: "10:00", EndsAt: "10:00"}},
		"кривой формат":       {{Weekday: 1, StartsAt: "10-00", EndsAt: "20:00"}},
		"часы больше 23":      {{Weekday: 1, StartsAt: "25:00", EndsAt: "26:00"}},
		"минуты больше 59":    {{Weekday: 1, StartsAt: "10:75", EndsAt: "20:00"}},
		"день дважды": {
			{Weekday: 1, StartsAt: "10:00", EndsAt: "14:00"},
			{Weekday: 1, StartsAt: "15:00", EndsAt: "20:00"},
		},
	}

	for name, days := range cases {
		t.Run(name, func(t *testing.T) {
			repo := &stubScheduleRepo{}

			err := newScheduleService(repo).SetEmployeeSchedule(1, days)

			var validationErr models.ValidationError
			if !errors.As(err, &validationErr) {
				t.Fatalf("ожидалась ValidationError, получено %v", err)
			}

			if repo.called {
				t.Error("график не должен сохраняться при невалидных данных")
			}
		})
	}
}

func TestParseAndFormatDayTimeRoundTrip(t *testing.T) {
	for _, value := range []string{"00:00", "09:05", "23:59"} {
		minutes, err := models.ParseDayTime(value)
		if err != nil {
			t.Fatalf("ParseDayTime(%q): %v", value, err)
		}

		if got := models.FormatDayTime(minutes); got != value {
			t.Errorf("FormatDayTime(ParseDayTime(%q)) = %q", value, got)
		}
	}
}

func TestSetEmployeeScheduleBreakValidation(t *testing.T) {
	day := func(breakFrom, breakTo string) []models.WorkingDay {
		return []models.WorkingDay{{
			Weekday: 1, StartsAt: "10:00", EndsAt: "20:00",
			BreakStartsAt: breakFrom, BreakEndsAt: breakTo,
		}}
	}

	cases := map[string][]models.WorkingDay{
		"только начало":           day("13:00", ""),
		"только конец":            day("", "14:00"),
		"конец раньше начала":     day("14:00", "13:00"),
		"начинается до открытия":  day("09:00", "11:00"),
		"заканчивается после дня": day("19:30", "20:30"),
		"кривой формат":           day("13-00", "14:00"),
	}

	for name, days := range cases {
		t.Run(name, func(t *testing.T) {
			repo := &stubScheduleRepo{}

			err := newScheduleService(repo).SetEmployeeSchedule(1, days)

			var validationErr models.ValidationError
			if !errors.As(err, &validationErr) {
				t.Fatalf("ожидалась ValidationError, получено %v", err)
			}
			if repo.called {
				t.Error("график не должен сохраняться при невалидном перерыве")
			}
		})
	}
}

func TestSetEmployeeScheduleAcceptsBreakOnTheEdges(t *testing.T) {
	days := []models.WorkingDay{{
		Weekday: 1, StartsAt: "10:00", EndsAt: "20:00",
		BreakStartsAt: "10:00", BreakEndsAt: "11:00",
	}}

	if err := newScheduleService(&stubScheduleRepo{}).SetEmployeeSchedule(1, days); err != nil {
		t.Fatalf("перерыв с самого открытия допустим, получено %v", err)
	}
}
