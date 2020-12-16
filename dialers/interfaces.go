package dialers

import (
	"context"
	"net"

	"github.com/valyala/fasthttp"
)

// Dialer defines an interface for instances which can dial to the
// target.
type Dialer interface {
	// Dial defines a method which main purpose is to establish a plain
	// TCP with a target netloc. It need to make all ceremonies under
	// the hood. If TLS connection is mandatory and UpgradeToTLS is not
	// optional, it also has to be done there.
	Dial(ctx context.Context, host, port string) (net.Conn, error)

	// UpgradeToTLS transforms a plain TCP connection to secured one.
	// Hostname is a hostname we connect to. Sometimes we can reuse cached
	// TLS sessions based on this parameter, for example.
	UpgradeToTLS(ctx context.Context, tcpConn net.Conn, host, port string) (net.Conn, error)

	// PatchHTTPRequest has to patch HTTP request so it can be passed
	// down to a socket. Sometimes you want to do something special with
	// this request.
	PatchHTTPRequest(*fasthttp.Request)
}
