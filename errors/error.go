package errors

import "fmt"

type MissingArgError struct {
	error
}

func NewRequireError(field, message string) *MissingArgError {
	return &MissingArgError{error: fmt.Errorf("arg '%v' is required: %v", field, message)}
}

func RequireField(field string) *MissingArgError {
	return &MissingArgError{error: fmt.Errorf("arg '%v' is required", field)}
}
