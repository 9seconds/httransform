package cache

import (
	"time"

	"github.com/jahnestacado/tlru"
)

type cache struct {
	cache tlru.TLRU
}

func (c *cache) Add(key string, value interface{}) {
	c.cache.Set(tlru.Entry{ // nolint: errcheck
		Key:   key,
		Value: value,
	})
}

func (c *cache) Get(key string) interface{} {
	if entry := c.cache.Get(key); entry != nil {
		return entry.Value
	}

	return nil
}

// New returns a new LRU/LFU cache based on given parameters.
func New(size int, ttl time.Duration, callback EvictCallback) Interface {
	evictionChannel := make(chan tlru.EvictedEntry)

	go func() {
		for evt := range evictionChannel {
			callback(evt.Key, evt.Value)
		}
	}()

	return &cache{
		cache: tlru.New(tlru.Config{
			Size:            size,
			TTL:             ttl,
			EvictionChannel: &evictionChannel,
		}),
	}
}
