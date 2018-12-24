package client

import (
	"crypto/tls"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/valyala/fasthttp"
)

type ClientTestSuite struct {
	suite.Suite
}

func (suite *ClientTestSuite) TestOK() {
	dialer, _ := NewSimpleDialer(FastHTTPBaseDialer, time.Second)
	client := NewClient(dialer, &tls.Config{InsecureSkipVerify: true})
	req := &fasthttp.Request{}
	resp := &fasthttp.Response{}

	req.SetRequestURI("https://httpbin.scrapinghub.com/stream/10")

	err := client.Do(req, resp)
	fmt.Println("ERROR", err)
	fmt.Println("HEADER", string(resp.Header.Header()))
	fmt.Println("BODY", string(resp.Body()))
}

func TestClient(t *testing.T) {
	suite.Run(t, &ClientTestSuite{})
}
