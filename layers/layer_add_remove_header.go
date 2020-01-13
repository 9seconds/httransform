package layers

import "github.com/9seconds/httransform/v2/headers"

type AddRemoveHeaderLayer struct {
	FrozenRequestHeaders   map[string]string
	FrozenResponseHeaders  map[string]string
	DefaultRequestHeaders  map[string]string
	DefaultResponseHeaders map[string]string
	AbsentRequestHeaders   []string
	AbsentResponseHeaders  []string
}

func (a *AddRemoveHeaderLayer) OnRequest(ctx *LayerContext) error {
	a.apply(a.FrozenRequestHeaders, a.DefaultRequestHeaders, a.AbsentRequestHeaders, ctx.RequestHeaders)
	return nil
}

func (a *AddRemoveHeaderLayer) OnResponse(ctx *LayerContext, _ error) {
	a.apply(a.FrozenResponseHeaders, a.DefaultResponseHeaders, a.AbsentResponseHeaders, ctx.ResponseHeaders)
}

func (a *AddRemoveHeaderLayer) apply(frozen, defaults map[string]string, absent []string, headers *headers.Set) {
	for k, v := range defaults {
		if _, ok := headers.GetString(v); !ok {
			headers.SetString(k, v)
		}
	}

	for k, v := range frozen {
		headers.SetString(k, v)
	}

	for _, v := range absent {
		headers.DeleteString(v)
	}
}
