package ca

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha512"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/binary"
	"math/big"
	"net"
	"time"

	"github.com/9seconds/httransform/v2/events"
	lru "github.com/hashicorp/golang-lru"
)

const RSAKeyLength = 2048

type workerRequest struct {
	host     string
	response chan<- *tls.Config
}

type worker struct {
	ca              tls.Certificate
	ctx             context.Context
	cache           *lru.Cache
	channelEvents   chan<- events.Event
	channelRequests chan workerRequest
}

func (w *worker) Get(host string) (*tls.Config, error) {
	if cert, ok := w.cache.Get(host); ok {
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

			if cert, ok := w.cache.Get(req.host); ok {
				conf = cert.(*tls.Config)
			} else {
				conf = w.process(req.host)
				w.cache.Add(req.host, conf)

				select {
				case <-w.ctx.Done():
				case w.channelEvents <- events.New(events.EventTypeNewCertificate, req.host):
				}
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
		NotBefore:             adjustDate(now, -1),
		NotAfter:              adjustDate(now, 1),
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

	hash := sha512.New()

	hash.Write([]byte(host))
	binary.Write(hash, binary.LittleEndian, now.UnixNano())
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

	return &tls.Config{
		InsecureSkipVerify: true,
		Certificates:       []tls.Certificate{certificate},
	}
}

func adjustDate(now time.Time, incr int) time.Time {
	return time.Date(now.Year()+incr, now.Month(), 0, 0, 0, 0, 0, now.Location())
}
