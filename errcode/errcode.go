package errcode

import (
	"net/http"
)

const (
	CodeOK = 0
)

var _ error = (*Error)(nil)

type Error struct {
	Code           int `json:"code"`
	httpCode       int
	Message        string `json:"msg"`
	errLogDisabled bool
}

func New(code int, msg string, httpCode ...int) *Error {
	hc := http.StatusInternalServerError
	if len(httpCode) != 0 {
		hc = httpCode[0]
	}

	return &Error{Code: code, httpCode: hc, Message: msg}
}

// HTTPCode default 500.
func (e *Error) HTTPCode() int {
	return e.httpCode
}

func (e *Error) Error() string {
	return e.Message
}

func (e *Error) DisableErrorLog(v bool) *Error {
	e.errLogDisabled = v
	return e
}

func (e *Error) IsErrorLogDisabled() bool {
	return e.errLogDisabled
}

var (
	// basic.

	OK                    = New(CodeOK, "OK", http.StatusOK)
	ErrBadRequest         = New(http.StatusBadRequest, "ErrBadRequest", http.StatusBadRequest)
	ErrForbidden          = New(http.StatusForbidden, "ErrForbidden", http.StatusForbidden)
	ErrNotFound           = New(http.StatusNotFound, "ErrNotFound", http.StatusNotFound)
	ErrMethodNotAllowed   = New(http.StatusMethodNotAllowed, "ErrMethodNotAllowed", http.StatusMethodNotAllowed)
	ErrTooManyRequests    = New(http.StatusTooManyRequests, "ErrTooManyRequests", http.StatusTooManyRequests)
	ErrInternalServer     = New(http.StatusInternalServerError, "ErrInternalServer")
	ErrServiceUnavailable = New(http.StatusServiceUnavailable, "ErrServiceUnavailable", http.StatusServiceUnavailable)

	// req bad request.

	ErrBadRequestFormat        = New(1001, "ErrBadRequestFormat", http.StatusBadRequest)
	ErrBadRequestFormatNumeric = New(1002, "ErrBadRequestFormatNumeric", http.StatusBadRequest)
	ErrBadRequestFormatTime    = New(1003, "ErrBadRequestFormatTime", http.StatusBadRequest)
	ErrBadRequestFormatJSON    = New(1004, "ErrBadRequestFormatJSON", http.StatusBadRequest)

	// server common error.

	// https://bugzilla.mozilla.org/show_bug.cgi?id=907800
	// Do not use http.StatusRequestTimeout, as it may cause Firefox to automatically retry the request
	ErrServiceTimeout = New(1101, "ErrServiceTimeout", http.StatusServiceUnavailable).DisableErrorLog(true)
)
