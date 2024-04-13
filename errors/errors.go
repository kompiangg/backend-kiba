package errors

import (
	"errors"
	"fmt"
	"runtime"

	"arduino-serial/errors/metadata"

	errorsx "github.com/go-errors/errors"
	"github.com/rs/zerolog/log"
)

type Error struct {
	*metadata.Metadata
	Err     error
	Message string

	stackMessage string
	stack        []uintptr
	frames       []StackFrame
}

// The maximum number of stackframes on any error.
var MaxStackDepth = 50

func (e Error) Error() string {
	return e.Message
}

func Wrap(cause interface{}, msg string) *Error {
	if cause == nil {
		return nil
	}

	var err error
	switch e := cause.(type) {
	case *Error:
		e.stackMessage = fmt.Sprintf("%s -- %s", e.stackMessage, msg)
		e.Message = msg
		return e
	case error:
		err = e
	default:
		err = fmt.Errorf("%v", e)
	}

	stack := make([]uintptr, MaxStackDepth)
	length := runtime.Callers(2, stack[:])
	return &Error{
		Err:          err,
		stackMessage: fmt.Sprintf("%s -- %s", err.Error(), msg),
		Message:      msg,
		stack:        stack[:length],
		Metadata:     nil,
	}
}

func (err *Error) Unwrap() error {
	return err.Err
}

func ErrorStack(paramErr error) {
	var newErr *Error

	if !As(paramErr, &newErr) {
		newErr = New(paramErr)
	}

	log.Error().Stack().Err(newErr).Msg(newErr.stackMessage)
}

func New(e interface{}) *Error {
	var err error

	switch e := e.(type) {
	case error:
		err = e
	default:
		err = fmt.Errorf("%v", e)
	}

	stack := make([]uintptr, MaxStackDepth)
	length := runtime.Callers(2, stack[:])
	return &Error{
		Err:          err,
		stackMessage: err.Error(),
		Message:      err.Error(),
		Metadata:     nil,
		stack:        stack[:length],
	}
}

func As(err error, target any) bool {
	return errorsx.As(err, target)
}

func Is(err error, target error) bool {
	if errors.Is(err, target) {
		return true
	}

	if e, ok := err.(Error); ok {
		return Is(e.Err, target)
	}

	if target, ok := target.(Error); ok {
		return Is(err, target.Err)
	}

	return false
}

func (e *Error) SetHTTPMetadata(code int) *Error {
	if e.Metadata == nil {
		e.Metadata = &metadata.Metadata{}
	}

	e.Metadata.SetHTTPMetadata(code)
	return e
}
