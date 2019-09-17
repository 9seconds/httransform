package ca

import (
	"crypto/hmac"
	"crypto/rsa"
	"crypto/sha1" // nolint: gosec
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/binary"
	"hash"
	"math/big"
	"math/rand"
	"net"
	"runtime"
	"sync"
	"time"

	"github.com/karlseguin/ccache"
	"golang.org/x/xerrors"
)

const (
	// TTLForCertificate is minimal time until certificate is
	// considered as expired. This is a time from the last usage. If the
	// certificate is expired, it does not necessary wiped out from LRU
	// cache.
	TTLForCertificate = 10 * time.Minute

	// RSAKeyLength is the bit size of RSA private key for the certificate.
	RSAKeyLength = 2048
)

var (
	certWorkerCount = uint32(runtime.NumCPU())
	bigBangTime     = time.Unix(0, 0)
)

// CertificateMetrics is a subset of the main Metrics interface which
// provides callbacks for certificates.
type CertificateMetrics interface {
	NewCertificate()
	DropCertificate()
}

// CA is a datastructure which presents TLS CA (certificate authority).
// The main purpose of this type is to generate TLS certificates
// on-the-fly, using given CA certificate and private key.
//
// CA generates certificates concurrently but in thread-safe way. The
// number of concurrently generated certificates is equal to the number
// of CPUs.
type CA struct {
	ca           tls.Certificate
	orgNames     []string
	secret       []byte
	requestChans []chan *signRequest
	cache        *ccache.Cache
	wg           *sync.WaitGroup
	metrics      CertificateMetrics
}

// Get returns generated TLSConfig instance for the given hostname.
func (c *CA) Get(host string) (TLSConfig, error) {
	item := c.cache.TrackingGet(host)

	if item == ccache.NilTracked {
		newRequest := signRequestPool.Get().(*signRequest)
		defer signRequestPool.Put(newRequest)

		newRequest.host = host
		c.getWorkerChan(host) <- newRequest
		response := <-newRequest.response
		defer signResponsePool.Put(response)

		if response.err != nil {
			return TLSConfig{}, xerrors.Errorf("cannot create TLS certificate for host %s: %w",
				host, response.err)
		}

		item = response.item
	}

	return TLSConfig{item}, nil
}

// Close stops CA instance. This includes all signing workers and LRU
// cache.
func (c *CA) Close() error {
	for _, ch := range c.requestChans {
		close(ch)
	}
	c.wg.Wait()
	c.cache.Stop()

	return nil
}

func (c *CA) worker(requests chan *signRequest, wg *sync.WaitGroup) {
	defer wg.Done()

	for req := range requests {
		resp := signResponsePool.Get().(*signResponse)
		resp.err = nil

		if item := c.cache.TrackingGet(req.host); item != ccache.NilTracked {
			resp.item = item
			req.response <- resp
			continue
		}

		cert, err := c.sign(req.host)
		if err != nil {
			resp.err = err
			req.response <- resp
			continue
		}
		c.metrics.NewCertificate()

		conf := &tls.Config{InsecureSkipVerify: true} // nolint: gosec
		conf.Certificates = append(conf.Certificates, cert)
		c.cache.Set(req.host, conf, TTLForCertificate)
		resp.item = c.cache.TrackingGet(req.host)
		req.response <- resp
	}
}

func (c *CA) sign(host string) (tls.Certificate, error) {
	template := x509.Certificate{
		SerialNumber:          &big.Int{},
		Issuer:                c.ca.Leaf.Subject,
		Subject:               pkix.Name{Organization: c.orgNames},
		NotBefore:             bigBangTime,
		NotAfter:              timeNotAfter(),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	hash := hmac.New(sha1.New, c.secret)
	hash.Write([]byte(host)) // nolint: errcheck
	if ip := net.ParseIP(host); ip != nil {
		template.IPAddresses = append(template.IPAddresses, ip)
	} else {
		template.DNSNames = append(template.DNSNames, host)
		template.Subject.CommonName = host
	}
	hashed := hash.Sum(nil)
	template.SerialNumber.SetBytes(hashed)
	hash.Write(c.secret) // nolint: errcheck

	randSeed := int64(binary.LittleEndian.Uint64(hash.Sum(nil)[:8]))
	randGen := rand.New(rand.NewSource(randSeed))

	certpriv, err := rsa.GenerateKey(randGen, RSAKeyLength)
	if err != nil {
		panic(err)
	}
	derBytes, err := x509.CreateCertificate(randGen, &template, c.ca.Leaf,
		&certpriv.PublicKey, c.ca.PrivateKey)
	if err != nil {
		return tls.Certificate{}, xerrors.Errorf("cannot generate TLS certificate: %w", err)
	}

	return tls.Certificate{
		Certificate: [][]byte{derBytes, c.ca.Certificate[0]},
		PrivateKey:  certpriv,
	}, nil
}

func (c *CA) getWorkerChan(host string) chan<- *signRequest {
	newHash := hashPool.Get().(hash.Hash32)
	newHash.Reset()
	newHash.Write([]byte(host)) // nolint: errcheck
	chanNumber := newHash.Sum32() % certWorkerCount
	hashPool.Put(newHash)

	return c.requestChans[chanNumber]
}

// NewCA creates new instance of TLS CA.
func NewCA(certCA, certKey []byte, metrics CertificateMetrics,
	cacheMaxSize int64, cacheItemsToPrune uint32, orgNames ...string) (CA, error) {
	ca, err := tls.X509KeyPair(certCA, certKey)
	if err != nil {
		return CA{}, xerrors.Errorf("invalid certificates: %w", err)
	}
	if ca.Leaf, err = x509.ParseCertificate(ca.Certificate[0]); err != nil {
		return CA{}, xerrors.Errorf("invalid certificates: %w", err)
	}

	ccacheConf := ccache.Configure()
	ccacheConf = ccacheConf.MaxSize(cacheMaxSize)
	ccacheConf = ccacheConf.ItemsToPrune(cacheItemsToPrune)
	ccacheConf = ccacheConf.OnDelete(func(_ *ccache.Item) { metrics.DropCertificate() })

	obj := CA{
		ca:           ca,
		metrics:      metrics,
		secret:       certKey,
		orgNames:     orgNames,
		cache:        ccache.New(ccacheConf),
		requestChans: make([]chan *signRequest, 0, certWorkerCount),
		wg:           &sync.WaitGroup{},
	}
	for i := 0; i < int(certWorkerCount); i++ {
		newChan := make(chan *signRequest)
		obj.requestChans = append(obj.requestChans, newChan)
		obj.wg.Add(1)
		go obj.worker(newChan, obj.wg)
	}

	return obj, nil
}

func timeNotAfter() time.Time {
	now := time.Now()
	return time.Date(now.Year()+10, now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
}
