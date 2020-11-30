package cache

// EvictCallback is going to be executed on an entry eviction.
type EvictCallback func(string, interface{})

// Interface defines a minimal set of methods and their signatures which
// is going to be used everywhere in this library.
type Interface interface {
	// Add puts a value into the cache.
	Add(key string, value interface{})

	// Get extracts value from the cache. If it is impossible to return
	// a value, then it returns a nil.
	Get(key string) interface{}
}
