package layers

type Layer interface {
	OnRequest(*LayerContext) error
	OnResponse(*LayerContext, error)
}

type fasthttpHeader interface {
	Reset()
	DisableNormalizing()
	SetBytesKV([]byte, []byte)
	ContentLength() int
	SetContentLength(int)
}
