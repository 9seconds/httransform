package ca

import "crypto/tls"

type workerRequest struct {
	host     string
	response chan<- *tls.Config
}
