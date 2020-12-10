package dns_test

import (
	"context"
	"net"
	"testing"

	"github.com/9seconds/httransform/v2/cache"
	"github.com/9seconds/httransform/v2/dns"
	"github.com/stretchr/testify/suite"
)

type DNSTestSuite struct {
	suite.Suite

	cache dns.Interface
}

func (suite *DNSTestSuite) SetupTest() {
	suite.cache = dns.New(dns.CacheSize, dns.CacheTTL, cache.NoopEvictCallback)
}

func (suite *DNSTestSuite) TestResolveNames() {
	ctx := context.Background()
	names, err := suite.cache.Lookup(ctx, "amazon.com")

	suite.NoError(err)

	names2, err := net.DefaultResolver.LookupHost(ctx, "amazon.com")

	suite.NoError(err)
	suite.ElementsMatch(names, names2)
}

func (suite *DNSTestSuite) TestResolveIP() {
	ctx := context.Background()
	names, err := suite.cache.Lookup(ctx, "64.233.165.102")

	suite.NoError(err)
	suite.NotEmpty(names)
}

func TestDNS(t *testing.T) {
	suite.Run(t, &DNSTestSuite{})
}
