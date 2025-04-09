package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	g18n "github.com/litsea/gin-i18n"
	"github.com/litsea/i18n"
	"github.com/stretchr/testify/assert"
	"golang.org/x/text/language"

	"github.com/litsea/gin-api/errcode"
	"github.com/litsea/gin-api/log"
	"github.com/litsea/gin-api/testdata"
)

var (
	errCustom503   = errcode.New(999, "errCustom503", http.StatusServiceUnavailable)
	errUnknownCode = errors.New("unknown error code")
)

func TestResponse(t *testing.T) {
	t.Parallel()

	type args struct {
		lng language.Tag
		uri string
	}
	type want struct {
		httpCode  int
		code      int
		msg       string
		detailMsg string
	}

	tests := []struct {
		name string
		args args
		want want
	}{
		// English
		{
			name: "en-403",
			args: args{
				uri: "/403",
				lng: language.English,
			},
			want: want{
				httpCode: errcode.ErrForbidden.HTTPCode(),
				code:     errcode.ErrForbidden.Code,
				msg:      "Forbidden",
			},
		},
		{
			name: "en-custom-503",
			args: args{
				uri: "/custom-503",
				lng: language.English,
			},
			want: want{
				httpCode: errCustom503.HTTPCode(),
				code:     errCustom503.Code,
				msg:      "Customized service not available",
			},
		},
		{
			name: "en-unknown-500",
			args: args{
				uri: "/unknown-500",
				lng: language.English,
			},
			want: want{
				httpCode: errcode.ErrInternalServer.HTTPCode(),
				code:     errcode.ErrInternalServer.Code,
				msg:      "Internal Server Error",
			},
		},
		{
			name: "en-validation-required",
			args: args{
				uri: "/validation/required",
				lng: language.English,
			},
			want: want{
				httpCode:  errcode.ErrBadRequest.HTTPCode(),
				code:      errcode.ErrBadRequest.Code,
				msg:       "Bad Request",
				detailMsg: "Name must not be empty",
			},
		},
		{
			name: "en-validation-enum",
			args: args{
				uri: "/validation/enum?status=3",
				lng: language.English,
			},
			want: want{
				httpCode:  errcode.ErrBadRequest.HTTPCode(),
				code:      errcode.ErrBadRequest.Code,
				msg:       "Bad Request",
				detailMsg: "Status must be one of the values [1 2]",
			},
		},
		{
			name: "en-validation-enum-valid",
			args: args{
				uri: "/validation/enum?status=1",
				lng: language.English,
			},
			want: want{
				httpCode: errcode.OK.HTTPCode(),
				code:     errcode.OK.Code,
				msg:      "Success",
			},
		},
		// German
		{
			name: "de-403",
			args: args{
				uri: "/403",
				lng: language.German,
			},
			want: want{
				httpCode: errcode.ErrForbidden.HTTPCode(),
				code:     errcode.ErrForbidden.Code,
				msg:      "Verbotene",
			},
		},
		{
			name: "de-custom-503",
			args: args{
				uri: "/custom-503",
				lng: language.German,
			},
			want: want{
				httpCode: errCustom503.HTTPCode(),
				code:     errCustom503.Code,
				msg:      "Benutzerdefinierter Service nicht verfügbar",
			},
		},
		{
			name: "de-unknown-500",
			args: args{
				uri: "/unknown-500",
				lng: language.German,
			},
			want: want{
				httpCode: errcode.ErrInternalServer.HTTPCode(),
				code:     errcode.ErrInternalServer.Code,
				msg:      "Interner Server Fehler",
			},
		},
		{
			name: "de-validation-required",
			args: args{
				uri: "/validation/required",
				lng: language.German,
			},
			want: want{
				httpCode:  errcode.ErrBadRequest.HTTPCode(),
				code:      errcode.ErrBadRequest.Code,
				msg:       "Schlechte Anfrage",
				detailMsg: "Name darf nicht leer sein",
			},
		},
		{
			name: "de-validation-enum",
			args: args{
				uri: "/validation/enum?status=3",
				lng: language.German,
			},
			want: want{
				httpCode:  errcode.ErrBadRequest.HTTPCode(),
				code:      errcode.ErrBadRequest.Code,
				msg:       "Schlechte Anfrage",
				detailMsg: "Status muss einer der Werte [1 2] sein",
			},
		},
		// French (fallback)
		{
			name: "fr-403",
			args: args{
				uri: "/403",
				lng: language.French,
			},
			want: want{
				httpCode: errcode.ErrForbidden.HTTPCode(),
				code:     errcode.ErrForbidden.Code,
				msg:      "Forbidden",
			},
		},
		{
			name: "fr-custom-503",
			args: args{
				uri: "/custom-503",
				lng: language.French,
			},
			want: want{
				httpCode: errCustom503.HTTPCode(),
				code:     errCustom503.Code,
				msg:      "Customized service not available",
			},
		},
		{
			name: "fr-unknown-500",
			args: args{
				uri: "/unknown-500",
				lng: language.French,
			},
			want: want{
				httpCode: errcode.ErrInternalServer.HTTPCode(),
				code:     errcode.ErrInternalServer.Code,
				msg:      "Internal Server Error",
			},
		},
		{
			name: "fr-validation-required",
			args: args{
				uri: "/validation/required",
				lng: language.French,
			},
			want: want{
				httpCode:  errcode.ErrBadRequest.HTTPCode(),
				code:      errcode.ErrBadRequest.Code,
				msg:       "Bad Request",
				detailMsg: "Name must not be empty",
			},
		},
		{
			name: "fr-validation-enum",
			args: args{
				uri: "/validation/enum?status=3",
				lng: language.French,
			},
			want: want{
				httpCode:  errcode.ErrBadRequest.HTTPCode(),
				code:      errcode.ErrBadRequest.Code,
				msg:       "Bad Request",
				detailMsg: "Status must be one of the values [1 2]",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			code, data := makeRequest(tt.args.lng, tt.args.uri)

			var r Response
			err := json.Unmarshal([]byte(data), &r)

			assert.NoError(t, err)
			assert.Equal(t, tt.want.httpCode, code)
			assert.Equal(t, tt.want.code, r.Code)
			assert.Equal(t, tt.want.msg, r.Message)
			if tt.want.detailMsg != "" {
				var msg string
				if len(r.Errors) > 0 {
					msg = r.Errors[0].Message
				}
				assert.Equal(t, tt.want.detailMsg, msg)
			}
		})
	}
}

type enum interface {
	IsValidEnum() bool
}

type status int

func (s status) IsValidEnum() bool {
	return s == 1 || s == 2
}

func init() {
	// Avoid race
	bindValidator()
}

func newServer(mw ...gin.HandlerFunc) *gin.Engine {
	// TODO: data race warning for gin mode
	// https://github.com/gin-gonic/gin/pull/1580 (not yet released)
	// gin.SetMode(gin.ReleaseMode)

	r := gin.New()
	r.Use(mw...)

	r.GET("/403", func(ctx *gin.Context) {
		Error(ctx, errcode.ErrForbidden)
	})

	r.GET("/custom-503", func(ctx *gin.Context) {
		Error(ctx, fmt.Errorf("test: %w", errCustom503))
	})

	r.GET("/unknown-500", func(ctx *gin.Context) {
		Error(ctx, fmt.Errorf("test: %w", errUnknownCode))
	})

	r.GET("/invalid-err-func-invoke", func(ctx *gin.Context) {
		Error(ctx, nil)
	})

	r.GET("/validation/required", func(ctx *gin.Context) {
		req := &struct {
			Name string `binding:"required" form:"name"`
		}{}
		if err := ctx.ShouldBind(&req); err != nil {
			VError(ctx, err, req)

			return
		}

		Success(ctx, nil)
	})

	r.GET("/validation/enum", func(ctx *gin.Context) {
		req := &struct {
			Status status `binding:"enum=[1 2]" form:"status"`
		}{}
		if err := ctx.ShouldBind(&req); err != nil {
			VError(ctx, err, req)

			return
		}

		Success(ctx, nil)
	})

	return r
}

func makeRequest(lng language.Tag, uri string) (int, string) {
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, uri, http.NoBody)
	req.Header.Add("Accept-Language", lng.String())

	gi := g18n.New(
		g18n.WithOptions(
			i18n.WithLanguages(language.English, language.German),
			i18n.WithLoaders(
				i18n.EmbedLoader(Localize, "./localize/"),
				i18n.EmbedLoader(testdata.Localize, "./localize/"),
			),
		),
	)

	w := httptest.NewRecorder()
	s := newServer(gi.Localize())
	s.ServeHTTP(w, req)

	return w.Code, w.Body.String()
}

func bindValidator() {
	v, ok := binding.Validator.Engine().(*validator.Validate)
	if !ok {
		return
	}

	_ = v.RegisterValidation("enum", validateEnum)
}

func validateEnum(fl validator.FieldLevel) bool {
	if v, ok := fl.Field().Interface().(enum); ok {
		return v.IsValidEnum()
	}

	return false
}

func TestLoggerInResponse(t *testing.T) {
	t.Parallel()

	type args struct {
		uri string
		lv  slog.Level
	}

	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "en-403",
			args: args{
				uri: "/403",
				lv:  slog.LevelInfo,
			},
			want: nil,
		},
		{
			name: "custom-503",
			args: args{
				uri: "/custom-503",
				lv:  slog.LevelInfo,
			},
			want: []string{
				"level=ERROR",
				"API error",
				fmt.Sprintf("code=%d", errCustom503.Code),
				errCustom503.Error(),
			},
		},
		{
			name: "unknown-500",
			args: args{
				uri: "/unknown-500",
				lv:  slog.LevelInfo,
			},
			want: []string{
				"level=ERROR",
				"HTTP error",
				fmt.Sprintf("code=%d", http.StatusInternalServerError),
				http.StatusText(http.StatusInternalServerError),
				errUnknownCode.Error(),
			},
		},
		{
			name: "invalid-err-func-invoke",
			args: args{
				uri: "/invalid-err-func-invoke",
				lv:  slog.LevelInfo,
			},
			want: []string{
				"level=ERROR",
				errInvokeErrorFuncWithoutError.Error(),
			},
		},
		{
			name: "validation-required-no-debug",
			args: args{
				uri: "/validation/required",
				lv:  slog.LevelInfo,
			},
			want: nil,
		},
		{
			name: "validation-required-debug",
			args: args{
				uri: "/validation/required",
				lv:  slog.LevelDebug,
			},
			want: []string{
				"level=DEBUG",
				"Failed to validate request data",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := makeRequestForLogger(tt.args.uri, tt.args.lv)
			if len(tt.want) == 0 {
				assert.Equal(t, "", got)
			} else {
				for _, w := range tt.want {
					assert.Contains(t, got, w)
				}
			}
		})
	}
}

func makeRequestForLogger(uri string, lv slog.Level) string {
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, uri, http.NoBody)

	buffer := new(strings.Builder)
	defer buffer.Reset()

	l := log.New(
		slog.New(slog.NewTextHandler(buffer, &slog.HandlerOptions{Level: lv})),
	)

	w := httptest.NewRecorder()
	s := newServer(log.Middleware(l))
	s.ServeHTTP(w, req)

	return buffer.String()
}
