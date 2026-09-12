package repository

import (
	"database/sql"

	"github.com/artemydottech/goclients/internal/models"
)

type ScheduleRepository struct {
	db *sql.DB
}

func NewScheduleRepository(db *sql.DB) *ScheduleRepository {
	return &ScheduleRepository{db: db}
}

// SetEmployeeSchedule заменяет недельный график целиком — половина расписания
// значила бы, что мастер внезапно не работает во вторник.
func (r *ScheduleRepository) SetEmployeeSchedule(employeeID int, days []models.WorkingDay) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec("DELETE FROM employee_schedules WHERE employee_id = ?", employeeID); err != nil {
		return err
	}

	statement, err := tx.Prepare(`
        INSERT INTO employee_schedules (employee_id, weekday, starts_at, ends_at)
        VALUES (?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	defer statement.Close()

	for _, day := range days {
		if _, err := statement.Exec(employeeID, day.Weekday, day.StartsAt, day.EndsAt); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *ScheduleRepository) GetEmployeeSchedule(employeeID int) ([]models.WorkingDay, error) {
	rows, err := r.db.Query(`
        SELECT employee_id, weekday, starts_at, ends_at
        FROM employee_schedules WHERE employee_id = ? ORDER BY weekday`, employeeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	days := []models.WorkingDay{}
	for rows.Next() {
		var day models.WorkingDay
		if err := rows.Scan(&day.EmployeeID, &day.Weekday, &day.StartsAt, &day.EndsAt); err != nil {
			return nil, err
		}
		days = append(days, day)
	}

	return days, rows.Err()
}

func (r *ScheduleRepository) GetWorkingDay(employeeID, weekday int) (models.WorkingDay, error) {
	var day models.WorkingDay

	err := r.db.QueryRow(`
        SELECT employee_id, weekday, starts_at, ends_at
        FROM employee_schedules WHERE employee_id = ? AND weekday = ?`, employeeID, weekday).
		Scan(&day.EmployeeID, &day.Weekday, &day.StartsAt, &day.EndsAt)
	if err != nil {
		return models.WorkingDay{}, err
	}

	return day, nil
}
