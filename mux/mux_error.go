package mux

import (
	"fmt"
	"net/http"
)

type Error struct {
	code int
	msg  string
}

func (e *Error) Error() string {
	return fmt.Sprintf("%d: %s", e.code, e.msg)
}

func NotFoundErr(msg string) *Error {
	return &Error{http.StatusNotFound, msg}
}

func InternalServerErr(msg string) *Error {
	return &Error{http.StatusInternalServerError, msg}
}

func BadRequestErr(msg string) *Error {
	return &Error{http.StatusBadRequest, msg}
}

func UnauthorizedErr(msg string) *Error {
	return &Error{http.StatusUnauthorized, msg}
}

func ForbiddenErr(msg string) *Error {
	return &Error{http.StatusForbidden, msg}
}

func ConflictErr(msg string) *Error {
	return &Error{http.StatusConflict, msg}
}

func TooManyRequestsErr(msg string) *Error {
	return &Error{http.StatusTooManyRequests, msg}
}

func ServiceUnavailableErr(msg string) *Error {
	return &Error{http.StatusServiceUnavailable, msg}
}

func UnprocessableEntityErr(msg string) *Error {
	return &Error{http.StatusUnprocessableEntity, msg}
}

func NotImplementedErr(msg string) *Error {
	return &Error{http.StatusNotImplemented, msg}
}

func GatewayTimeoutErr(msg string) *Error {
	return &Error{http.StatusGatewayTimeout, msg}
}

func MethodNotAllowedErr(msg string) *Error {
	return &Error{http.StatusMethodNotAllowed, msg}
}

func BadGatewayErr(msg string) *Error {
	return &Error{http.StatusBadGateway, msg}
}

func PreconditionFailedErr(msg string) *Error {
	return &Error{http.StatusPreconditionFailed, msg}
}

func PayloadTooLargeErr(msg string) *Error {
	return &Error{http.StatusRequestEntityTooLarge, msg}
}
