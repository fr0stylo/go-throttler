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

const UNLIMITED = -1

type middleware struct {
	cache                ICacheProvider
	log                  ILogger
	verbose              bool
	exhaustionCount      int64
	ipResolver           func(r *http.Request) (string, error)
	ipThrottlingStrategy map[*net.IPNet]int64
}

func ipResolver(r *http.Request) (string, error) {
	ipString := strings.Split(r.RemoteAddr, ":")[0]
	ip := net.ParseIP(ipString)

	if ip == nil {
		return "", errors.New("malformed ip")
	}

	return ip.String(), nil
}

func newMiddleware(opts *opts) *middleware {
	var cacheProvider ICacheProvider
	cacheProvider = NewCacheAdapter(time.Minute, time.Second*90)
	if opts.cache != nil {
		cacheProvider = opts.cache
	}

	var logger ILogger
	logger = log.Default()
	if opts.log != nil {
		logger = opts.log
	}

	verbose := true
	if opts.verbose != nil {
		verbose = *opts.verbose
	}

	ipResolveFn := ipResolver
	if opts.ipResolver != nil {
		ipResolveFn = *opts.ipResolver
	}

	return &middleware{
		cache:                cacheProvider,
		log:                  logger,
		verbose:              verbose,
		exhaustionCount:      opts.exhaustionCount,
		ipResolver:           ipResolveFn,
		ipThrottlingStrategy: opts.ipThrottlingStrategy,
	}
}

func (m *middleware) getExhaustValue(ip string) int64 {
	requestExhaustValue := m.exhaustionCount
	if m.ipThrottlingStrategy != nil {
		for k, v := range m.ipThrottlingStrategy {
			ipNet := *k
			if ipNet.Contains(net.ParseIP(ip)) {
				requestExhaustValue = v
			}
		}
	}

	return requestExhaustValue
}

func (m *middleware) throttle(log chan<- string, r *http.Request) error {
	ip, err := m.ipResolver(r)
	if err != nil {
		return err
	}

	value, err := m.cache.Increment(ip, 1)

	if err != nil {
		m.cache.AddItem(ip, 1)

		return nil
	}
	exhaustCount := m.getExhaustValue(ip)
	if exhaustCount != UNLIMITED && value > exhaustCount {
		log <- fmt.Sprintf("[Throttled] %s reached count", ip)
		return errors.New("Request limit reached, Cooldown a bit !")
	}

	return nil
}

// Middleware function that is supposed to throttle requests over period of time
// It return http.HandlerFunc that can be attached to any server providers.
func Middleware(opts *opts) func(next http.HandlerFunc) http.HandlerFunc {
	mwObject := newMiddleware(opts)
	logChan := mwObject.startLogRoutine()
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if mwObject.exhaustionCount == UNLIMITED && mwObject.ipThrottlingStrategy == nil {
				next.ServeHTTP(w, r)
				return
			}

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
