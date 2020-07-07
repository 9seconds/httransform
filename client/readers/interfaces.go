package readers

import "io"

type Reader interface {
	io.ReadCloser
	Release()
}
