package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/artemydottech/goclients/internal/models"
)

const timeLayout = time.RFC3339

type AppointmentRepository struct {
	db *sql.DB
}

func NewAppointmentRepository(db *sql.DB) *AppointmentRepository {
	return &AppointmentRepository{db: db}
}

func (r *AppointmentRepository) CreateIfFree(ctx context.Context, a models.Appointment) (int64, bool, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, false, err
	}
	defer tx.Rollback()

	busy, err := hasOverlap(ctx, tx, a.EmployeeID, a.StartsAt, a.EndsAt, 0)
	if err != nil {
		return 0, false, err
	}
	if busy {
		return 0, true, nil
	}

	id, err := insertAppointment(ctx, tx, a)
	if err != nil {
		return 0, false, err
	}

	return id, false, tx.Commit()
}

type execer interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

func (r *AppointmentRepository) Create(ctx context.Context, a models.Appointment) (int64, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	id, err := insertAppointment(ctx, tx, a)
	if err != nil {
		return 0, err
	}
	return id, tx.Commit()
}

func insertAppointment(ctx context.Context, e execer, a models.Appointment) (int64, error) {
	res, err := e.ExecContext(ctx, `
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

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	a.ID = int(id)
	a.StartsAt = a.StartsAt.UTC().Truncate(time.Second)
	a.EndsAt = a.EndsAt.UTC().Truncate(time.Second)
	if err := recordAppointmentEvent(ctx, e, "created", nil, a); err != nil {
		return 0, err
	}
	return id, nil
}

func (r *AppointmentRepository) GetAllAppointments(ctx context.Context) ([]models.Appointment, error) {
	return r.query(ctx, `
        SELECT id, company_id, client_id, employee_id, service_id, starts_at, ends_at, status, comment, price
        FROM appointments ORDER BY starts_at`)
}

func (r *AppointmentRepository) GetAppointmentsByEmployee(ctx context.Context, employeeID int, from, to time.Time) ([]models.Appointment, error) {
	return r.query(ctx, `
        SELECT id, company_id, client_id, employee_id, service_id, starts_at, ends_at, status, comment, price
        FROM appointments
        WHERE employee_id = ? AND starts_at < ? AND ends_at > ?
        ORDER BY starts_at`,
		employeeID, to.UTC().Format(timeLayout), from.UTC().Format(timeLayout))
}

func (r *AppointmentRepository) GetAppointmentsByClient(ctx context.Context, clientID int) ([]models.Appointment, error) {
	return r.query(ctx, `
        SELECT id, company_id, client_id, employee_id, service_id, starts_at, ends_at, status, comment, price
        FROM appointments WHERE client_id = ? ORDER BY starts_at`, clientID)
}

func (r *AppointmentRepository) query(ctx context.Context, query string, args ...any) ([]models.Appointment, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
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
		return models.Appointment{}, mapNoRows(err)
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

func (r *AppointmentRepository) GetAppointmentById(ctx context.Context, id int) (models.Appointment, error) {
	row := r.db.QueryRowContext(ctx, `
        SELECT id, company_id, client_id, employee_id, service_id, starts_at, ends_at, status, comment, price
        FROM appointments WHERE id = ?`, id)

	return scanAppointment(row)
}

func (r *AppointmentRepository) HasOverlap(ctx context.Context, employeeID int, from, to time.Time) (bool, error) {
	return hasOverlap(ctx, r.db, employeeID, from, to, 0)
}

func (r *AppointmentRepository) MoveIfFree(ctx context.Context, id int, a models.Appointment) (bool, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return false, err
	}
	defer tx.Rollback()
	before, err := appointmentIn(ctx, tx, id)
	if err != nil {
		return false, err
	}

	busy, err := hasOverlap(ctx, tx, a.EmployeeID, a.StartsAt, a.EndsAt, id)
	if err != nil {
		return false, err
	}
	if busy {
		return true, nil
	}

	res, err := tx.ExecContext(ctx,
		"UPDATE appointments SET employee_id = ?, starts_at = ?, ends_at = ?, reminded_at = '' WHERE id = ?",
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
		return false, errNoRows
	}

	after, err := appointmentIn(ctx, tx, id)
	if err != nil {
		return false, err
	}
	if before != after {
		if err := recordAppointmentEvent(ctx, tx, "rescheduled", &before, after); err != nil {
			return false, err
		}
	}
	return false, tx.Commit()
}

type querier interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

func hasOverlap(ctx context.Context, q querier, employeeID int, from, to time.Time, excludeID int) (bool, error) {
	var exists int

	err := q.QueryRowContext(ctx, `
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

func (r *AppointmentRepository) UpdateStatus(ctx context.Context, id int, status models.AppointmentStatus) error {
	return r.updateWithHistory(ctx, id, "status", func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, "UPDATE appointments SET status=? WHERE id=?", string(status), id)
		return err
	})
}

func (r *AppointmentRepository) DeleteAppointmentById(ctx context.Context, id int) error {
	res, err := r.db.ExecContext(ctx, "DELETE FROM appointments WHERE id = ?", id)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return errNoRows
	}

	return nil
}

func (r *AppointmentRepository) GetClientStats(ctx context.Context, clientID int, now time.Time) (models.ClientStats, error) {
	stats := models.ClientStats{ClientID: clientID}

	var (
		lastVisit sql.NullString
		nextVisit sql.NullString
		nowText   = now.UTC().Format(timeLayout)
	)

	err := r.db.QueryRowContext(ctx, `
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

type reminderScanner struct {
	row   rowScanner
	extra []any
}

func (s reminderScanner) Scan(dest ...any) error {
	return s.row.Scan(append(dest, s.extra...)...)
}

func (r *AppointmentRepository) GetDueReminders(ctx context.Context, from, to time.Time) ([]models.Reminder, error) {
	rows, err := r.db.QueryContext(ctx, `
        SELECT a.id, a.company_id, a.client_id, a.employee_id, a.service_id,
               a.starts_at, a.ends_at, a.status, a.comment, a.price, c.name, c.phone
        FROM appointments a
        JOIN clients c ON c.id = a.client_id
        WHERE a.status IN (?, ?) AND a.reminded_at = ''
          AND a.starts_at > ? AND a.starts_at <= ?
        ORDER BY a.starts_at`,
		string(models.AppointmentPending), string(models.AppointmentConfirmed),
		from.UTC().Format(timeLayout), to.UTC().Format(timeLayout),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	reminders := []models.Reminder{}
	for rows.Next() {
		var reminder models.Reminder
		scanner := reminderScanner{row: rows, extra: []any{&reminder.ClientName, &reminder.ClientPhone}}

		reminder.Appointment, err = scanAppointment(scanner)
		if err != nil {
			return nil, err
		}
		reminders = append(reminders, reminder)
	}

	return reminders, rows.Err()
}

func (r *AppointmentRepository) MarkReminded(ctx context.Context, id int, at time.Time) error {
	res, err := r.db.ExecContext(ctx,
		"UPDATE appointments SET reminded_at = ? WHERE id = ?",
		at.UTC().Format(timeLayout), id,
	)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return errNoRows
	}

	return nil
}

func (r *AppointmentRepository) MarkRemindedIfUnchanged(ctx context.Context, id int, startsAt, at time.Time) error {
	_, err := r.db.ExecContext(ctx, `UPDATE appointments SET reminded_at=? WHERE id=? AND starts_at=? AND status IN ('pending','confirmed')`, at.UTC().Format(timeLayout), id, startsAt.UTC().Format(timeLayout))
	return err
}
