package layers

import "github.com/9seconds/httransform/v2/headers"

// HeadersLayer defines a general layer which adds / modifies and
// removes headers.
type HeadersLayer struct {
	// These headers are set via ctx.RequestHeaders.Set method.
	RequestSet []headers.Header

	// These headers are set via ctx.RequestHeaders.SetExact method.
	RequestSetExact []headers.Header

	// These headers are removed via ctx.RequestHeaders.Remove method.
	RequestRemove []string

	// These headers are removed via ctx.RequestHeaders.RemoveExact method.
	RequestRemoveExact []string

	// These headers are set via ctx.ResponseHeaders.Set method if we see
	// no failures.
	ResponseOkSet []headers.Header

	// These headers are set via ctx.ResponseHeaders.SetExact method if we
	// see no failures.
	ResponseOkSetExact []headers.Header

	// These headers are removed via ctx.ResponseHeaders.Remove method if we
	// see no failures.
	ResponseOkRemove []string

	// These headers are removed via ctx.ResponseHeaders.RemoveExact method
	// if we see no failures.
	ResponseOkRemoveExact []string

	// These headers are set via ctx.ResponseHeaders.Set method if we see
	// failures.
	ResponseErrSet []headers.Header

	// These headers are set via ctx.ResponseHeaders.SetExact method if we
	// see failures.
	ResponseErrSetExact []headers.Header

	// These headers are removed via ctx.ResponseHeaders.Remove method if we
	// see failures.
	ResponseErrRemove []string

	// These headers are removed via ctx.ResponseHeaders.RemoveExact method
	// if we see failures.
	ResponseErrRemoveExact []string
}

// OnRequest is to conform Layer interface.
func (h *HeadersLayer) OnRequest(ctx *Context) error {
	h.execute(&ctx.RequestHeaders,
		h.RequestSet, h.RequestSetExact,
		h.RequestRemove, h.RequestRemoveExact)

	return nil
}

// OnResponse is to conform Layer interface.
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
		hdrs.Set(set[i].Name(), set[i].Value(), true)
	}

	for i := range setExact {
		hdrs.SetExact(setExact[i].Name(), setExact[i].Value(), true)
	}

	for i := range remove {
		hdrs.Remove(remove[i])
	}

	for i := range removeExact {
		hdrs.RemoveExact(removeExact[i])
	}
}
