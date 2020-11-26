package ca

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"runtime"

	"github.com/9seconds/httransform/v2/events"
	"github.com/OneOfOne/xxhash"
	lru "github.com/hashicorp/golang-lru"
)

type CA struct {
	workers    []worker
	lenWorkers uint64
}

func (c *CA) Get(host string) (*tls.Config, error) {
	chosenWorker := xxhash.ChecksumString64(host) % c.lenWorkers

	conf, err := c.workers[int(chosenWorker)].Get(host)
	if err != nil {
		return nil, fmt.Errorf("cannot get tls config for host: %w", err)
	}

	return conf, nil
}

func NewCA(ctx context.Context,
	channelEvents events.EventChannel,
	certCA []byte,
	privateKey []byte,
	cacheSize int) (*CA, error) {
	ca, err := tls.X509KeyPair(certCA, privateKey)
	if err != nil {
		return nil, fmt.Errorf("cannot make a x509 keypair: %w", err)
	}

	if ca.Leaf, err = x509.ParseCertificate(ca.Certificate[0]); err != nil {
		return nil, fmt.Errorf("invalid certificates: %w", err)
	}

	cache, err := lru.NewWithEvict(cacheSize, func(key, _ interface{}) {
		select {
		case <-ctx.Done():
		case channelEvents <- events.AcquireEvent(events.EventTypeDropCertificate, key, key.(string)):
		}
	})
	if err != nil {
		return nil, fmt.Errorf("cannot build a new cache: %w", err)
	}

	obj := &CA{
		workers:    make([]worker, 0, runtime.NumCPU()),
		lenWorkers: uint64(runtime.NumCPU()),
	}

	for i := 0; i < runtime.NumCPU(); i++ {
		wrk := worker{
			ca:              ca,
			ctx:             ctx,
			cache:           cache,
			channelEvents:   channelEvents,
			channelRequests: make(chan workerRequest),
		}

		go wrk.Run()

		obj.workers = append(obj.workers, wrk)
	}

	return obj, nil
}
