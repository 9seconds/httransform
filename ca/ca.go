package ca

import (
	"crypto/hmac"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/binary"
	"hash"
	"math/big"
	"math/rand"
	"net"
	"runtime"
	"time"

	"github.com/juju/errors"
	"github.com/karlseguin/ccache"
)

const (
	defaultTTLForCertificate = 10 * time.Minute
	defaultRSAKeyLength      = 1024
)

var certWorkerCount uint32

type CA struct {
	ca           tls.Certificate
	orgNames     []string
	secret       []byte
	cache        *ccache.Cache
	requestChans []chan *signRequest
}

func (c *CA) Get(host string) (TLSConfig, error) {
	item := c.cache.TrackingGet(host)

	if item == ccache.NilTracked {
		newHash := hashPool.Get().(hash.Hash32)
		defer hashPool.Put(newHash)

		newRequest := signRequestPool.Get().(*signRequest)
		defer signRequestPool.Put(newRequest)

		newHash.Reset()
		newHash.Write([]byte(host))
		chanNumber := newHash.Sum32() % certWorkerCount

		newRequest.host = host
		c.requestChans[chanNumber] <- newRequest
		response := <-newRequest.response
		defer signResponsePool.Put(response)

		item = response.item
	}

	return TLSConfig{item}, nil
}

func (c *CA) worker(requests chan *signRequest) {
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

		conf := &tls.Config{InsecureSkipVerify: true}
		conf.Certificates = append(conf.Certificates, cert)
		c.cache.Set(req.host, conf, defaultTTLForCertificate)
		resp.item = c.cache.TrackingGet(req.host)
		req.response <- resp
	}
}

func (c *CA) sign(host string) (tls.Certificate, error) {
	template := x509.Certificate{
		SerialNumber:          &big.Int{},
		Issuer:                c.ca.Leaf.Subject,
		Subject:               pkix.Name{Organization: c.orgNames},
		NotBefore:             time.Unix(0, 0),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	hash := hmac.New(sha1.New, c.secret)
	hash.Write([]byte(host))
	if ip := net.ParseIP(host); ip != nil {
		template.IPAddresses = append(template.IPAddresses, ip)
	} else {
		template.DNSNames = append(template.DNSNames, host)
		template.Subject.CommonName = host
	}
	hashed := hash.Sum(nil)
	template.SerialNumber.SetBytes(hashed)
	hash.Write(c.secret)

	randSeed := int64(binary.LittleEndian.Uint64(hash.Sum(nil)[:8]))
	randGen := rand.New(rand.NewSource(randSeed))

	certpriv, err := rsa.GenerateKey(randGen, defaultRSAKeyLength)
	if err != nil {
		panic(err)
	}
	derBytes, err := x509.CreateCertificate(randGen, &template, c.ca.Leaf,
		&certpriv.PublicKey, c.ca.PrivateKey)
	if err != nil {
		return tls.Certificate{}, errors.Annotate(err, "Cannot generate TLS certificate")
	}

	return tls.Certificate{
		Certificate: [][]byte{derBytes, c.ca.Certificate[0]},
		PrivateKey:  certpriv,
	}, nil
}

func NewCA(certCA, certKey []byte, cacheMaxSize int64, cacheItemsToPrune uint32,
	orgNames ...string) (CA, error) {
	ca, err := tls.X509KeyPair(certCA, certKey)
	if err != nil {
		return CA{}, errors.Annotate(err, "Invalid certificates")
	}
	if ca.Leaf, err = x509.ParseCertificate(ca.Certificate[0]); err != nil {
		return CA{}, errors.Annotate(err, "Invalid certificates")
	}

	obj := CA{
		ca:           ca,
		secret:       certKey,
		orgNames:     orgNames,
		cache:        ccache.New(ccache.Configure().MaxSize(cacheMaxSize).ItemsToPrune(cacheItemsToPrune)),
		requestChans: make([]chan *signRequest, 0, certWorkerCount),
	}
	for i := 0; i < int(certWorkerCount); i++ {
		newChan := make(chan *signRequest)
		obj.requestChans = append(obj.requestChans, newChan)
		go obj.worker(newChan)
	}

	return obj, nil
}

func init() {
	certWorkerCount = uint32(runtime.NumCPU())
}
