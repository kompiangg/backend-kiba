package errors

import "errors"

// Repository layer related errors
var (
	ErrNotFound       = errors.New("not found")
	ErrDuplicateValue = errors.New("duplicate value")
)
