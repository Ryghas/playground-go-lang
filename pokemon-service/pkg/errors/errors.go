package errors

import "errors"

const (
	CodeNotFound   = "NOT_FOUND"
	CodeConflict   = "CONFLICT"
	CodeBadRequest = "BAD_REQUEST"
	CodeInternal   = "INTERNAL_ERROR"
)

type AppError struct {
	Code    string
	Message string
	Err     error
}

func (e *AppError) Error() string { return e.Message }
func (e *AppError) Unwrap() error { return e.Err }

func NewNotFound(msg string) *AppError {
	return &AppError{Code: CodeNotFound, Message: msg}
}

func NewConflict(msg string) *AppError {
	return &AppError{Code: CodeConflict, Message: msg}
}

func NewBadRequest(msg string) *AppError {
	return &AppError{Code: CodeBadRequest, Message: msg}
}

func IsNotFound(err error) bool   { return hasCode(err, CodeNotFound) }
func IsConflict(err error) bool   { return hasCode(err, CodeConflict) }
func IsBadRequest(err error) bool { return hasCode(err, CodeBadRequest) }

// As mengambil *AppError dari error chain (jika ada).
func As(err error) (*AppError, bool) {
	var e *AppError
	if errors.As(err, &e) {
		return e, true
	}
	return nil, false
}

func hasCode(err error, code string) bool {
	e, ok := As(err)
	return ok && e.Code == code
}
