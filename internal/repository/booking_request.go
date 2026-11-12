package repository

import (
	"context"

	"github.com/artemydottech/goclients/internal/models"
)

func readBookingRequest(ctx context.Context, q querier, clientID int, key string) (models.BookingRequest, error) {
	var result models.BookingRequest
	err := q.QueryRowContext(ctx, "SELECT fingerprint,appointment_id FROM booking_requests WHERE client_id=? AND request_key=?", clientID, key).Scan(&result.Fingerprint, &result.AppointmentID)
	return result, mapNoRows(err)
}

func (r *AppointmentRepository) FindBookingRequest(ctx context.Context, clientID int, key string) (models.BookingRequest, error) {
	return readBookingRequest(ctx, r.db, clientID, key)
}

func (r *AppointmentRepository) ReserveBooking(ctx context.Context, key, hash string, a models.Appointment) (models.BookingResult, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return models.BookingResult{}, err
	}
	defer tx.Rollback()
	reservation, err := tx.ExecContext(ctx, "INSERT OR IGNORE INTO booking_requests(client_id,request_key,fingerprint) VALUES(?,?,?)", a.ClientID, key, hash)
	if err != nil {
		return models.BookingResult{}, err
	}
	inserted, err := reservation.RowsAffected()
	if err != nil {
		return models.BookingResult{}, err
	}
	if inserted == 0 {
		previous, err := readBookingRequest(ctx, tx, a.ClientID, key)
		if err != nil {
			return models.BookingResult{}, err
		}
		if previous.Fingerprint != hash {
			return models.BookingResult{}, models.ErrConflict
		}
		return models.BookingResult{ID: previous.AppointmentID, Replayed: true}, nil
	}
	busy, err := hasOverlap(ctx, tx, a.EmployeeID, a.StartsAt, a.EndsAt, 0)
	if err != nil {
		return models.BookingResult{}, err
	}
	if busy {
		return models.BookingResult{}, models.Invalid("this time is already taken for the employee")
	}
	id, err := insertAppointment(ctx, tx, a)
	if err != nil {
		return models.BookingResult{}, err
	}
	if _, err = tx.ExecContext(ctx, "UPDATE booking_requests SET appointment_id=? WHERE client_id=? AND request_key=?", id, a.ClientID, key); err != nil {
		return models.BookingResult{}, err
	}
	return models.BookingResult{ID: id}, tx.Commit()
}
