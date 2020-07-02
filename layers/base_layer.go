package layers

type BaseLayer struct{}

func (b *BaseLayer) OnRequest(ctx *LayerContext) error {
	return nil
}

func (b *BaseLayer) OnResponse(ctx *LayerContext, err error) {}
