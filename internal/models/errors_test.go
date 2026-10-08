package models

import (
	"errors"
	"fmt"
	"testing"
)

func TestInvalidFormatsMessage(t *testing.T) {
	err := Invalid("field %s is required", "name")

	var validationErr ValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("expected ValidationError, got %T", err)
	}
	if validationErr.Message != "field name is required" {
		t.Errorf("unexpected message %q", validationErr.Message)
	}
}

func TestErrNotFoundSurvivesWrapping(t *testing.T) {
	err := fmt.Errorf("load user: %w", ErrNotFound)

	if !errors.Is(err, ErrNotFound) {
		t.Errorf("wrapped error must match ErrNotFound")
	}
}
