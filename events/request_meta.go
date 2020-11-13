package events

import (
	"bytes"
	"sync"

	"github.com/9seconds/httransform/v2/layers"
	"github.com/valyala/fasthttp"
)

type RequestMeta struct {
	RequestID string
	Method    string
	URI       fasthttp.URI
}

func (r *RequestMeta) Init(ctx *layers.Context) {
	r.RequestID = ctx.RequestID()
	request := ctx.Request()
	r.Method = string(bytes.ToUpper(request.Header.Method()))
	request.URI().CopyTo(&r.URI)
}

func (r *RequestMeta) Reset() {
	r.RequestID = ""
	r.Method = ""
	r.URI.Reset()
}

var poolRequestMeta = sync.Pool{
	New: func() interface{} {
		return &RequestMeta{
			URI: fasthttp.URI{
				DisablePathNormalizing: true,
			},
		}
	},
}

func AcquireRequestMeta() *RequestMeta {
	return poolRequestMeta.Get().(*RequestMeta)
}

func ReleaseRequestMeta(req *RequestMeta) {
	req.Reset()
	poolRequestMeta.Put(req)
}
