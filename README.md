# gin-api

## CORS

```golang
import (
	"github.com/litsea/gin-api/cors"
)

r := gin.New()

r.Use(cors.New(
	cors.WithAllowOrigin([]string{"https://foo.com", "https://*.foo.com"}),
))
```

### Config

* Default:
  * [cors.New()](cors/middleware.go)
* Custom: [cors.Option](cors/option.go)
  * `WithAllowMethods()`
  * `WithAllowHeaders()`
  * `WithAllowOrigin()`
  * `WithAllowWildcard()`
  * `WithAllowCredentials()`
  * `WithMaxAge()`

## Error Code

```golang
import (
	"github.com/litsea/gin-api/errcode"
)

var (
	ErrFooBar       = errcode.New(100001, "ErrFooBar")
	ErrWithHTTPCode = errcode.New(100002, "ErrWithHTTPCode", http.StatusForbidden)
)
```

> * Default HTTP code is `http.StatusInternalServerError`
> * The message(msgID) in `errcode.Error` will be used for translation
> * There are some built-in error codes, see: [errcode.go](errcode/errcode.go)
> * To avoid confusion, code should not be duplicated

### Error message translation

Format: `msgID: "translated value"`

Example:

```yaml
ErrFooBar: "Foo bar error"
ErrLongMsg: >-
  long long long
  long long error message
```

> * See: https://github.com/litsea/gin-i18n
> * Built-in localize folder: [localize](localize)
> * Embedded translations: [api.Localize](api.go),

## HTTP Response

```golang
import (
	"github.com/gin-gonic/gin"
	"github.com/litsea/gin-api/errcode"
	"github.com/litsea/gin-api"
)

// Success
api.Success(ctx, data)

// Error
api.Error(ctx, errcode.ErrXXX)

// Validation Error
r := gin.New()
r.GET("/validation/required", func(ctx *gin.Context) {
	req := &struct {
		Name string `binding:"required" form:"name"`
	}{}
	if err := ctx.ShouldBind(&req); err != nil {
		api.VError(ctx, err, req)

		return
	}

	api.Success(ctx, nil)
})
```

## Logger and Error Wrapping

### Logger

```golang
import (
	"log/slog"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/litsea/gin-api/log"
)

r := gin.New()

l := log.New(
	slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{})),
)
r.Use(log.Middleware(l))
```

You can use [AddAttributesFunc](log/log.go) to have other gin logging middlewares handle the error logging.

Example for [slog-gin](https://github.com/samber/slog-gin)

```golang
import (
	"log/slog"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/litsea/gin-api/log"
	ginslog "github.com/litsea/gin-slog"
	sloggin "github.com/samber/slog-gin"
)

r := gin.New()

l := log.New(
	slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{})),
	ginslog.AddCustomAttributes,
)
r.Use(log.Middleware(l))
```

See:

* https://github.com/litsea/gin-slog
* https://github.com/litsea/sentry-slog
* https://github.com/litsea/log-slog
* https://github.com/litsea/gin-example

### Error Wrapping

Errors returned to the frontend do not need to be logged repeatedly, use `fmt.Errorf()` to wrap the error and log the details. When the error type of the outer wrapper is [errcode.Error](errcode/errcode.go), the front end will only receive the translated message of this error code, and the logger will record all the error contexts.

Example:

Router `Login()`

```golang
err := srv.Login(...)
if err != nil {
	api.Error(ctx, err)

	return
}
```

Service `Login()`

```golang
err := model.LoginCheck(...)
if err != nil {
	return nil, fmt.Errorf("service.Login: %w, username=%s, %w",
		errcode.ErrLoginCheckFailed, username, err)
}
```

Model `LoginCheck()`

```golang
err := db.Find(...)
if err != nil {
	return nil, fmt.Errorf("model.LoginCheck: %w", err)
}
```

> * The frontend will only get the translated error message of `errcode.ErrLoginCheckFailed`
> * The log message can be `service.Login: ErrLoginCheckFailed, username=abc, model.LoginCheck: dial tcp 10.0.0.1:3306: connect: connection refused`
