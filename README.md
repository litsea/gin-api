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
	log.WithRequestIDHeaderKey("X-Request-ID"),
	log.WithRequestHeader(true),
	log.WithRequestBody(true),
	log.WithUserAgent(true),
	log.WithStackTrace(true),
	log.WithExtraAttrs(map[string]any{ ... }),
)
r.Use(log.Middleware(l))
```

Default config: [log.New()](log/log.go)

Logging with gin request context:

```golang
l.ErrorRequest(ctx, msgErr, map[string]any{
	"status": httpCode,
	"err":    err,
})
```

See also:

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

### Panic Recovery

```golang
import (
	"github.com/gin-gonic/gin"
	"github.com/litsea/gin-api"
)

r := gin.New()

r.Use(
	api.Recovery(api.HandleRecovery()),
)
```

## Graceful Shutdown

```golang
import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/litsea/gin-api/graceful"
	apilog "github.com/litsea/gin-api/log"
	log "github.com/litsea/log-slog"
)

l := apilog.New( ... )
r := gin.New()

g := graceful.New(
	r,
	graceful.WithAddr(
		fmt.Sprintf("%s:%d", "0.0.0.0", 8080),
	),
	graceful.WithReadTimeout(15*time.Second),
	graceful.WithWriteTimeout(15*time.Second),
	graceful.WithLogger(l),
	graceful.WithCleanup(func() {
		log.Info("gracefulRunServer: test cleanup...")
		time.Sleep(5 * time.Second)
	}),
)

g.Run()
// Wait for send event to Sentry when server start failed
time.Sleep(3 * time.Second)
```

## Rate Limit

```golang
import (
	"github.com/gin-gonic/gin"
	"github.com/litsea/gin-api/ratelimit"
)

// Max 10 requests in one minute
var ipLimiter = ratelimit.NewLimiter(10, time.Minute)

r := gin.New()

r.GET("/rate-limit", ipLimiter.Middleware(), func(ctx *gin.Context) {
	// ...
})
```

### Rate Limit Response Header

* `X-RateLimit-Limit`: Limit requests
* `X-RateLimit-Remaining`: Remaining requests
* `X-RateLimit-Reset`: Limit reset seconds
