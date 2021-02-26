package throttler

import (
	"testing"
	"time"
)

func TestWithOpts(t *testing.T) {
	opt := WithOpts(UNLIMITED)

	if opt.ExhaustionCount != UNLIMITED {
		t.Errorf("Exhaustion count should be equal %d was %d", UNLIMITED, opt.ExhaustionCount)
	}

	log := NewSpyLogger()
	opt.WithLogger(log)

	if opt.Log != log {
		t.Error("Logger references not match")
	}

	opt.WithVerbose(true)
	if *opt.Verbose != true {
		t.Error("Verbose not set")
	}

	adapter := NewCacheAdapter(time.Second, time.Second)
	opt.WithCache(adapter)
	if opt.Cache != adapter {
		t.Error("Cache references not match")
	}
}
