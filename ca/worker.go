package ca

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"net"
	"time"

	"github.com/9seconds/httransform/v2/cache"
	"github.com/9seconds/httransform/v2/events"
)

// RSAKeyLength defines a bit length of generated RSA key. This is a
// good default for fake certificates, you usually do not need anything
// more than that.
const RSAKeyLength = 2048

type workerRequest struct {
	host     string
	response chan<- *tls.Config
}

type worker struct {
	ca              tls.Certificate
	ctx             context.Context
	cache           cache.Interface
	channelEvents   events.Channel
	channelRequests chan workerRequest
}

func (w *worker) Get(host string) (*tls.Config, error) {
	if cert := w.cache.Get(host); cert != nil {
		return cert.(*tls.Config), nil
	}

	response := make(chan *tls.Config)
	req := workerRequest{
		host:     host,
		response: response,
	}

	select {
	case <-w.ctx.Done():
		return nil, ErrContextClosed
	case w.channelRequests <- req:
		return <-response, nil
	}
}

func (w *worker) Run() {
	for {
		select {
		case <-w.ctx.Done():
			return
		case req := <-w.channelRequests:
			var conf *tls.Config

			if cert := w.cache.Get(req.host); cert != nil {
				conf = cert.(*tls.Config)
			} else {
				conf = w.process(req.host)
				w.cache.Add(req.host, conf)

				w.channelEvents.Send(w.ctx, events.EventTypeNewCertificate, req.host, req.host)
			}

			req.response <- conf
			close(req.response)
		}
	}
}

func (w *worker) process(host string) *tls.Config {
	now := time.Now()

	template := x509.Certificate{
		SerialNumber:          &big.Int{},
		Issuer:                w.ca.Leaf.Subject,
		Subject:               pkix.Name{},
		NotBefore:             now.AddDate(0, 0, -1),
		NotAfter:              now.AddDate(0, 3, 0),
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

	randBytes := make([]byte, 64)
	rand.Read(randBytes) // nolint: errcheck
	template.SerialNumber.SetBytes(randBytes)

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

	return &tls.Config{
		ServerName:         host,
		InsecureSkipVerify: true, // nolint: gosec
		Certificates:       []tls.Certificate{certificate},
	}
}
