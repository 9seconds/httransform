package cache_test

import (
	"testing"
	"time"

	"github.com/9seconds/httransform/v2/cache"
	"github.com/stretchr/testify/suite"
)

type CacheTestSuite struct {
	suite.Suite

	cache cache.Interface
}

func (suite *CacheTestSuite) SetupTest() {
	suite.cache = cache.New(10, time.Second, cache.NoopEvictCallback)
}

func (suite *CacheTestSuite) TestGetNothing() {
	suite.Nil(suite.cache.Get("nothing"))
}

func (suite *CacheTestSuite) TestAddGet() {
	suite.cache.Add("key", 1)
	suite.EqualValues(1, suite.cache.Get("key"))
}

func (suite *CacheTestSuite) TestEvict() {
	var foundKey string
	var foundValue interface{}

	cc := cache.New(10, 200*time.Millisecond, func(key string, value interface{}) {
		foundKey = key
		foundValue = value
	})

	cc.Add("key", 1)
	suite.Eventually(func() bool {
		return cc.Get("key") == nil
	}, time.Second, 50*time.Millisecond)

	suite.Equal("key", foundKey)
	suite.EqualValues(1, foundValue)
}

func TestCache(t *testing.T) {
	suite.Run(t, &CacheTestSuite{})
}
