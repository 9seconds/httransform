package dns

import (
	"time"

	"github.com/9seconds/httransform/v2/cache"
)

const (
	// CacheSize is a size of DNS cache.
	CacheSize = 512

	// CacheTTL is a TTL of a DNS record in a cache.
	CacheTTL = 5 * time.Minute
)

// Default is a default DNS resolver you usually want to use everywhere.
var Default Interface = New(CacheSize, CacheTTL, cache.NoopEvictCallback)
