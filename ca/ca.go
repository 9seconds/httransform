package ca

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"runtime"
	"time"

	"github.com/9seconds/httransform/v2/cache"
	"github.com/9seconds/httransform/v2/events"
	"github.com/OneOfOne/xxhash"
)

const (
	// CACacheSize defines a size of LFU cache to use. Each hostname
	// corresponds to a certain entry in this cache and each hostname is
	// responsible for a single generated certificate.
	//
	// It may sound scary but scales well on practice. Usually you do
	// not need to alter this parameter. Please remember we talk about
	// LFU cache.
	CACacheSize = 1024

	// CACacheTTL defines TTL for each generated TLS certificate.
	// Actually, this parameter can be up to 3 months but it will be
	// better to regenerate it more frequently.
	CACacheTTL = 7 * 24 * time.Hour
)

// CA defines an authority which generates TLS certificates for given
// hostnames.
type CA struct {
	workers    []worker
	lenWorkers uint64
}

// Get returns tls.Config instance for the given hostname.
func (c *CA) Get(host string) (*tls.Config, error) {
	chosenWorker := xxhash.ChecksumString64(host) % c.lenWorkers

	conf, err := c.workers[int(chosenWorker)].Get(host)
	if err != nil {
		return nil, fmt.Errorf("cannot get tls config for host: %w", err)
	}

	return conf, nil
}

// NewCA generates CA instance.
func NewCA(ctx context.Context,
	channelEvents events.Channel,
	certCA []byte,
	privateKey []byte) (*CA, error) {
	ca, err := tls.X509KeyPair(certCA, privateKey)
	if err != nil {
		return nil, fmt.Errorf("cannot make a x509 keypair: %w", err)
	}

	if ca.Leaf, err = x509.ParseCertificate(ca.Certificate[0]); err != nil {
		return nil, fmt.Errorf("invalid certificates: %w", err)
	}

	obj := &CA{
		workers:    make([]worker, 0, runtime.NumCPU()),
		lenWorkers: uint64(runtime.NumCPU()),
	}
	cacheIf := cache.New(CACacheSize, CACacheTTL, func(key string, _ interface{}) {
		channelEvents.Send(ctx, events.EventTypeDropCertificate, key, key)
	})

	for i := 0; i < runtime.NumCPU(); i++ {
		wrk := worker{
			ca:              ca,
			ctx:             ctx,
			channelEvents:   channelEvents,
			cache:           cacheIf,
			channelRequests: make(chan workerRequest),
		}

		go wrk.Run()

		obj.workers = append(obj.workers, wrk)
	}

	return obj, nil
}
