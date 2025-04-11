package ratelimit

import (
	"fmt"
	"log/slog"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/didip/tollbooth/v8"
	"github.com/didip/tollbooth/v8/errors"
	"github.com/didip/tollbooth/v8/limiter"
	cache "github.com/go-pkgz/expirable-cache/v3"
	"golang.org/x/time/rate"
)

type Limiter struct {
	lmt *limiter.Limiter
	// Map of limiters with TTL
	tokenBuckets  cache.Cache[string, *rate.Limiter]
	ipHeaderKey   string
	maxPerTTL     float64
	expirationTTL time.Duration
}

func NewLimiter(maxPerTTL float64, ttl time.Duration, opts ...Option) *Limiter {
	l := &Limiter{
		ipHeaderKey:   "RemoteAddr",
		maxPerTTL:     maxPerTTL,
		expirationTTL: ttl,
	}

	for _, opt := range opts {
		opt(l)
	}

	lmt := tollbooth.NewLimiter(maxPerTTL, &limiter.ExpirableOptions{
		DefaultExpirationTTL: ttl,
	})

	lmt.SetIPLookup(limiter.IPLookup{
		Name:           l.ipHeaderKey,
		IndexFromRight: 0,
	})

	l.lmt = lmt

	l.tokenBuckets = cache.NewCache[string, *rate.Limiter]().WithTTL(ttl)

	return l
}

func (l *Limiter) LimitByRequest(w http.ResponseWriter, r *http.Request) *errors.HTTPError {
	shouldSkip := tollbooth.ShouldSkipLimiter(l.lmt, r)
	if shouldSkip {
		return nil
	}

	sliceKeys := tollbooth.BuildKeys(l.lmt, r)

	// Get the lowest value over all keys to return in headers.
	// Start with high arbitrary number so that any limit returned would be lower and would
	// overwrite the value we start with.
	var (
		remain = math.MaxInt32
		reset  = time.Now().Unix()
	)

	// Loop sliceKeys and check if one of them has error.
	for _, keys := range sliceKeys {
		err, keysRemain, keysReset := l.LimitByKeysAndReturn(keys)
		if remain > keysRemain {
			remain = keysRemain
		}
		if reset < keysReset {
			reset = keysReset
		}

		if err != nil {
			l.setRateLimitResponseHeaders(w, remain, reset)
			return err
		}
	}

	l.setRateLimitResponseHeaders(w, remain, reset)
	return nil
}

func (l *Limiter) LimitByKeysAndReturn(keys []string) (*errors.HTTPError, int, int64) {
	reached, remain, reset := l.LimitReached(strings.Join(keys, "|"))
	if reached {
		return &errors.HTTPError{Message: l.lmt.GetMessage(), StatusCode: l.lmt.GetStatusCode()},
			int(remain), reset
	}

	return nil, int(remain), reset
}

func (l *Limiter) LimitReached(key string) (bool, float64, int64) {
	ttl := l.lmt.GetTokenBucketExpirationTTL()

	if ttl <= 0 {
		ttl = (&limiter.ExpirableOptions{}).DefaultExpirationTTL
	}

	return l.limitReachedWithTokenBucketTTL(key, ttl)
}

func (l *Limiter) limitReachedWithTokenBucketTTL(key string, ttl time.Duration) (bool, float64, int64) {
	lmtBurst := l.lmt.GetBurst()
	l.lmt.Lock()
	defer l.lmt.Unlock()

	if _, found := l.tokenBuckets.Get(key); !found {
		l.tokenBuckets.Set(
			key,
			rate.NewLimiter(rate.Every(l.expirationTTL), lmtBurst),
			ttl,
		)
	}

	expiringMap, found := l.tokenBuckets.Get(key)
	if !found {
		return false, l.maxPerTTL, time.Now().Unix()
	}

	exp, b := l.tokenBuckets.GetExpiration(key)
	if !b {
		exp = time.Now()
	}
	slog.Info(exp.String())
	return !expiringMap.Allow(), expiringMap.TokensAt(time.Now()), exp.Unix()
}

func (l *Limiter) setRateLimitResponseHeaders(w http.ResponseWriter, remain int, reset int64) {
	w.Header().Add("X-RateLimit-Limit", fmt.Sprintf("%d", int(math.Round(l.maxPerTTL))))
	w.Header().Add("X-RateLimit-Reset", strconv.FormatInt(reset, 10))
	w.Header().Add("X-RateLimit-Remaining", fmt.Sprintf("%d", remain))
}
