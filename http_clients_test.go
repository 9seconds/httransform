package httransform

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/valyala/fasthttp"
)

type HTTPClientsTestSuite struct {
	suite.Suite
}

func (suite *HTTPClientsTestSuite) TestPropertiesDefaultClient() {
	suite.propertyTestClient(HTTP.(*fasthttp.Client))
}

func (suite *HTTPClientsTestSuite) TestPropertiesDefaultExecutorClient() {
	suite.propertyTestClient(executorDefaultHTTPClient.(*fasthttp.Client))
}

func (suite *HTTPClientsTestSuite) TestPropertiesDefaultMakeClient() {
	suite.propertyTestClient(makeHTTPClient())
}

func (suite *HTTPClientsTestSuite) TestPropertiesPublicMakeClient() {
	suite.propertyTestClient(MakeHTTPClient().(*fasthttp.Client))
}

func (suite *HTTPClientsTestSuite) TestPropertiesPublicSOCKS5Client() {
	u := &url.URL{
		Scheme: "socks5",
		Host:   "localhost:4040",
	}

	client, err := MakeProxySOCKS5Client(u)
	suite.Nil(err)
	suite.propertyTestClient(client.(*fasthttp.Client))
}

func (suite *HTTPClientsTestSuite) TestPropertiesHTTPSProxyClient() {
	u := &url.URL{
		Scheme: "http",
		Host:   "localhost:4040",
	}

	suite.propertyTestClient(MakeHTTPSProxyClient(u).(*fasthttp.Client))
}

func (suite *HTTPClientsTestSuite) TestPropertiesHTTPProxyClient() {
	u := &url.URL{
		Scheme: "http",
		Host:   "localhost:4040",
	}

	suite.propertyTestHostClient(MakeHTTPProxyClient(u).(*fasthttp.HostClient))
}

func (suite *HTTPClientsTestSuite) TestPropertiesHostClient() {
	suite.propertyTestHostClient(makeHTTPHostClient("hostname:80"))
}

func (suite *HTTPClientsTestSuite) propertyTestClient(client *fasthttp.Client) {
	suite.True(client.DialDualStack)
	suite.True(client.DisableHeaderNamesNormalizing)
	suite.Equal(client.MaxConnsPerHost, MaxConnsPerHost)
	suite.True(client.TLSConfig.InsecureSkipVerify)
}

func (suite *HTTPClientsTestSuite) propertyTestHostClient(client *fasthttp.HostClient) {
	suite.True(client.DialDualStack)
	suite.True(client.DisableHeaderNamesNormalizing)
	suite.Equal(client.MaxConns, MaxConnsPerHost)
	suite.True(client.TLSConfig.InsecureSkipVerify)
}

func TestHTTPClients(t *testing.T) {
	suite.Run(t, &HTTPClientsTestSuite{})
}
