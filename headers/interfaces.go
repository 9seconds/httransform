package headers

import "io"

type FastHTTPHeaderWrapper interface {
	Read(io.Reader) error
	DisableNormalizing()
	ResetConnectionClose()
	Headers() []byte
}
