package dialers_test

import (
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

	httpHttpbin   *httptest.Server
	tlsHttpbin    *httptest.Server
	proxyListener net.Listener
	dialer        dialers.Dialer
}

func (suite *HttpProxyTestSuite) SetupSuite() {
	httpbinApp := httpbin.NewHTTPBin()
	suite.httpHttpbin = httptest.NewServer(httpbinApp.Handler())
	suite.tlsHttpbin = httptest.NewTLSServer(httpbinApp.Handler())

	suite.proxyListener, _ = net.Listen("tcp", "127.0.0.1:0")
	proxy, _ := httpproxy.NewProxy()
	srv := &http.Server{Handler: proxy}

	go srv.Serve(suite.proxyListener)

	auth, _ := dialers.NewProxyAuth(suite.proxyListener.Addr().String(), "", "")
	opts := dialers.Opts{TLSSkipVerify: true, Timeout: 5 * time.Second}

	suite.dialer = dialers.NewHTTPProxy(opts, auth)
}

func (suite *HttpProxyTestSuite) TearDownSuite() {
	suite.proxyListener.Close()
	suite.httpHttpbin.Close()
	suite.tlsHttpbin.Close()
}

func TestHTTPProxy(t *testing.T) {
	suite.Run(t, &HttpProxyTestSuite{})
}
