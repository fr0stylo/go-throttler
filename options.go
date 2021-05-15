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

type ILogger interface {
	Print(v ...interface{})
}

func WithOpts(exhaustCount int64) *opts {
	return &opts{exhaustionCount: exhaustCount}
}

func (o *opts) WithVerbose(value bool) *opts {
	b := value

	o.verbose = &b

	return o
}

func (o *opts) WithLimitReachedMessage(value string) *opts {
	b := value

	o.limitReachedMessage = &b

	return o
}

func (o *opts) WithIpResolver(value *func(r *http.Request) (string, error)) *opts {
	o.ipResolver = value

	return o
}

func (o *opts) WithCache(value ICacheProvider) *opts {
	b := value

	o.cache = b

	return o
}

func (o *opts) WithLogger(value ILogger) *opts {
	b := value

	o.log = b

	return o
}

func (o *opts) WithIpThrottlingStrategy(value map[*net.IPNet]int64) *opts {
	o.ipThrottlingStrategy = value

	return o
}
