package terrors

import (
	"fmt"
	"net/http"
)

type Error struct {
	Code    int
	Message string
	Err     error
}

func (e *Error) Error() string {
	return fmt.Sprintf("code: %d, msg: %s, err: %v", e.Code, e.Message, e.Err)
}

func InternalServerError(err error, msg string) *Error {
	return &Error{
		Code:    http.StatusInternalServerError,
		Message: msg,
		Err:     err,
	}
}

func NotFound(err error, msg string) *Error {
	return &Error{
		Code:    http.StatusNotFound,
		Message: msg,
		Err:     err,
	}
}

func Unauthorized(err error, msg string) *Error {
	return &Error{
		Code:    http.StatusUnauthorized,
		Message: msg,
		Err:     err,
	}
}

func InvalidRequest(err error, msg string) *Error {
	return &Error{
		Code:    http.StatusBadRequest,
		Message: msg,
		Err:     err,
	}
}

func Forbidden(err error, msg string) *Error {
	return &Error{
		Code:    http.StatusForbidden,
		Message: msg,
		Err:     err,
	}
}
