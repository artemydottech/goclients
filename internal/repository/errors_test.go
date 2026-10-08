package repository

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/artemydottech/goclients/internal/models"
)

func TestMapNoRows(t *testing.T) {
	err := mapNoRows(sql.ErrNoRows)

	if !errors.Is(err, models.ErrNotFound) {
		t.Errorf("mapped error must match models.ErrNotFound, got %v", err)
	}
}

func TestMapNoRowsKeepsOtherErrors(t *testing.T) {
	boom := errors.New("boom")

	if got := mapNoRows(boom); got != boom {
		t.Errorf("got %v, want the original error", got)
	}
	if got := mapNoRows(nil); got != nil {
		t.Errorf("got %v, want nil", got)
	}
}
