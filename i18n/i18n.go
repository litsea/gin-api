package i18n

import (
	"reflect"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/litsea/gin-i18n"
)

// E translate error message.
func E(ctx *gin.Context, msgID string) string {
	return i18n.T(ctx, msgID, nil)
}

// V translate validation error.
func V(ctx *gin.Context, fe validator.FieldError) string {
	msgID := genValidateMsgID(fe)

	msg := i18n.T(ctx, genValidateMsgID(fe), map[any]any{
		"field": fe.Field(),
		"value": fe.Param(),
	})
	if msgID == msg {
		msg = i18n.T(ctx, "validation-failed", map[any]any{
			"field": fe.Field(),
			"value": fe.Tag(),
		})
	}

	return msg
}

var suffixTypes = map[string]struct{}{
	"len": {}, "min": {}, "max": {},
	"lt": {}, "lte": {}, "gt": {}, "gte": {},
}

func genValidateMsgID(fe validator.FieldError) string {
	msgID := "validation-" + fe.Tag()

	if _, ok := suffixTypes[fe.Tag()]; ok {
		var typ string

		kind := fe.Kind()
		if kind == reflect.Ptr {
			kind = fe.Type().Elem().Kind()
		}

		switch kind { //nolint:exhaustive
		case reflect.String:
			typ = "string"
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
			reflect.Float32, reflect.Float64:
			typ = "number"
		case reflect.Slice, reflect.Map, reflect.Array:
			typ = "items"
		case reflect.Struct:
			if fe.Type() == reflect.TypeOf(time.Time{}) {
				typ = "datetime"
			}
		default:
			typ = ""
		}

		if typ != "" {
			msgID += "-" + typ
		}
	}

	return msgID
}
