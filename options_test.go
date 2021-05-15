package throttler

import (
	"net"
	"net/http"
	"reflect"
	"testing"
	"time"
)

func TestWithOpts(t *testing.T) {
	opt := WithOpts(UNLIMITED)

	if opt.exhaustionCount != UNLIMITED {
		t.Errorf("Exhaustion count should be equal %d was %d", UNLIMITED, opt.exhaustionCount)
	}

	log := NewSpyLogger()
	opt.WithLogger(log)

	if opt.log != log {
		t.Error("Logger references not match")
	}

	opt.WithVerbose(true)
	if *opt.verbose != true {
		t.Error("verbose not set")
	}

	adapter := NewCacheAdapter(time.Second, time.Second)
	opt.WithCache(adapter)
	if opt.cache != adapter {
		t.Error("cache references not match")
	}

	message := "Message"
	opt.WithLimitReachedMessage(message)
	if *opt.limitReachedMessage != message {
		t.Error("Message do not match")
	}

	resolver := func(r *http.Request) (string, error) {
		return "", nil
	}

	opt.WithIpResolver(&resolver)
	if opt.ipResolver != &resolver {
		t.Error("IP Resolver no match")
	}

	ipfilter := map[*net.IPNet]int64{}
	opt.WithIpThrottlingStrategy(ipfilter)
	if !reflect.DeepEqual(opt.ipThrottlingStrategy, ipfilter) {
		t.Error("IP throttling strategy no match")
	}
}
