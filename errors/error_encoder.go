package errors

import (
	"encoding/json"
	"strings"
	"sync"
)

type errorJSON struct {
	Error struct {
		Code    string                `json:"code"`
		Message string                `json:"message"`
		Stack   []errorJSONStackEntry `json:"stack"`
	} `json:"error"`
}

type errorJSONStackEntry struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	StatusCode int    `json:"status_code"`
}

type errorEncoder struct {
	obj     errorJSON
	buf     strings.Builder
	encoder *json.Encoder
}

func (e *errorEncoder) Encode() string {
	if err := e.encoder.Encode(e.obj); err != nil {
		panic(err)
	}

	encoded := e.buf.String()

	e.buf.Reset()

	return encoded[:len(encoded)-1]
}

func (e *errorEncoder) Reset() {
	e.buf.Reset()
	e.obj.Error.Stack = e.obj.Error.Stack[:0]
}

var poolErrorEncoder = sync.Pool{
	New: func() interface{} {
		enc := &errorEncoder{}
		enc.encoder = json.NewEncoder(&enc.buf)

		enc.encoder.SetEscapeHTML(false)

		return enc
	},
}

func acquireErrorEncoder() *errorEncoder {
	return poolErrorEncoder.Get().(*errorEncoder)
}

func releaseErrorEncoder(enc *errorEncoder) {
	enc.Reset()
	poolErrorEncoder.Put(enc)
}
