package ca

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"hash/fnv"
	"runtime"

	"github.com/9seconds/httransform/v2/events"
	lru "github.com/hashicorp/golang-lru"
)

type CA struct {
	workers []worker
}

func (c *CA) Get(host string) (*tls.Config, error) {
	hashFunc := fnv.New64a()
	hashFunc.Write([]byte(host))

	chosenWorker := int(hashFunc.Sum64() % uint64(len(c.workers)))

	conf, err := c.workers[chosenWorker].Get(host)
	if err != nil {
		return nil, fmt.Errorf("cannot get tls config for host: %w", err)
	}

	return conf, nil
}

func NewCA(ctx context.Context,
	channelEvents chan<- events.Event,
	certCA []byte,
	privateKey []byte,
	orgName string,
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
		case channelEvents <- events.New(events.EventTypeDropCertificate, key):
		}
	})
	if err != nil {
		return nil, fmt.Errorf("cannot build a new cache: %w", err)
	}

	obj := &CA{
		workers: make([]worker, 0, runtime.NumCPU()),
	}

	for i := 0; i < runtime.NumCPU(); i++ {
		wrk := worker{
			ca:              ca,
			ctx:             ctx,
			cache:           cache,
			orgName:         orgName,
			channelEvents:   channelEvents,
			channelRequests: make(chan workerRequest),
		}

		go wrk.Run()

		obj.workers = append(obj.workers, wrk)
	}

	return obj, nil
}
