package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/litsea/gin-api/errcode"
	"github.com/litsea/gin-api/logger"
	"github.com/litsea/gin-api/testdata"
	g18n "github.com/litsea/gin-i18n"
	"github.com/litsea/i18n"
	"github.com/stretchr/testify/assert"
	"golang.org/x/text/language"
)

func TestResponse(t *testing.T) {
	t.Parallel()

	type args struct {
		lng language.Tag
		uri string
	}
	type want struct {
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
				code: 403,
				msg:  "Forbidden",
			},
		},
		{
			name: "en-validation-required",
			args: args{
				uri: "/validation/required",
				lng: language.English,
			},
			want: want{
				code:      400,
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
				code:      400,
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
				code: 0,
				msg:  "Success",
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
				code: 403,
				msg:  "Verbotene",
			},
		},
		{
			name: "de-validation-required",
			args: args{
				uri: "/validation/required",
				lng: language.German,
			},
			want: want{
				code:      400,
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
				code:      400,
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
				code: 403,
				msg:  "Forbidden",
			},
		},
		{
			name: "fr-validation-required",
			args: args{
				uri: "/validation/required",
				lng: language.French,
			},
			want: want{
				code:      400,
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
				code:      400,
				msg:       "Bad Request",
				detailMsg: "Status must be one of the values [1 2]",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			data := makeRequest(tt.args.lng, tt.args.uri)

			var r Response
			err := json.Unmarshal([]byte(data), &r)

			assert.NoError(t, err)
			assert.Equal(t, tt.want.code, r.Code)
			assert.Equal(t, tt.want.msg, r.Message)
			if tt.want.detailMsg != "" {
				assert.Equal(t, tt.want.detailMsg, r.Errors[0].Message)
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

func newServer() *gin.Engine {
	logger.SetDefaultLogger()
	bindValidator()

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	gi := g18n.New(
		g18n.WithOptions(
			i18n.WithLanguages(language.English, language.German),
			i18n.WithLoaders(
				i18n.EmbedLoader(Localize, "./localize/"),
				i18n.EmbedLoader(testdata.Localize, "./localize/"),
			),
		),
	)

	r.Use(gi.Localize())

	r.GET("/403", func(ctx *gin.Context) {
		Error(ctx, errcode.ErrForbidden)
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

func makeRequest(lng language.Tag, path string) string {
	req, _ := http.NewRequestWithContext(context.Background(), "GET", path, nil)
	req.Header.Add("Accept-Language", lng.String())

	w := httptest.NewRecorder()
	s := newServer()
	s.ServeHTTP(w, req)

	return w.Body.String()
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
