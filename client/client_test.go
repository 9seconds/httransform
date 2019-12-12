package client

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/mccutchen/go-httpbin/httpbin"
	"github.com/stretchr/testify/suite"
	"github.com/valyala/fasthttp"
)

type ClientTestSuite struct {
	suite.Suite

	srv *httptest.Server
}

type BrokenHTTPConnection struct {
	resp string
}

func (b *BrokenHTTPConnection) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	hj, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "webserver doesn't support hijacking", http.StatusInternalServerError)
		return
	}

	conn, bufrw, err := hj.Hijack()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	bufrw.WriteString(b.resp) // nolint: errcheck
	bufrw.Flush()
	conn.Close()
}

func (suite *ClientTestSuite) SetupSuite() {
	handler := httpbin.New().Handler()
	suite.srv = httptest.NewServer(handler)
}

func (suite *ClientTestSuite) TearDownSuite() {
	suite.srv.Close()
}

func (suite *ClientTestSuite) TestStreamContent() {
	dialer, _ := NewSimpleDialer(FastHTTPBaseDialer, time.Second)
	client := NewClient(dialer) // nolint: gosec
	req := &fasthttp.Request{}
	resp := &fasthttp.Response{}

	req.SetRequestURI(suite.srv.URL + "/status/200")
	err := client.Do(req, resp)

	suite.Nil(err)
	suite.Equal(resp.Body(), []byte{})
}

func (suite *ClientTestSuite) TestPooledDialer() {
	dialer, _ := NewPooledDialer(FastHTTPBaseDialer, time.Second, 5)
	go dialer.Run()

	client := NewClient(dialer) // nolint: gosec
	req := &fasthttp.Request{}
	resp := &fasthttp.Response{}

	req.SetRequestURI(suite.srv.URL + "/bytes/100")
	err := client.Do(req, resp)
	suite.Nil(err)
	suite.Len(resp.Body(), 100)

	req.SetRequestURI(suite.srv.URL + "/bytes/100")
	err = client.Do(req, resp)
	suite.Nil(err)
	suite.Len(resp.Body(), 100)
}

func (suite *ClientTestSuite) TestPooledDialerSameSocket() {
	dialer, _ := NewPooledDialer(FastHTTPBaseDialer, time.Second, 5)
	go dialer.Run()

	client := NewClient(dialer) // nolint: gosec
	req := &fasthttp.Request{}
	resp := &fasthttp.Response{}

	req.SetRequestURI(suite.srv.URL + "/ip")

	err := client.Do(req, resp)
	suite.Nil(err)

	body := resp.Body()

	req.SetRequestURI(suite.srv.URL + "/ip")
	err = client.Do(req, resp)
	suite.Nil(err)
	suite.Equal(body, resp.Body())
}

func (suite *ClientTestSuite) TestBrokenResponseHeader() {
	srv := httptest.NewServer(&BrokenHTTPConnection{
		resp: "HTTP/1.1",
	})

	defer srv.Close()

	dialer, _ := NewPooledDialer(FastHTTPBaseDialer, time.Second, 5)

	go dialer.Run()

	client := NewClient(dialer) // nolint: gosec
	req := &fasthttp.Request{}
	resp := &fasthttp.Response{}

	req.SetRequestURI(srv.URL + "/ip")
	err := client.Do(req, resp)
	suite.NotNil(err)
}

func TestClient(t *testing.T) {
	suite.Run(t, &ClientTestSuite{})
}
