package dialers_test

import (
	"bufio"
	"context"
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/9seconds/httransform/v2/dialers"
	"github.com/mccutchen/go-httpbin/httpbin"
	"github.com/stretchr/testify/suite"
	"github.com/valyala/fasthttp"
)

type BaseTestSuite struct {
	suite.Suite

	httpHttpbin *httptest.Server
	httpClient  *http.Client
	tlsHttpbin  *httptest.Server
	dialer      dialers.Dialer
}

func (suite *BaseTestSuite) SetupSuite() {
	httpbinApp := httpbin.NewHTTPBin()

	suite.httpHttpbin = httptest.NewServer(httpbinApp.Handler())
	suite.tlsHttpbin = httptest.NewTLSServer(httpbinApp.Handler())

	opt := dialers.Opts{
		TLSSkipVerify: true,
	}
	suite.dialer = dialers.NewBase(opt)

	suite.httpClient = &http.Client{
		Transport: &http.Transport{
			DialContext: dialers.StdDialerWrapper{Dialer: suite.dialer}.DialContext,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
		Timeout: 5 * time.Second,
	}
}

func (suite *BaseTestSuite) TearDownSuite() {
	suite.httpHttpbin.Close()
	suite.tlsHttpbin.Close()
}

func (suite *BaseTestSuite) TestPlainAccess() {
	resp, err := suite.httpClient.Get(suite.httpHttpbin.URL + "/get")
	defer resp.Body.Close()

	suite.NoError(err)
	suite.Equal(http.StatusOK, resp.StatusCode)
}

func (suite *BaseTestSuite) TestTLSAccess() {
	resp, err := suite.httpClient.Get(suite.tlsHttpbin.URL + "/get")
	defer resp.Body.Close()

	suite.NoError(err)
	suite.Equal(http.StatusOK, resp.StatusCode)
}

func (suite *BaseTestSuite) TestNeedToUpgradeToTLS() {
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

func (suite *BaseTestSuite) TestUpgradeToTLS() {
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

func (suite *BaseTestSuite) TestCannotUpgradeToTLS() {
	parsedURL, _ := url.Parse(suite.httpHttpbin.URL)
	host, port := parsedURL.Hostname(), parsedURL.Port()
	conn, err := suite.dialer.Dial(context.Background(), host, port)

	defer conn.Close()

	_, err = suite.dialer.UpgradeToTLS(context.Background(), conn, host, port)

	suite.Error(err)
}

func TestBase(t *testing.T) {
	suite.Run(t, &BaseTestSuite{})
}
