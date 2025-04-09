package log

type Option func(cfg *Config)

func WithRequestIDHeaderKey(v string) Option {
	return func(cfg *Config) {
		cfg.requestIDHeaderKey = v
	}
}

func WithUserAgent(v bool) Option {
	return func(cfg *Config) {
		cfg.withUserAgent = v
	}
}

func WithRequestHeader(v bool) Option {
	return func(cfg *Config) {
		cfg.withRequestHeader = v
	}
}

func WithRequestBody(v bool) Option {
	return func(cfg *Config) {
		cfg.withRequestBody = v
	}
}

func WithStackTrace(v bool) Option {
	return func(cfg *Config) {
		cfg.withStackTrace = v
	}
}

func WithExtraAttrs(attrs map[string]any) Option {
	return func(cfg *Config) {
		if len(attrs) > 0 {
			cfg.extraAttrs = attrs
		}
	}
}
