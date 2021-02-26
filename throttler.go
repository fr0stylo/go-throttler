package throttler

import (
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
	"time"
)

type ICacheProvider interface {
	AddItem(k string, x int64) error
	Increment(string, int64) (int64, error)
}

type ThrottlerOpts struct {
	ExhaustionCount     int64
	Cache               ICacheProvider
	Verbose             *bool
	Log                 ILogger
	LimitReachedMessage *string
}

type middleware struct {
	cache           ICacheProvider
	log             ILogger
	verbose         bool
	exhaustionCount int64
}

type ILogger interface {
	Print(v ...interface{})
}

func newMiddleware(opts *ThrottlerOpts) *middleware {
	var cacheProvider ICacheProvider
	cacheProvider = NewCacheAdapter(time.Minute, time.Second*90)
	if opts.Cache != nil {
		cacheProvider = opts.Cache
	}

	var logger ILogger
	logger = log.Default()
	if opts.Log != nil {
		logger = opts.Log
	}

	verbose := true
	if opts.Verbose != nil {
		verbose = *opts.Verbose
	}

	return &middleware{
		cache:           cacheProvider,
		log:             logger,
		verbose:         verbose,
		exhaustionCount: opts.ExhaustionCount,
	}
}

func (m *middleware) throttle(log chan<- string, r *http.Request) error {
	ipString := strings.Split(r.RemoteAddr, ":")[0]
	ip := net.ParseIP(ipString)
	if ip != nil {
		value, err := m.cache.Increment(ip.String(), 1)

		if err != nil {
			m.cache.AddItem(ip.String(), 1)

			return nil
		}

		if value > m.exhaustionCount {
			log <- fmt.Sprintf("[Throttled] %s reached count", ip)
			return errors.New("Request limit reached, Cooldown a bit !")
		}
	}

	return nil
}

func Middleware(opts *ThrottlerOpts) func(next http.HandlerFunc) http.HandlerFunc {
	mwObject := newMiddleware(opts)
	logChan := mwObject.startLogRoutine()
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if err := mwObject.throttle(logChan, r); err != nil {
				w.WriteHeader(http.StatusTooManyRequests)
				w.Write([]byte(err.Error()))

				return
			}

			next.ServeHTTP(w, r)
		}
	}
}

func (m *middleware) startLogRoutine() chan<- string {
	logChan := make(chan string)

	go func() {
		isOpen := true
		for isOpen {
			select {
			case logToPrint, ok := <-logChan:
				isOpen = ok

				if m.verbose {
					m.log.Print(logToPrint)
				}
			}
		}
	}()

	return logChan
}
