package layers

type Layer interface {
	OnRequest(*LayerContext) error
	OnResponse(*LayerContext, error)
}

type fasthttpHeader interface {
	Reset()
	DisableNormalizing()
	Add(string, string)
	ContentLength() int
	SetContentLength(int)
}
