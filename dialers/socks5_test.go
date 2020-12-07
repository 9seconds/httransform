package dialers_test

import (
	"bufio"
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/9seconds/httransform/v2/dialers"
	"github.com/armon/go-socks5"
	"github.com/mccutchen/go-httpbin/httpbin"
	"github.com/stretchr/testify/suite"
	"github.com/valyala/fasthttp"
)

type Socks5TestSuite struct {
	suite.Suite

	httpHttpbin   *httptest.Server
	tlsHttpbin    *httptest.Server
	socksListener net.Listener
	dialer        dialers.Dialer
}

func (suite *Socks5TestSuite) SetupSuite() {
	httpbinApp := httpbin.NewHTTPBin()
	suite.httpHttpbin = httptest.NewServer(httpbinApp.Handler())
	suite.tlsHttpbin = httptest.NewTLSServer(httpbinApp.Handler())

	suite.socksListener, _ = net.Listen("tcp", "127.0.0.1:0")
	socksServer, _ := socks5.New(&socks5.Config{})

	auth, _ := dialers.NewProxyAuth(suite.socksListener.Addr().String(), "", "")
	opts := dialers.Opts{TLSSkipVerify: true}

	dialer, _ := dialers.NewSocks5(opts, auth)

	suite.dialer = dialer

	go socksServer.Serve(suite.socksListener)
}

func (suite *Socks5TestSuite) TestNeedToUpgradeToTLS() {
	parsedURL, _ := url.Parse(suite.tlsHttpbin.URL)
	host, port := parsedURL.Hostname(), parsedURL.Port()

	conn, err := suite.dialer.Dial(context.Background(), host, port)
	defer conn.Close()

	suite.NoError(err)

	req := fasthttp.Request{}
	resp := fasthttp.Response{}

	req.SetRequestURI(suite.tlsHttpbin.URL + "/get")
	suite.dialer.PatchHTTPRequest(&req)

	_, err = req.WriteTo(conn)

	suite.NoError(resp.Read(bufio.NewReader(conn)))
	suite.Equal(http.StatusBadRequest, resp.StatusCode())
}

func (suite *Socks5TestSuite) TestUpgradeToTLS() {
	parsedURL, _ := url.Parse(suite.tlsHttpbin.URL)
	host, port := parsedURL.Hostname(), parsedURL.Port()
	conn, err := suite.dialer.Dial(context.Background(), host, port)

	defer conn.Close()

	conn, err = suite.dialer.UpgradeToTLS(context.Background(), conn, host, port)

	suite.NoError(err)

	req := fasthttp.Request{}
	resp := fasthttp.Response{}

	req.SetRequestURI(suite.tlsHttpbin.URL + "/get")
	suite.dialer.PatchHTTPRequest(&req)

	_, err = req.WriteTo(conn)

	suite.NoError(resp.Read(bufio.NewReader(conn)))
	suite.Equal(http.StatusOK, resp.StatusCode())
}

func (suite *Socks5TestSuite) TestCannotUpgradeToTLS() {
	parsedURL, _ := url.Parse(suite.httpHttpbin.URL)
	host, port := parsedURL.Hostname(), parsedURL.Port()
	conn, err := suite.dialer.Dial(context.Background(), host, port)

	defer conn.Close()

	_, err = suite.dialer.UpgradeToTLS(context.Background(), conn, host, port)

	suite.Error(err)
}

func (suite *Socks5TestSuite) TearDownSuite() {
	suite.socksListener.Close()
	suite.httpHttpbin.Close()
	suite.tlsHttpbin.Close()
}

func TestSocks5(t *testing.T) {
	suite.Run(t, &Socks5TestSuite{})
}
