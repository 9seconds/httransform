package ca

import (
	"crypto/hmac"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/binary"
	"math/big"
	"math/rand"
	"net"
	"time"

	"github.com/juju/errors"
	"github.com/karlseguin/ccache"
)

const (
	defaultTTLForCertificate = 10 * time.Minute
	defaultRSAKeyLength      = 1024
)

type CA struct {
	ca          tls.Certificate
	orgNames    []string
	secret      []byte
	cache       *ccache.Cache
	workerQueue chan *signRequest
}

func (c *CA) Get(host string) (TLSConfig, error) {
	item := c.cache.TrackingGet(host)

	if item == ccache.NilTracked {
		request := signRequestPool.Get().(*signRequest)
		defer signRequestPool.Put(request)

		request.host = host
		c.workerQueue <- request

		response := <-request.responseChan
		defer signResponsePool.Put(response)

		if response.err != nil {
			return TLSConfig{}, errors.Annotate(response.err, "Cannot create Certificate for the host")
		}

		item = response.item
	}

	return TLSConfig{item}, nil
}

func (c *CA) worker() {
	for request := range c.workerQueue {
		response := signResponsePool.Get().(*signResponse)
		response.item = nil
		response.err = nil

		item := c.cache.TrackingGet(request.host)
		if item != ccache.NilTracked {
			response.item = item
			request.responseChan <- response
			continue
		}

		cert, err := c.sign(request.host)
		if err != nil {
			response.err = err
			request.responseChan <- response
			continue
		}

		conf := &tls.Config{InsecureSkipVerify: true}
		conf.Certificates = append(conf.Certificates, cert)
		c.cache.Set(request.host, conf, defaultTTLForCertificate)
		response.item = c.cache.TrackingGet(request.host)
		request.responseChan <- response
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
		ca:          ca,
		secret:      certKey,
		orgNames:    orgNames,
		cache:       ccache.New(ccache.Configure().MaxSize(cacheMaxSize).ItemsToPrune(cacheItemsToPrune)),
		workerQueue: make(chan *signRequest),
	}
	go obj.worker()

	return obj, nil
}
