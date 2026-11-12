package models

import (
	"errors"
	"fmt"
)

var ErrNotFound = errors.New("not found")
var ErrConflict = errors.New("conflict")

type ValidationError struct {
	Message string
}

func (e ValidationError) Error() string {
	return e.Message
}

func Invalid(format string, args ...any) error {
	return ValidationError{Message: fmt.Sprintf(format, args...)}
}
