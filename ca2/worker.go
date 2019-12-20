package ca2

import (
	"context"
	"crypto/hmac"
	"crypto/md5" // nolint: gosec
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"errors"
	"math/big"
	"net"
	"sync"
	"time"

	lru "github.com/hashicorp/golang-lru"
)

var (
	bigBangTime = time.Unix(0, 0)

	DefaultTLSConfig = &tls.Config{
		InsecureSkipVerify: true, // nolint: gosec
	}
)

// RSAKeyLength defines a length of the key to generate
const RSAKeyLength = 2048

type workerRequest struct {
	host     string
	response chan<- *tls.Config
}

type worker struct {
	ca       tls.Certificate
	cache    *lru.Cache
	orgNames []string
	secret   []byte
	ctx      context.Context
	metrics  CertificateMetrics

	channelRequests chan workerRequest
}

func (w *worker) run(wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case <-w.ctx.Done():
			return
		case req := <-w.channelRequests:
			item, ok := w.cache.Get(req.host)
			if !ok {
				item = w.makeConfig(req.host)
				w.cache.Add(req.host, item)
			}

			req.response <- item.(*tls.Config)
			close(req.response)
		}
	}
}

func (w *worker) get(host string) (*tls.Config, error) {
	response := make(chan *tls.Config)
	req := workerRequest{
		host:     host,
		response: response,
	}

	select {
	case w.channelRequests <- req:
		return <-response, nil
	case <-w.ctx.Done():
		return nil, errors.New("context is closed")
	}
}

func (w *worker) makeConfig(host string) *tls.Config {
	template := x509.Certificate{
		SerialNumber:          &big.Int{},
		Issuer:                w.ca.Leaf.Subject,
		Subject:               pkix.Name{Organization: w.orgNames},
		NotBefore:             bigBangTime,
		NotAfter:              timeNotAfter(),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	if ip := net.ParseIP(host); ip != nil {
		template.IPAddresses = append(template.IPAddresses, ip)
	} else {
		template.DNSNames = append(template.DNSNames, host)
		template.Subject.CommonName = host
	}

	hash := hmac.New(md5.New, w.secret)
	hash.Write([]byte(host)) // nolint: errcheck

	template.SerialNumber.SetBytes(hash.Sum(nil))

	certPriv, err := rsa.GenerateKey(rand.Reader, RSAKeyLength)
	if err != nil {
		panic(err)
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, w.ca.Leaf,
		&certPriv.PublicKey, w.ca.PrivateKey)
	if err != nil {
		panic(err)
	}

	certificate := tls.Certificate{
		Certificate: [][]byte{derBytes, w.ca.Certificate[0]},
		PrivateKey:  certPriv,
	}
	config := DefaultTLSConfig.Clone()
	config.Certificates = append(config.Certificates, certificate)

	w.metrics.NewCertificate()

	return config
}

func timeNotAfter() time.Time {
	now := time.Now()
	return time.Date(now.Year()+10, now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
}
