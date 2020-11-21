package headers

import "io"

type FastHTTPHeaderWrapper interface {
	Read(io.Reader) error
	DisableNormalizing()
	Headers() []byte
}
