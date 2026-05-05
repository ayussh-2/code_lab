package utils

import "fmt"

type AppError struct {
	Status  int
	Message string
	Err     error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func NewAppError(status int, message string, err error) *AppError {
	return &AppError{
		Status:  status,
		Message: message,
		Err:     err,
	}
}
