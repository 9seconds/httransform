package ca

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"hash/fnv"
	"io"
	"runtime"
	"sync"
	"time"

	"github.com/PumpkinSeed/errors"
	"zvelo.io/ttlru"
)

const (
	caDefaultMaxSize = 1024
	caCertificateTTL = 24 * time.Hour
)

var ErrCAInvalidCertificates = errors.New("invalid ca certificate")

type CA struct {
	cancel  context.CancelFunc
	cache   ttlru.Cache
	workers []worker
	wg      sync.WaitGroup
}

func (c *CA) Get(host string) (*tls.Config, error) {
	if item, ok := c.cache.Get(host); ok {
		return item.(*tls.Config), nil
	}

	hashFunc := fnv.New32a()
	io.WriteString(hashFunc, host)

	num := int(hashFunc.Sum32() % uint32(len(c.workers)))

	return c.workers[num].get(host)
}

func (c *CA) Close() {
	c.cancel()
	c.wg.Wait()
}

func NewCA(ctx context.Context, certCA, certKey []byte, maxSize int, orgName string) (*CA, error) {
	ca, err := tls.X509KeyPair(certCA, certKey)
	if err != nil {
		return nil, errors.Wrap(err, ErrCAInvalidCertificates)
	}

	if ca.Leaf, err = x509.ParseCertificate(ca.Certificate[0]); err != nil {
		return nil, errors.Wrap(err, ErrCAInvalidCertificates)
	}

	if maxSize <= 0 {
		maxSize = caDefaultMaxSize
	}

	ctx, cancel := context.WithCancel(ctx)
	obj := &CA{
		cancel:  cancel,
		cache:   ttlru.New(maxSize, ttlru.WithTTL(caCertificateTTL)),
		workers: make([]worker, runtime.NumCPU()),
	}

	obj.wg.Add(len(obj.workers))

	for i := range obj.workers {
		obj.workers[i] = worker{
			ca:              ca,
			cache:           obj.cache,
			orgName:         orgName,
			ctx:             ctx,
			channelRequests: make(chan workerRequest),
		}
		go obj.workers[i].run(&obj.wg)
	}

	return obj, nil
}
