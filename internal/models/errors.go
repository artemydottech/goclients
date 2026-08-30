package models

import "fmt"

// ValidationError — некорректный ввод, а не сбой хранилища: обработчик
// отвечает на неё 400, а не 500.
type ValidationError struct {
	Message string
}

func (e ValidationError) Error() string {
	return e.Message
}

func Invalid(format string, args ...any) error {
	return ValidationError{Message: fmt.Sprintf(format, args...)}
}
