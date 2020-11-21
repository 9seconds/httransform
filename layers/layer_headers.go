package layers

import "github.com/9seconds/httransform/v2/headers"

type HeadersLayer struct {
	RequestSet         []headers.Header
	RequestSetExact    []headers.Header
	RequestRemove      []string
	RequestRemoveExact []string

	ResponseOkSet         []headers.Header
	ResponseOkSetExact    []headers.Header
	ResponseOkRemove      []string
	ResponseOkRemoveExact []string

	ResponseErrSet         []headers.Header
	ResponseErrSetExact    []headers.Header
	ResponseErrRemove      []string
	ResponseErrRemoveExact []string
}

func (h *HeadersLayer) OnRequest(ctx *Context) error {
	h.execute(&ctx.RequestHeaders,
		h.RequestSet, h.RequestSetExact,
		h.RequestRemove, h.RequestRemoveExact)

	return nil
}

func (h *HeadersLayer) OnResponse(ctx *Context, err error) error {
	if err != nil {
		h.execute(&ctx.ResponseHeaders,
			h.ResponseErrSet, h.ResponseErrSetExact,
			h.ResponseErrRemove, h.ResponseErrRemoveExact)

		return err
	}

	h.execute(&ctx.ResponseHeaders,
		h.ResponseOkSet, h.ResponseOkSetExact,
		h.ResponseOkRemove, h.ResponseOkRemoveExact)

	return nil
}

func (h *HeadersLayer) execute(hdrs *headers.Headers,
	set, setExact []headers.Header,
	remove, removeExact []string) {
	for i := range set {
		hdrs.Set(set[i].Name, set[i].Value, true)
	}

	for i := range setExact {
		hdrs.Set(setExact[i].Name, setExact[i].Value, true)
	}

	for i := range remove {
		hdrs.Remove(remove[i])
	}

	for i := range removeExact {
		hdrs.RemoveExact(removeExact[i])
	}
}
