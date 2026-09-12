package repository

import (
	"database/sql"
	"time"

	"github.com/artemydottech/goclients/internal/models"
)

// timeLayout — записи хранятся текстом в UTC: SQLite не различает зоны, а
// сравнение строк в этом формате совпадает со сравнением моментов времени.
const timeLayout = time.RFC3339

type AppointmentRepository struct {
	db *sql.DB
}

func NewAppointmentRepository(db *sql.DB) *AppointmentRepository {
	return &AppointmentRepository{db: db}
}

func (r *AppointmentRepository) Create(a models.Appointment) (int64, error) {
	res, err := r.db.Exec(`
        INSERT INTO appointments
            (company_id, client_id, employee_id, service_id, starts_at, ends_at, status, comment)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		a.CompanyID, a.ClientID, a.EmployeeID, a.ServiceID,
		a.StartsAt.UTC().Format(timeLayout), a.EndsAt.UTC().Format(timeLayout),
		string(a.Status), a.Comment,
	)
	if err != nil {
		return 0, err
	}

	return res.LastInsertId()
}

func (r *AppointmentRepository) GetAllAppointments() ([]models.Appointment, error) {
	return r.query(`
        SELECT id, company_id, client_id, employee_id, service_id, starts_at, ends_at, status, comment
        FROM appointments ORDER BY starts_at`)
}

func (r *AppointmentRepository) GetAppointmentsByEmployee(employeeID int, from, to time.Time) ([]models.Appointment, error) {
	return r.query(`
        SELECT id, company_id, client_id, employee_id, service_id, starts_at, ends_at, status, comment
        FROM appointments
        WHERE employee_id = ? AND starts_at < ? AND ends_at > ?
        ORDER BY starts_at`,
		employeeID, to.UTC().Format(timeLayout), from.UTC().Format(timeLayout))
}

func (r *AppointmentRepository) GetAppointmentsByClient(clientID int) ([]models.Appointment, error) {
	return r.query(`
        SELECT id, company_id, client_id, employee_id, service_id, starts_at, ends_at, status, comment
        FROM appointments WHERE client_id = ? ORDER BY starts_at`, clientID)
}

func (r *AppointmentRepository) query(query string, args ...any) ([]models.Appointment, error) {
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	appointments := []models.Appointment{}
	for rows.Next() {
		a, err := scanAppointment(rows)
		if err != nil {
			return nil, err
		}
		appointments = append(appointments, a)
	}

	return appointments, rows.Err()
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanAppointment(row rowScanner) (models.Appointment, error) {
	var (
		a        models.Appointment
		startsAt string
		endsAt   string
		status   string
	)

	err := row.Scan(&a.ID, &a.CompanyID, &a.ClientID, &a.EmployeeID, &a.ServiceID,
		&startsAt, &endsAt, &status, &a.Comment)
	if err != nil {
		return models.Appointment{}, err
	}

	if a.StartsAt, err = time.Parse(timeLayout, startsAt); err != nil {
		return models.Appointment{}, err
	}
	if a.EndsAt, err = time.Parse(timeLayout, endsAt); err != nil {
		return models.Appointment{}, err
	}
	a.Status = models.AppointmentStatus(status)

	return a, nil
}

func (r *AppointmentRepository) GetAppointmentById(id int) (models.Appointment, error) {
	row := r.db.QueryRow(`
        SELECT id, company_id, client_id, employee_id, service_id, starts_at, ends_at, status, comment
        FROM appointments WHERE id = ?`, id)

	return scanAppointment(row)
}

// HasOverlap ищет чужую запись, накрывающую интервал того же мастера.
// Границы касаются, а не пересекаются: запись 10:00–11:00 не мешает 11:00–12:00.
func (r *AppointmentRepository) HasOverlap(employeeID int, from, to time.Time) (bool, error) {
	var exists int

	err := r.db.QueryRow(`
        SELECT 1 FROM appointments
        WHERE employee_id = ? AND status != ?
          AND starts_at < ? AND ends_at > ?
        LIMIT 1`,
		employeeID, string(models.AppointmentCancelled),
		to.UTC().Format(timeLayout), from.UTC().Format(timeLayout),
	).Scan(&exists)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return true, nil
}

func (r *AppointmentRepository) UpdateStatus(id int, status models.AppointmentStatus) error {
	res, err := r.db.Exec("UPDATE appointments SET status = ? WHERE id = ?", string(status), id)
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

func (r *AppointmentRepository) DeleteAppointmentById(id int) error {
	res, err := r.db.Exec("DELETE FROM appointments WHERE id = ?", id)
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
