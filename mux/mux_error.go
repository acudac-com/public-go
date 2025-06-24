package mux

import (
	"fmt"
	"net/http"
)

type HttpErr struct {
	code int
	msg  string
}

func (e *HttpErr) Error() string {
	return fmt.Sprintf("%d: %s", e.code, e.msg)
}

func NotFoundErr(msg string) *HttpErr {
	return &HttpErr{http.StatusNotFound, msg}
}

func InternalServerErr(msg string) *HttpErr {
	return &HttpErr{http.StatusInternalServerError, msg}
}

func BadRequestErr(msg string) *HttpErr {
	return &HttpErr{http.StatusBadRequest, msg}
}

func UnauthorizedErr(msg string) *HttpErr {
	return &HttpErr{http.StatusUnauthorized, msg}
}

func ForbiddenErr(msg string) *HttpErr {
	return &HttpErr{http.StatusForbidden, msg}
}

func ConflictErr(msg string) *HttpErr {
	return &HttpErr{http.StatusConflict, msg}
}

func TooManyRequestsErr(msg string) *HttpErr {
	return &HttpErr{http.StatusTooManyRequests, msg}
}

func ServiceUnavailableErr(msg string) *HttpErr {
	return &HttpErr{http.StatusServiceUnavailable, msg}
}

func UnprocessableEntityErr(msg string) *HttpErr {
	return &HttpErr{http.StatusUnprocessableEntity, msg}
}

func NotImplementedErr(msg string) *HttpErr {
	return &HttpErr{http.StatusNotImplemented, msg}
}

func GatewayTimeoutErr(msg string) *HttpErr {
	return &HttpErr{http.StatusGatewayTimeout, msg}
}

func MethodNotAllowedErr(msg string) *HttpErr {
	return &HttpErr{http.StatusMethodNotAllowed, msg}
}

func BadGatewayErr(msg string) *HttpErr {
	return &HttpErr{http.StatusBadGateway, msg}
}

func PreconditionFailedErr(msg string) *HttpErr {
	return &HttpErr{http.StatusPreconditionFailed, msg}
}

func PayloadTooLargeErr(msg string) *HttpErr {
	return &HttpErr{http.StatusRequestEntityTooLarge, msg}
}
