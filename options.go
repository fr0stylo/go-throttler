package throttler

import (
	"net"
	"net/http"
)

type opts struct {
	exhaustionCount      int64
	verbose              *bool
	limitReachedMessage  *string
	ipResolver           *func(r *http.Request) (string, error)
	cache                ICacheProvider
	log                  ILogger
	ipThrottlingStrategy map[*net.IPNet]int64
}

// Logger interface used all over the project
type ILogger interface {
	Print(v ...interface{})
}

// NewOpts creates new options object builder.
func NewOpts(exhaustCount int64) *opts {
	return &opts{exhaustionCount: exhaustCount}
}

// WithVerbose function that sets verbose flag
func (o *opts) WithVerbose(value bool) *opts {
	b := value

	o.verbose = &b

	return o
}

// WithLimitReachedMessage function that sets message that are displayed when threshold is reached
func (o *opts) WithLimitReachedMessage(value string) *opts {
	b := value

	o.limitReachedMessage = &b

	return o
}

// WithIpResolver function used to resolve ip address from request
// Use case: when requests original ip request are added to other header, ex: X-Forwarded-for
func (o *opts) WithIpResolver(value *func(r *http.Request) (string, error)) *opts {
	o.ipResolver = value

	return o
}

// WithCache function used to override cache provider. Any provider can be used, it must implement ICacheProvider interface
func (o *opts) WithCache(value ICacheProvider) *opts {
	b := value

	o.cache = b

	return o
}

// WithLogger function used to override logger to project used logger. Defaults to log package
func (o *opts) WithLogger(value ILogger) *opts {
	b := value

	o.log = b

	return o
}

// WithIpThrottlingStrategy used to create separate throttling resources over ips, or cdirs.
func (o *opts) WithIpThrottlingStrategy(value map[*net.IPNet]int64) *opts {
	o.ipThrottlingStrategy = value

	return o
}
