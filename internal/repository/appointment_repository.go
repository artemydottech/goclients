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

// CreateIfFree проверяет пересечение и вставляет запись в одной транзакции.
// Раздельные HasOverlap и Create пропускали два параллельных запроса на один
// слот: оба видели свободное время и оба вставляли.
func (r *AppointmentRepository) CreateIfFree(a models.Appointment) (int64, bool, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return 0, false, err
	}
	defer tx.Rollback()

	busy, err := hasOverlap(tx, a.EmployeeID, a.StartsAt, a.EndsAt, 0)
	if err != nil {
		return 0, false, err
	}
	if busy {
		return 0, true, nil
	}

	id, err := insertAppointment(tx, a)
	if err != nil {
		return 0, false, err
	}

	return id, false, tx.Commit()
}

type execer interface {
	Exec(query string, args ...any) (sql.Result, error)
}

func (r *AppointmentRepository) Create(a models.Appointment) (int64, error) {
	return insertAppointment(r.db, a)
}

func insertAppointment(e execer, a models.Appointment) (int64, error) {
	res, err := e.Exec(`
        INSERT INTO appointments
            (company_id, client_id, employee_id, service_id, starts_at, ends_at, status, comment, price)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		a.CompanyID, a.ClientID, a.EmployeeID, a.ServiceID,
		a.StartsAt.UTC().Format(timeLayout), a.EndsAt.UTC().Format(timeLayout),
		string(a.Status), a.Comment, a.Price,
	)
	if err != nil {
		return 0, err
	}

	return res.LastInsertId()
}

func (r *AppointmentRepository) GetAllAppointments() ([]models.Appointment, error) {
	return r.query(`
        SELECT id, company_id, client_id, employee_id, service_id, starts_at, ends_at, status, comment, price
        FROM appointments ORDER BY starts_at`)
}

func (r *AppointmentRepository) GetAppointmentsByEmployee(employeeID int, from, to time.Time) ([]models.Appointment, error) {
	return r.query(`
        SELECT id, company_id, client_id, employee_id, service_id, starts_at, ends_at, status, comment, price
        FROM appointments
        WHERE employee_id = ? AND starts_at < ? AND ends_at > ?
        ORDER BY starts_at`,
		employeeID, to.UTC().Format(timeLayout), from.UTC().Format(timeLayout))
}

func (r *AppointmentRepository) GetAppointmentsByClient(clientID int) ([]models.Appointment, error) {
	return r.query(`
        SELECT id, company_id, client_id, employee_id, service_id, starts_at, ends_at, status, comment, price
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
		&startsAt, &endsAt, &status, &a.Comment, &a.Price)
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
        SELECT id, company_id, client_id, employee_id, service_id, starts_at, ends_at, status, comment, price
        FROM appointments WHERE id = ?`, id)

	return scanAppointment(row)
}

// HasOverlap ищет чужую запись, накрывающую интервал того же мастера.
// Границы касаются, а не пересекаются: запись 10:00–11:00 не мешает 11:00–12:00.
func (r *AppointmentRepository) HasOverlap(employeeID int, from, to time.Time) (bool, error) {
	return hasOverlap(r.db, employeeID, from, to, 0)
}

// MoveIfFree переносит запись, если новое время свободно. Сама переносимая
// запись в проверку не попадает — сдвиг на полчаса внутри своего же интервала
// не конфликт.
func (r *AppointmentRepository) MoveIfFree(id int, a models.Appointment) (bool, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return false, err
	}
	defer tx.Rollback()

	busy, err := hasOverlap(tx, a.EmployeeID, a.StartsAt, a.EndsAt, id)
	if err != nil {
		return false, err
	}
	if busy {
		return true, nil
	}

	res, err := tx.Exec(
		"UPDATE appointments SET employee_id = ?, starts_at = ?, ends_at = ? WHERE id = ?",
		a.EmployeeID, a.StartsAt.UTC().Format(timeLayout), a.EndsAt.UTC().Format(timeLayout), id,
	)
	if err != nil {
		return false, err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	if rows == 0 {
		return false, sql.ErrNoRows
	}

	return false, tx.Commit()
}

type querier interface {
	QueryRow(query string, args ...any) *sql.Row
}

func hasOverlap(q querier, employeeID int, from, to time.Time, excludeID int) (bool, error) {
	var exists int

	err := q.QueryRow(`
        SELECT 1 FROM appointments
        WHERE employee_id = ? AND status != ? AND id != ?
          AND starts_at < ? AND ends_at > ?
        LIMIT 1`,
		employeeID, string(models.AppointmentCancelled), excludeID,
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

func (r *AppointmentRepository) GetClientStats(clientID int, now time.Time) (models.ClientStats, error) {
	stats := models.ClientStats{ClientID: clientID}

	var (
		lastVisit sql.NullString
		nextVisit sql.NullString
		nowText   = now.UTC().Format(timeLayout)
	)

	err := r.db.QueryRow(`
        SELECT
            COUNT(*),
            COALESCE(SUM(CASE WHEN status = ? THEN 1 ELSE 0 END), 0),
            COALESCE(SUM(CASE WHEN status = ? THEN 1 ELSE 0 END), 0),
            COALESCE(SUM(CASE WHEN status = ? THEN 1 ELSE 0 END), 0),
            COALESCE(SUM(CASE WHEN status IN (?, ?) AND starts_at > ? THEN 1 ELSE 0 END), 0),
            COALESCE(SUM(CASE WHEN status = ? THEN price ELSE 0 END), 0),
            MAX(CASE WHEN status = ? THEN starts_at END),
            MIN(CASE WHEN status IN (?, ?) AND starts_at > ? THEN starts_at END)
        FROM appointments WHERE client_id = ?`,
		string(models.AppointmentCompleted),
		string(models.AppointmentNoShow),
		string(models.AppointmentCancelled),
		string(models.AppointmentPending), string(models.AppointmentConfirmed), nowText,
		string(models.AppointmentCompleted),
		string(models.AppointmentCompleted),
		string(models.AppointmentPending), string(models.AppointmentConfirmed), nowText,
		clientID,
	).Scan(
		&stats.Appointments, &stats.CompletedVisits, &stats.NoShows, &stats.Cancelled,
		&stats.Upcoming, &stats.TotalSpent, &lastVisit, &nextVisit,
	)
	if err != nil {
		return models.ClientStats{}, err
	}

	if stats.LastVisitAt, err = parseOptionalTime(lastVisit); err != nil {
		return models.ClientStats{}, err
	}
	if stats.NextVisitAt, err = parseOptionalTime(nextVisit); err != nil {
		return models.ClientStats{}, err
	}

	return stats, nil
}

func parseOptionalTime(value sql.NullString) (*time.Time, error) {
	if !value.Valid {
		return nil, nil
	}

	parsed, err := time.Parse(timeLayout, value.String)
	if err != nil {
		return nil, err
	}

	return &parsed, nil
}
