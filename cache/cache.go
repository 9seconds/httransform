package cache

import (
	"context"
	"time"

	"github.com/jahnestacado/tlru"
)

type tlruCache struct {
	ctx   context.Context
	cache tlru.TLRU
}

func (t *tlruCache) Add(key string, value interface{}) {
	select {
	case <-t.ctx.Done():
	default:
		t.cache.Set(tlru.Entry{Key: key, Value: value}) // nolint: errcheck
	}
}

func (t *tlruCache) Get(key string) interface{} {
	select {
	case <-t.ctx.Done():
	default:
		if entry := t.cache.Get(key); entry != nil {
			return entry.Value
		}
	}

	return nil
}

func New(ctx context.Context, size int, ttl time.Duration, callback EvictCallback) Interface {
	channelEvictions := make(chan tlru.EvictedEntry, 1)
	config := tlru.Config{
		Size:            size,
		TTL:             ttl,
		EvictionPolicy:  tlru.LRA,
		EvictionChannel: &channelEvictions,
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case evt := <-channelEvictions:
				callback(evt.Key, evt.Value)
			}
		}
	}()

	return &tlruCache{
		cache: tlru.New(config),
	}
}
