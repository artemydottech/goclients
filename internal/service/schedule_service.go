package service

import (
	"database/sql"
	"errors"

	"github.com/artemydottech/goclients/internal/models"
)

type ScheduleRepo interface {
	SetEmployeeSchedule(employeeID int, days []models.WorkingDay) error
	GetEmployeeSchedule(employeeID int) ([]models.WorkingDay, error)
	GetWorkingDay(employeeID, weekday int) (models.WorkingDay, error)
}

type ScheduleService struct {
	repo      ScheduleRepo
	employees EmployeeLookup
}

func NewScheduleService(repo ScheduleRepo, employees EmployeeLookup) *ScheduleService {
	return &ScheduleService{repo: repo, employees: employees}
}

func (s *ScheduleService) SetEmployeeSchedule(employeeID int, days []models.WorkingDay) error {
	if _, err := s.employees.GetEmployeeById(employeeID); err != nil {
		return err
	}

	seen := make(map[int]struct{}, len(days))
	for i := range days {
		days[i].EmployeeID = employeeID

		if err := days[i].Validate(); err != nil {
			return err
		}

		if _, duplicate := seen[days[i].Weekday]; duplicate {
			return models.Invalid("День недели %d указан дважды!", days[i].Weekday)
		}
		seen[days[i].Weekday] = struct{}{}
	}

	return s.repo.SetEmployeeSchedule(employeeID, days)
}

func (s *ScheduleService) GetEmployeeSchedule(employeeID int) ([]models.WorkingDay, error) {
	return s.repo.GetEmployeeSchedule(employeeID)
}

func (s *ScheduleService) WorkingDay(employeeID, weekday int) (models.WorkingDay, bool, error) {
	day, err := s.repo.GetWorkingDay(employeeID, weekday)
	if errors.Is(err, sql.ErrNoRows) {
		return models.WorkingDay{}, false, nil
	}
	if err != nil {
		return models.WorkingDay{}, false, err
	}

	return day, true, nil
}
