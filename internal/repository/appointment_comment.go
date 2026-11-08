package repository

import (
	"context"

	"github.com/artemydottech/goclients/internal/models"
)

func (r *AppointmentRepository) UpdateComment(ctx context.Context, id int, comment string) error {
	result, err := r.db.ExecContext(ctx, "UPDATE appointments SET comment=? WHERE id=?", comment, id)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return models.ErrNotFound
	}
	return nil
}
