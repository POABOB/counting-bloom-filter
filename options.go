package counting_bloom_filter

import "time"

const (
	NO_EXPIRATION      = 0
	LAZY_EXPIRATION    = 1
	RESET_EVERY_PERIOD = 2
	EXPIRY_DURATION    = 3
)

// Option represents the optional function.
type Option func(opts *Options)

type ExpiryStrategy int

type Options struct {
	ExpiryStrategy ExpiryStrategy
	Duration       time.Duration
}

func loadOptions(options ...Option) *Options {
	opts := new(Options)
	for _, option := range options {
		option(opts)
	}
	return opts
}

func WithOptions(options Options) Option {
	return func(opts *Options) {
		*opts = options
	}
}

// WithExpiryDuration sets up the interval time of cleaning up the bloom filter
func WithExpiryDuration(expiryStrategy ExpiryStrategy, expiryDuration time.Duration) Option {
	return func(opts *Options) {
		opts.ExpiryStrategy = expiryStrategy
		opts.Duration = expiryDuration
	}
}
