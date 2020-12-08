package layers

type ProxyHeadersLayer struct{}

func (p ProxyHeadersLayer) OnRequest(ctx *Context) error {
	ctx.RequestHeaders.Remove("Proxy-Authenticate")
	ctx.RequestHeaders.Remove("Proxy-Authorization")

	return nil
}

func (p ProxyHeadersLayer) OnResponse(_ *Context, err error) error {
	return err
}
