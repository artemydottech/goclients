package repository

import (
	"context"
	"encoding/json"
	"github.com/artemydottech/goclients/internal/models"
	"time"
)

func appointmentIn(ctx context.Context, q querier, id int) (models.Appointment, error) {
	return scanAppointment(q.QueryRowContext(ctx, `SELECT id,company_id,client_id,employee_id,service_id,starts_at,ends_at,status,comment,price FROM appointments WHERE id=?`, id))
}
func recordAppointmentEvent(ctx context.Context, e execer, kind string, before *models.Appointment, after models.Appointment) error {
	previous, err := json.Marshal(before)
	if err != nil {
		return err
	}
	next, err := json.Marshal(after)
	if err != nil {
		return err
	}
	_, err = e.ExecContext(ctx, "INSERT INTO appointment_events(appointment_id,kind,occurred_at,before_json,after_json) VALUES(?,?,?,?,?)", after.ID, kind, time.Now().UTC().Format(time.RFC3339Nano), string(previous), string(next))
	return err
}
func (r *AppointmentRepository) GetAppointmentHistory(ctx context.Context, id int, page models.Page) ([]models.AppointmentEvent, error) {
	query, args := paginate("SELECT id,appointment_id,kind,occurred_at,before_json,after_json FROM appointment_events WHERE appointment_id=? ORDER BY id DESC", []any{id}, page)
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []models.AppointmentEvent{}
	for rows.Next() {
		var event models.AppointmentEvent
		var at, before, after string
		if err := rows.Scan(&event.ID, &event.AppointmentID, &event.Kind, &at, &before, &after); err != nil {
			return nil, err
		}
		event.OccurredAt, err = time.Parse(time.RFC3339Nano, at)
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal([]byte(before), &event.Before); err != nil {
			return nil, err
		}
		if err := json.Unmarshal([]byte(after), &event.After); err != nil {
			return nil, err
		}
		result = append(result, event)
	}
	return result, rows.Err()
}
