package httransform_test

import (
	"context"
	"net"

	"github.com/9seconds/httransform/v2"
	"github.com/9seconds/httransform/v2/dialers"
	"github.com/9seconds/httransform/v2/executor"
)

func ExampleServer_proxyExecutor() {
	dialer, _ := dialers.DialerFromURL(dialers.Opts{}, "http://user:password@myproxy.host:3128")
	opts := httransform.ServerOpts{
		Executor: executor.MakeDefaultExecutor(dialer),
	}
	proxy, _ := httransform.NewServer(context.Background(), opts)
	listener, _ := net.Listen("tcp", ":3128")

	proxy.Serve(listener)
}
