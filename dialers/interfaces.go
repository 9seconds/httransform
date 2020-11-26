package dialers

import (
	"context"
	"net"

	"github.com/valyala/fasthttp"
)

type Dialer interface {
	Dial(context.Context, string, string) (net.Conn, error)
	UpgradeToTLS(context.Context, net.Conn, string) (net.Conn, error)
	PatchHTTPRequest(*fasthttp.Request)
}
