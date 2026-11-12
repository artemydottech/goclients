package repository

import (
	"context"
	"database/sql"
)

func (r *AppointmentRepository) UpdateComment(ctx context.Context, id int, comment string) error {
	return r.updateWithHistory(ctx, id, "comment", func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, "UPDATE appointments SET comment=? WHERE id=?", comment, id)
		return err
	})
}
