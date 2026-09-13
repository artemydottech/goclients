package repository

import (
	"database/sql"
	"time"

	"github.com/artemydottech/goclients/internal/models"
)

type TimeOffRepository struct {
	db *sql.DB
}

func NewTimeOffRepository(db *sql.DB) *TimeOffRepository {
	return &TimeOffRepository{db: db}
}

func (r *TimeOffRepository) Create(t models.TimeOff) (int64, error) {
	res, err := r.db.Exec(
		"INSERT INTO employee_time_off (employee_id, starts_at, ends_at, reason) VALUES (?, ?, ?, ?)",
		t.EmployeeID, t.StartsAt.UTC().Format(timeLayout), t.EndsAt.UTC().Format(timeLayout), t.Reason,
	)
	if err != nil {
		return 0, err
	}

	return res.LastInsertId()
}

func (r *TimeOffRepository) GetTimeOffByEmployee(employeeID int) ([]models.TimeOff, error) {
	return r.query(`
        SELECT id, employee_id, starts_at, ends_at, reason
        FROM employee_time_off WHERE employee_id = ? ORDER BY starts_at`, employeeID)
}

func (r *TimeOffRepository) GetTimeOffInRange(employeeID int, from, to time.Time) ([]models.TimeOff, error) {
	return r.query(`
        SELECT id, employee_id, starts_at, ends_at, reason
        FROM employee_time_off
        WHERE employee_id = ? AND starts_at < ? AND ends_at > ?
        ORDER BY starts_at`,
		employeeID, to.UTC().Format(timeLayout), from.UTC().Format(timeLayout))
}

func (r *TimeOffRepository) query(query string, args ...any) ([]models.TimeOff, error) {
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	periods := []models.TimeOff{}
	for rows.Next() {
		var (
			t        models.TimeOff
			startsAt string
			endsAt   string
		)

		if err := rows.Scan(&t.ID, &t.EmployeeID, &startsAt, &endsAt, &t.Reason); err != nil {
			return nil, err
		}
		if t.StartsAt, err = time.Parse(timeLayout, startsAt); err != nil {
			return nil, err
		}
		if t.EndsAt, err = time.Parse(timeLayout, endsAt); err != nil {
			return nil, err
		}

		periods = append(periods, t)
	}

	return periods, rows.Err()
}

func (r *TimeOffRepository) DeleteTimeOffById(id int) error {
	res, err := r.db.Exec("DELETE FROM employee_time_off WHERE id = ?", id)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return sql.ErrNoRows
	}

	return nil
}
