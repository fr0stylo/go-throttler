package throttler

import (
	"net"
	"net/http"
)



type opts struct {
	ExhaustionCount      int64
	Verbose              *bool
	LimitReachedMessage  *string
	IpResolver           *func(r *http.Request) (string, error)
	Cache                ICacheProvider
	Log                  ILogger
	IpThrottlingStrategy map[*net.IPNet]int64
}

type ILogger interface {
	Print(v ...interface{})
}

func WithOpts(exhaustCount int64) *opts {
	return &opts{ExhaustionCount: exhaustCount}
}

func (o *opts) WithVerbose(value bool) *opts {
	b := value

	o.Verbose = &b

	return o
}

func (o *opts) WithLimitReachedMessage(value string) *opts {
	b := value

	o.LimitReachedMessage = &b

	return o
}

func (o *opts) WithIpResolver(value *func(r *http.Request) (string, error)) *opts {
	b := value

	o.IpResolver = b

	return o
}

func (o *opts) WithCache(value ICacheProvider) *opts {
	b := value

	o.Cache = b

	return o
}

func (o *opts) WithLogger(value ILogger) *opts {
	b := value

	o.Log = b

	return o
}

func (o *opts)WithIpThrottlingStrategy(value map[*net.IPNet]int64) *opts {
	b := value

	o.IpThrottlingStrategy = b

	return o
}