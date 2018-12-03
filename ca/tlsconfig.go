package ca

import (
	"crypto/tls"

	"github.com/karlseguin/ccache"
)

// TLSConfig is a wrapper over *tls.Config. This wrapper is required
// because CA uses LRU cache to store generated certificates. API of
// this structure transparently handles transaction-style access to TLS
// certificate.
type TLSConfig struct {
	item ccache.TrackedItem
}

// Get returns stored TLS configuration instance.
func (t *TLSConfig) Get() *tls.Config {
	if t.item == nil {
		return nil
	}

	return t.item.Value().(*tls.Config)
}

// Release returns TLS configuration back to the LRU cache.
func (t *TLSConfig) Release() {
	if t.item != nil {
		t.item.Release()
	}
}
