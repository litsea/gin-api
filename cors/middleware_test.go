package cors

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestCORS(t *testing.T) {
	t.Parallel()

	type args struct {
		m  string
		o  string
		os []string
	}
	type want struct {
		code int
		resp string
	}

	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "default-allowed",
			args: args{o: "https://test.com", os: nil},
			want: want{code: 200, resp: "OK"},
		},
		{
			name: "all-allowed",
			args: args{o: "https://test.com", os: []string{"*"}},
			want: want{code: 200, resp: "OK"},
		},
		{
			name: "empty-origin-allowed",
			args: args{os: []string{"https://test.com"}},
			want: want{code: 200, resp: "OK"},
		},
		{
			name: "match-exactly-allowed",
			args: args{o: "https://test.com", os: []string{"https://test.com"}},
			want: want{code: 200, resp: "OK"},
		},
		{
			name: "match-exactly-allowed2",
			args: args{o: "https://foo.test.com", os: []string{"https://foo.test.com"}},
			want: want{code: 200, resp: "OK"},
		},
		{
			name: "match-exactly-not-allowed",
			args: args{o: "https://foo.com", os: []string{"https://test.com"}},
			want: want{code: 403},
		},
		{
			name: "match-exactly-not-allowed2",
			args: args{o: "https://foo.test.com", os: []string{"https://test.com"}},
			want: want{code: 403},
		},
		{
			name: "wildcard-allowed",
			args: args{o: "https://foo.test.com", os: []string{"https://*.test.com"}},
			want: want{code: 200, resp: "OK"},
		},
		{
			name: "wildcard-allowed2",
			args: args{o: "https://foo.bar.test.com", os: []string{"https://*.test.com"}},
			want: want{code: 200, resp: "OK"},
		},
		{
			name: "wildcard-not-allowed",
			args: args{o: "https://test.com", os: []string{"https://*.test.com"}},
			want: want{code: 403},
		},
		{
			name: "wildcard-not-allowed2",
			args: args{o: "https://foo.com", os: []string{"https://*.test.com"}},
			want: want{code: 403},
		},
		{
			name: "wildcard-not-allowed3",
			args: args{o: "http://foo.test.com", os: []string{"https://*.test.com"}},
			want: want{code: 403},
		},
		{
			name: "multi-origins-allowed",
			args: args{
				o:  "https://test.com",
				os: []string{"https://test.com", "https://*.test.com"},
			},
			want: want{code: 200, resp: "OK"},
		},
		{
			name: "multi-origins-allowed2",
			args: args{
				o:  "https://foo.bar.test.com",
				os: []string{"https://test.com", "https://*.test.com"},
			},
			want: want{code: 200, resp: "OK"},
		},
		{
			name: "multi-origins-allowed",
			args: args{
				o: "https://test.com",
				os: []string{
					"https://foo.com", "https://*.foo.com",
					"https://bar.com", "https://*.bar.com",
				},
			},
			want: want{code: 403},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			code, resp := makeRequest(tt.args.m, tt.args.o, tt.args.os)

			assert.Equal(t, tt.want.code, code)
			assert.Equal(t, tt.want.resp, resp)
		})
	}
}

func newServer(origins []string) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	r.Use(New(
		WithAllowOrigin(origins),
	))

	r.GET("/", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "OK")
	})

	return r
}

func makeRequest(method, origin string, origins []string) (int, string) {
	if method == "" {
		method = "GET"
	}

	req, _ := http.NewRequestWithContext(context.Background(), method, "/", nil)
	req.Header.Add("Origin", origin)

	w := httptest.NewRecorder()
	s := newServer(origins)
	s.ServeHTTP(w, req)

	return w.Code, w.Body.String()
}
