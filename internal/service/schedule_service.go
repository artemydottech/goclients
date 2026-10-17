package service

import (
	"context"
	"errors"

	"github.com/artemydottech/goclients/internal/models"
)

type ScheduleRepo interface {
	SetEmployeeSchedule(ctx context.Context, employeeID int, days []models.WorkingDay) error
	GetEmployeeSchedule(ctx context.Context, employeeID int) ([]models.WorkingDay, error)
	GetWorkingDay(ctx context.Context, employeeID, weekday int) (models.WorkingDay, error)
}

type ScheduleService struct {
	repo      ScheduleRepo
	employees EmployeeLookup
}

func NewScheduleService(repo ScheduleRepo, employees EmployeeLookup) *ScheduleService {
	return &ScheduleService{repo: repo, employees: employees}
}

func (s *ScheduleService) SetEmployeeSchedule(ctx context.Context, employeeID int, days []models.WorkingDay) error {
	if _, err := s.employees.GetEmployeeById(ctx, employeeID); err != nil {
		return err
	}

	seen := make(map[int]struct{}, len(days))
	for i := range days {
		days[i].EmployeeID = employeeID

		if err := days[i].Validate(); err != nil {
			return err
		}

		if _, duplicate := seen[days[i].Weekday]; duplicate {
			return models.Invalid("weekday %d is listed twice", days[i].Weekday)
		}
		seen[days[i].Weekday] = struct{}{}
	}

	return s.repo.SetEmployeeSchedule(ctx, employeeID, days)
}

func (s *ScheduleService) GetEmployeeSchedule(ctx context.Context, employeeID int) ([]models.WorkingDay, error) {
	return s.repo.GetEmployeeSchedule(ctx, employeeID)
}

func (s *ScheduleService) WorkingDay(employeeID, weekday int) (models.WorkingDay, bool, error) {
	day, err := s.repo.GetWorkingDay(context.TODO(), employeeID, weekday)
	if errors.Is(err, models.ErrNotFound) {
		return models.WorkingDay{}, false, nil
	}
	if err != nil {
		return models.WorkingDay{}, false, err
	}

	return day, true, nil
}
