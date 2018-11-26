package ca

import (
	"crypto/tls"

	"github.com/karlseguin/ccache"
)

type TLSConfig struct {
	item ccache.TrackedItem
}

func (t *TLSConfig) Get() *tls.Config {
	return t.item.Value().(*tls.Config)
}

func (t *TLSConfig) Release() {
	t.item.Release()
}
