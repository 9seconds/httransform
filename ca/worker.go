package ca

import (
	"context" // nolint: gosec
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"net"
	"sync"
	"time"

	"zvelo.io/ttlru"
)

var (
	workerBigBangMoment    = time.Unix(0, 0)
	workerDefaultTLSConfig = &tls.Config{
		InsecureSkipVerify: true,
	}
)

type worker struct {
	orgName string
	ca      tls.Certificate
	cache   ttlru.Cache
	secret  []byte
	ctx     context.Context

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
				item = w.createConfig(req.host)
				w.cache.Set(req.host, item)
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
		return nil, ErrContextClosed
	}
}

func (w *worker) createConfig(host string) *tls.Config {
	template := x509.Certificate{
		SerialNumber:          &big.Int{},
		Issuer:                w.ca.Leaf.Subject,
		Subject:               pkix.Name{Organization: []string{w.orgName}},
		NotBefore:             workerBigBangMoment,
		NotAfter:              time.Now().Add(WorkerCertificateTTL),
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

	randomness := make([]byte, 0, 16)
	rand.Read(randomness)
	template.SerialNumber.SetBytes(randomness)

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
	config := workerDefaultTLSConfig.Clone()
	config.Certificates = append(config.Certificates, certificate)

	return config
}
