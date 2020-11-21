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
	for i := range h.RequestSet {
		ctx.RequestHeaders.Set(h.RequestSet[i].Name, h.RequestSet[i].Value, true)
	}

	for i := range h.RequestSetExact {
		ctx.RequestHeaders.SetExact(h.RequestSetExact[i].Name, h.RequestSetExact[i].Value, true)
	}

	for i := range h.RequestRemove {
		ctx.RequestHeaders.Remove(h.RequestRemove[i])
	}

	for i := range h.RequestRemove {
		ctx.RequestHeaders.RemoveExact(h.RequestRemoveExact[i])
	}

	return nil
}

func (h *HeadersLayer) OnResponse(ctx *Context, err error) error {
	if err != nil {
		h.onResponseErr(ctx)

		return err
	}

	h.onResponseOk(ctx)

	return nil
}

func (h *HeadersLayer) onResponseErr(ctx *Context) {
	for i := range h.ResponseErrSet {
		ctx.ResponseHeaders.Set(h.ResponseErrSet[i].Name, h.ResponseErrSet[i].Value, true)
	}

	for i := range h.ResponseErrSetExact {
		ctx.ResponseHeaders.SetExact(h.ResponseErrSetExact[i].Name, h.ResponseErrSetExact[i].Value, true)
	}

	for i := range h.ResponseErrRemove {
		ctx.ResponseHeaders.Remove(h.ResponseErrRemove[i])
	}

	for i := range h.ResponseErrRemoveExact {
		ctx.ResponseHeaders.RemoveExact(h.ResponseErrRemoveExact[i])
	}
}

func (h *HeadersLayer) onResponseOk(ctx *Context) {
	for i := range h.ResponseOkSet {
		ctx.ResponseHeaders.Set(h.ResponseOkSet[i].Name, h.ResponseOkSet[i].Value, true)
	}

	for i := range h.ResponseOkSetExact {
		ctx.ResponseHeaders.SetExact(h.ResponseOkSetExact[i].Name, h.ResponseOkSetExact[i].Value, true)
	}

	for i := range h.ResponseOkRemove {
		ctx.ResponseHeaders.Remove(h.ResponseOkRemove[i])
	}

	for i := range h.ResponseErrRemoveExact {
		ctx.ResponseHeaders.RemoveExact(h.ResponseErrRemoveExact[i])
	}
}
