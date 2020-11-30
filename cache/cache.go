package cache

import (
	"context"
	"time"

	"github.com/bluele/gcache"
)

type cache struct {
	ctx   context.Context
	cache gcache.Cache
}

func (c *cache) Add(key string, value interface{}) {
	select {
	case <-c.ctx.Done():
	default:
		c.cache.Set(key, value) // nolint: errcheck
	}
}

func (c *cache) Get(key string) interface{} {
	select {
	case <-c.ctx.Done():
	default:
		if value, err := c.cache.GetIFPresent(key); err == nil && value != nil {
			return value
		}
	}

	return nil
}

func New(ctx context.Context, size int, ttl time.Duration, callback EvictCallback) Interface {
	return &cache{
		ctx: ctx,
		cache: gcache.New(size).
			LFU().
			Expiration(ttl).
			EvictedFunc(func(key, value interface{}) {
				callback(key.(string), value)
			}).Build(),
	}
}
