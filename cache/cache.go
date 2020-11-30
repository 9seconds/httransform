package cache

import (
	"time"

	"github.com/bluele/gcache"
)

type cache struct {
	cache gcache.Cache
}

func (c *cache) Add(key string, value interface{}) {
	c.cache.Set(key, value) // nolint: errcheck
}

func (c *cache) Get(key string) interface{} {
	if value, err := c.cache.GetIFPresent(key); err == nil && value != nil {
		return value
	}

	return nil
}

// New returns a new LRU/LFU cache based on given parameters.
func New(size int, ttl time.Duration, callback EvictCallback) Interface {
	return &cache{
		cache: gcache.New(size).
			LFU().
			Expiration(ttl).
			EvictedFunc(func(key, value interface{}) {
				callback(key.(string), value)
			}).Build(),
	}
}
