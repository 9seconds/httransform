package ca

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"hash/fnv"
	"runtime"
	"sync"

	lru "github.com/hashicorp/golang-lru"
	"golang.org/x/xerrors"
)

// CertificateMetrics is a subset of the main Metrics interface which
// provides callbacks for certificates.
type CertificateMetrics interface {
	NewCertificate()
	DropCertificate()
}

// DefaultMaxSize defines a default value for TLS certificates to store
// in LRU cache.
const DefaultMaxSize = 1024

// CA is a datastructure which presents TLS CA (certificate authority).
// The main purpose of this type is to generate TLS certificates
// on-the-fly, using given CA certificate and private key.
//
// CA generates certificates concurrently but in thread-safe way. The
// number of concurrently generated certificates is equal to the number
// of CPUs.
type CA struct {
	cache   *lru.Cache
	cancel  context.CancelFunc
	workers []worker
	wg      sync.WaitGroup
}

func (c *CA) Get(host string) (*tls.Config, error) {
	if item, ok := c.cache.Get(host); ok {
		return item.(*tls.Config), nil
	}

	hashFunc := fnv.New32a()
	hashFunc.Write([]byte(host))
	num := int(hashFunc.Sum32() % uint32(len(c.workers)))

	return c.workers[num].get(host)
}

// Close stops CA instance. This includes all signing workers and LRU
// cache.
func (c *CA) Close() {
	c.cancel()
	c.wg.Wait()
	c.cache.Purge()
}

func NewCA(certCA, certKey []byte, metrics CertificateMetrics, maxSize int, orgNames []string) (*CA, error) {
	ca, err := tls.X509KeyPair(certCA, certKey)
	if err != nil {
		return nil, xerrors.Errorf("invalid certificates: %w", err)
	}

	if ca.Leaf, err = x509.ParseCertificate(ca.Certificate[0]); err != nil {
		return nil, xerrors.Errorf("invalid certificates: %w", err)
	}

	if maxSize <= 0 {
		maxSize = DefaultMaxSize
	}

	cache, err := lru.NewWithEvict(maxSize, func(_, _ interface{}) {
		metrics.DropCertificate()
	})
	if err != nil {
		return nil, xerrors.Errorf("cannot make a new cache: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	obj := &CA{
		cache:   cache,
		workers: make([]worker, runtime.NumCPU()),
		cancel:  cancel,
	}

	obj.wg.Add(len(obj.workers))

	for i := range obj.workers {
		obj.workers[i] = worker{
			ca:              ca,
			cache:           cache,
			orgNames:        orgNames,
			secret:          certKey,
			ctx:             ctx,
			metrics:         metrics,
			channelRequests: make(chan workerRequest),
		}
		go obj.workers[i].run(&obj.wg)
	}

	return obj, nil
}
