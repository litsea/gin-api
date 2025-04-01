package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"net/http"
	"strconv"
	"time"

	"github.com/cdfmlr/ellipsis"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/litsea/gin-api/errcode"
	"github.com/litsea/gin-api/i18n"
	"github.com/litsea/gin-api/logger"
)

const (
	maxValidationErrorValueLength = 50
)

type Response struct {
	Code     int           `form:"code"             json:"code"`
	Message  string        `json:"msg,omitempty"`
	Data     any           `json:"data,omitempty"`
	Errors   []DetailError `json:"errors,omitempty"`
	httpCode int
}

type DetailError struct {
	Code    int    `json:"code,omitempty"`
	Field   string `json:"field,omitempty"`
	Message string `json:"msg"`
	Value   any    `json:"value,omitempty"`
}

func Success(ctx *gin.Context, data any) {
	r := Response{
		Code:    0,
		Message: "Success",
		Data:    data,
	}

	ctx.JSON(http.StatusOK, r)
}

func Error(ctx *gin.Context, err error) {
	httpCode := http.StatusInternalServerError
	code := httpCode

	var (
		message string
		ee      *errcode.Error
		ve      validator.ValidationErrors
	)

	switch {
	case errors.As(err, &ee):
		if ee.HTTPCode() > 0 {
			httpCode = ee.HTTPCode()
		}

		code = ee.Code
		message = i18n.E(ctx, ee.Error())

		switch ee.HTTPCode() {
		case http.StatusBadRequest:
			logger.Debug("HTTP bad request", "err", err)
		case http.StatusUnauthorized, http.StatusForbidden, http.StatusNotFound,
			http.StatusMethodNotAllowed, http.StatusTooManyRequests:
			// ignore log
		default:
			logger.Error("HTTP common error", "err", err)
		}
	case errors.As(err, &ve):
		validateError(ctx, ve)

		return
	default:
		if ctx.Writer.Status() != http.StatusOK {
			httpCode = ctx.Writer.Status()
			code = ctx.Writer.Status()
		}

		message = err.Error()
		logger.Error("HTTP error", "err", err)
	}

	ctx.JSON(httpCode, Response{
		Code:     code,
		Message:  message,
		httpCode: httpCode,
	})
}

// VError response validate errors.
func VError(ctx *gin.Context, err error, req any) {
	logger.Debug("Failed to validate request data", "err", err, "req", req)

	var ve validator.ValidationErrors

	switch {
	case errors.As(err, &ve):
		validateError(ctx, ve)
	default:
		validateParseError(ctx, err)
	}
}

// validateError response validate errors.
func validateError(ctx *gin.Context, ve validator.ValidationErrors) {
	ec := errcode.ErrBadRequest
	errs := make([]DetailError, len(ve))

	for i, fe := range ve {
		msg := i18n.V(ctx, fe)

		errs[i] = DetailError{
			0, fe.Field(), msg,
			html.EscapeString(ellipsis.Centering(
				fmt.Sprintf("%v", fe.Value()), maxValidationErrorValueLength),
			),
		}
	}

	ctx.JSON(ec.HTTPCode(), Response{
		Code:     ec.HTTPCode(),
		Message:  i18n.E(ctx, ec.Error()),
		Errors:   errs,
		httpCode: ec.HTTPCode(),
	})
}

// validateParseError response validate parse errors.
func validateParseError(ctx *gin.Context, err error) {
	var (
		ec *errcode.Error
		ne *strconv.NumError
		te *time.ParseError
		je *json.UnmarshalTypeError
	)

	var (
		field string
		value string
	)

	switch {
	case errors.As(err, &ne):
		ec = errcode.ErrBadRequestFormatNumeric
		value = ne.Num
	case errors.As(err, &te):
		ec = errcode.ErrBadRequestFormatTime
		value = te.Value
	case errors.As(err, &je):
		ec = errcode.ErrBadRequestFormatJSON
		field = je.Field
		value = fmt.Sprintf("expect: %s, actual: %s",
			je.Type.String(), je.Value)
	default:
		ec = errcode.ErrBadRequestFormat
		value = err.Error()
	}

	ctx.JSON(ec.HTTPCode(), Response{
		Code:    ec.HTTPCode(),
		Message: errcode.ErrBadRequest.Error(),
		Errors: []DetailError{
			{ec.Code, field, i18n.E(ctx, ec.Error()), value},
		},
		httpCode: ec.HTTPCode(),
	})
}
