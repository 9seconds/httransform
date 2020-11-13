package events

import "sync"

type ResponseMeta struct {
	Request    *RequestMeta
	StatusCode int
}

func (r *ResponseMeta) Reset() {
	r.Request = nil
	r.StatusCode = 0
}

var poolResponseMeta = sync.Pool{
	New: func() interface{} {
		return &ResponseMeta{}
	},
}

func AcquireResponseMeta() *ResponseMeta {
	return poolResponseMeta.Get().(*ResponseMeta)
}

func ReleaseResponseMeta(resp *ResponseMeta) {
	resp.Reset()
	poolResponseMeta.Put(resp)
}
