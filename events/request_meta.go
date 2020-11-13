package events

import (
	"sync"

	"github.com/valyala/fasthttp"
)

type RequestMeta struct {
	RequestID string
	Method    string
	URI       fasthttp.URI
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
