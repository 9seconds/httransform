package headers

import "io"

// FastHTTPHeaderWrapper defines a wrapper for fasthttp Headers which
// has to be put into Headers on Init.
//
// Unfortunately, fasthttp Headers are instances, not interfaces and
// some operations vary so a unified interface is required.
type FastHTTPHeaderWrapper interface {
	// Read resets an underlying structure and populates it with raw
	// data which is coming from the given io.Reader instance.
	Read(io.Reader) error

	// DisableNormalizing explicitly prevents from normalizing of header
	// names and order.
	DisableNormalizing()

	// ResetConnectionClose drops a status of Connection header to
	// default state: absence.
	ResetConnectionClose()

	// Headers return a bytes containing raw headers (with a first line!).
	Headers() []byte
}
