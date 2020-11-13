package layers

type Layer interface {
	OnRequest(*Context) error
	OnResponse(*Context, error)
}
