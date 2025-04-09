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
	"github.com/litsea/gin-api/log"
)

const (
	maxValidationErrorValueLength = 50
)

var errInvokeErrorFuncWithoutError = errors.New("invoke error function without error")

type Response struct {
	Code     int           `json:"code"`
	Message  string        `json:"msg"`
	Data     any           `json:"data,omitempty"`
	Errors   []DetailError `json:"errors,omitempty"`
	httpCode int
}

type PageResponse struct {
	Total int64 `json:"total"`           // Total number of records
	Size  int   `json:"size"`            // Page size
	Page  int   `json:"page"`            // Current page number
	Items any   `json:"items,omitempty"` // Data list
}

type CursorPageResp struct {
	Total int64 `json:"total"`           // Total number of records
	Size  int   `json:"size"`            // Page size
	Start any   `json:"start"`           // Current page cursor start
	Next  any   `json:"next"`            // Next page cursor start
	Items any   `json:"items,omitempty"` // Data list
}

type DetailError struct {
	Code    int    `json:"code,omitempty"`
	Field   string `json:"field,omitempty"`
	Message string `json:"msg"`
	Value   any    `json:"value,omitempty"`
}

func NewSuccessResponse(data any) *Response {
	return &Response{
		Code:    0,
		Message: "Success",
		Data:    data,
	}
}

func NewPageResponse(total int64, size, page int, items any) *Response {
	return NewSuccessResponse(&PageResponse{
		Total: total,
		Size:  size,
		Page:  page,
		Items: items,
	})
}

func NewCursorPageResponse(total int64, size int, start, next, items any) *Response {
	return NewSuccessResponse(&CursorPageResp{
		Total: total,
		Size:  size,
		Start: start,
		Next:  next,
		Items: items,
	})
}

func Success(ctx *gin.Context, data any) {
	ctx.JSON(http.StatusOK, NewSuccessResponse(data))
}

func PageSuccess(ctx *gin.Context, total int64, size, page int, items any) {
	ctx.JSON(http.StatusOK, NewPageResponse(total, size, page, items))
}

func CursorPageSuccess(ctx *gin.Context, total int64, size int, start, next, items any) {
	ctx.JSON(http.StatusOK, NewCursorPageResponse(total, size, start, next, items))
}

func Error(ctx *gin.Context, err error) {
	l := log.GetLoggerFromContext(ctx)

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
			l.Debug("HTTP bad request", "err", err)
		case http.StatusUnauthorized, http.StatusForbidden, http.StatusNotFound,
			http.StatusMethodNotAllowed, http.StatusTooManyRequests:
			// ignore log
		default:
			msgErr := fmt.Sprintf("API error: code=%d %s", code, ee.Error())
			l.ErrorRequest(ctx, msgErr, map[string]any{
				"status": httpCode,
				"err":    err,
			})
		}
	case errors.As(err, &ve):
		validateError(ctx, ve)

		return
	default:
		if ctx.Writer.Status() != http.StatusOK {
			httpCode = ctx.Writer.Status()
			code = httpCode
		}

		// Do not send unknown error messages to the frontend
		message = i18n.E(ctx, errcode.ErrInternalServer.Error())

		var (
			msgErr string
			rErr   error
		)
		if err != nil {
			msgErr = fmt.Sprintf("HTTP error: code=%d %s", code, http.StatusText(code))
			rErr = err
		} else {
			msgErr = "Incorrect error function invoke"
			rErr = errInvokeErrorFuncWithoutError
		}

		l.ErrorRequest(ctx, msgErr, map[string]any{
			"status": httpCode,
			"err":    rErr,
		})
	}

	ctx.JSON(httpCode, Response{
		Code:     code,
		Message:  message,
		httpCode: httpCode,
	})
}

// VError response validate errors.
func VError(ctx *gin.Context, err error, req any) {
	l := log.GetLoggerFromContext(ctx)

	params := map[string]string{}
	for _, p := range ctx.Params {
		params[p.Key] = p.Value
	}

	l.Debug("Failed to validate request data", "err", err, "params", params)

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
