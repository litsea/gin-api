package ratelimit

type Option func(*Limiter)

func WithIPHeaderKey(k string) Option {
	return func(l *Limiter) {
		l.ipHeaderKey = k
	}
}
