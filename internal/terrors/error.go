package terrors

import (
	"fmt"
	"net/http"
)

type Error struct {
	Code int
	Msg  string
	Err  error
}

func (e *Error) Error() string {
	return fmt.Sprintf("code: %d, msg: %s, err: %v", e.Code, e.Msg, e.Err)
}

func InternalServerError(err error, msg string) *Error {
	return &Error{
		Code: http.StatusInternalServerError,
		Msg:  msg,
		Err:  err,
	}
}

func NotFound(err error, msg string) *Error {
	return &Error{
		Code: http.StatusNotFound,
		Msg:  msg,
		Err:  err,
	}
}

func Unauthorized(err error, msg string) *Error {
	return &Error{
		Code: http.StatusUnauthorized,
		Msg:  msg,
		Err:  err,
	}
}

func InvalidRequest(err error, msg string) *Error {
	return &Error{
		Code: http.StatusBadRequest,
		Msg:  msg,
		Err:  err,
	}
}

func Forbidden(err error, msg string) *Error {
	return &Error{
		Code: http.StatusForbidden,
		Msg:  msg,
		Err:  err,
	}
}
