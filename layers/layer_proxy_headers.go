package layers

// ProxyHeadersLayer is simplified version of HeadersLayer with 0
// configuration but created only to wipe out some proxy headers from
// requests and responses.
type ProxyHeadersLayer struct{}

// OnRequest is to conform Layer interface.
func (p ProxyHeadersLayer) OnRequest(ctx *Context) error {
	ctx.RequestHeaders.Remove("Proxy-Authenticate")
	ctx.RequestHeaders.Remove("Proxy-Authorization")

	return nil
}

// OnResponse is to conform Layer interface.
func (p ProxyHeadersLayer) OnResponse(ctx *Context, err error) error {
	ctx.RequestHeaders.Remove("Proxy-Connection")

	return err
}
