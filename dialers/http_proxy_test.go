package dialers_test

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/9seconds/httransform/v2/dialers"
	"github.com/go-httpproxy/httpproxy"
	"github.com/mccutchen/go-httpbin/httpbin"
	"github.com/stretchr/testify/suite"
)

type HttpProxyTestSuite struct {
	suite.Suite

	endpoint      *httptest.Server
	proxyListener net.Listener
	dialer        dialers.Dialer
}

func (suite *HttpProxyTestSuite) SetupSuite() {
	httpbinApp := httpbin.NewHTTPBin()
	suite.endpoint = httptest.NewTLSServer(httpbinApp.Handler())

	suite.proxyListener, _ = net.Listen("tcp", "127.0.0.1:0")
	proxy, _ := httpproxy.NewProxy()
	srv := &http.Server{Handler: proxy}

	go srv.Serve(suite.proxyListener)

	time.Sleep(50 * time.Millisecond)

	auth, _ := dialers.NewProxyAuth(suite.proxyListener.Addr().String(), "", "")
	opts := dialers.Opts{TLSSkipVerify: true, Timeout: 5 * time.Second}

	suite.dialer = dialers.NewHTTPProxy(opts, auth)
}

func (suite *HttpProxyTestSuite) TestTLS() {
	httpClient := &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
			DialTLSContext: func(ctx context.Context, network, address string) (net.Conn, error) {
				host, port, _ := net.SplitHostPort(address)

				conn, err := suite.dialer.Dial(ctx, host, port)
				if err != nil {
					return nil, err
				}

				return suite.dialer.UpgradeToTLS(ctx, conn, host, port)
			},
		},
	}

	resp, err := httpClient.Get(suite.endpoint.URL + "/get")

	suite.NoError(err)
	suite.Equal(http.StatusOK, resp.StatusCode)
}

func (suite *HttpProxyTestSuite) TearDownSuite() {
	suite.proxyListener.Close()
	suite.endpoint.Close()
}

func TestHTTPProxy(t *testing.T) {
	suite.Run(t, &HttpProxyTestSuite{})
}
