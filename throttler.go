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

	ipResolveFn := ipResolver
	if opts.IpResolver != nil {
		ipResolveFn = *opts.IpResolver
	}

	return &middleware{
		cache:                cacheProvider,
		log:                  logger,
		verbose:              verbose,
		exhaustionCount:      opts.ExhaustionCount,
		ipResolver:           ipResolveFn,
		ipThrottlingStrategy: opts.IpThrottlingStrategy,
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

	if value > m.getExhaustValue(ip) && value != UNLIMITED {
		log <- fmt.Sprintf("[Throttled] %s reached count", ip)
		return errors.New("Request limit reached, Cooldown a bit !")
	}

	return nil
}

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
