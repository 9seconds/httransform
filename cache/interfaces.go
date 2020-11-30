package cache

type EvictCallback func(string, interface{})

type Interface interface {
	Add(key string, value interface{})
	Get(key string) interface{}
}
