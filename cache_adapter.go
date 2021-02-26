package throttler

import (
	"github.com/patrickmn/go-cache"
	"time"
)

type ICacheProvider interface {
	AddItem(k string, x int64) error
	Increment(string, int64) (int64, error)
}

type cacheAdapter struct {
	*cache.Cache
}

func (c *cacheAdapter) Increment(k string, n int64) (int64, error) {
	return c.IncrementInt64(k, n)
}

func (c *cacheAdapter) AddItem(k string, item int64) error {
	return c.Add(k, item, cache.DefaultExpiration)
}

func NewCacheAdapter(defaultExpiration, cleanupInterval time.Duration) *cacheAdapter {
	return &cacheAdapter{
		cache.New(defaultExpiration, cleanupInterval),
	}
}
