package cache

import (
	"time"

	"github.com/dgraph-io/ristretto"
)

// please remove it when ristretto >0.0.3 will be released.
type cacheItem struct {
	key   string
	value interface{}
}

type cache struct {
	cache *ristretto.Cache
	ttl   time.Duration
}

func (c *cache) Add(key string, value interface{}) {
	item := &cacheItem{
		key:   key,
		value: value,
	}

	c.cache.SetWithTTL(key, item, 1, c.ttl)
}

func (c *cache) Get(key string) interface{} {
	if entry, ok := c.cache.Get(key); ok && entry != nil {
		return entry.(*cacheItem).value
	}

	return nil
}

// New returns a new LRU/LFU cache based on given parameters.
func New(size int, ttl time.Duration, callback EvictCallback) Interface {
	config := &ristretto.Config{
		// 10x is recommended in official documentation
		NumCounters: int64(10 * size), // nolint: gomnd
		MaxCost:     int64(size),
		// 64 is also taken from official documentation.
		BufferItems: 64, // nolint: gomnd
		Metrics:     false,
		OnEvict: func(_, _ uint64, value interface{}, _ int64) {
			vv, _ := value.(*cacheItem)
			callback(vv.key, vv.value)
		},
	}

	obj, err := ristretto.NewCache(config)
	if err != nil {
		panic(err)
	}

	return &cache{
		cache: obj,
		ttl:   ttl,
	}
}
