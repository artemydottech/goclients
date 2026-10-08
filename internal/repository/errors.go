package repository

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/artemydottech/goclients/internal/models"
)

var errNoRows = fmt.Errorf("%w: %w", models.ErrNotFound, sql.ErrNoRows)

func mapNoRows(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return errNoRows
	}

	return err
}
