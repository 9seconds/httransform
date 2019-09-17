package client

import (
	"net"
	"time"

	"golang.org/x/xerrors"
)

type releaseRequest struct {
	conn net.Conn
	addr string
}

type dialRequest struct {
	response chan<- getConnResponse
	addr     string
}

// PooledDialer is a dialer which manages pools of connections to
// different hosts.
type PooledDialer struct {
	base             BaseDialer
	addresses        map[string]*conns
	timeout          time.Duration
	limit            int
	dialRequests     chan dialRequest
	releaseRequests  chan releaseRequest
	closedRequests   chan string
	obsoleteRequests chan string
}

// Dial establishes connection to the given host.
func (d *PooledDialer) Dial(addr string) (net.Conn, error) {
	response := make(chan getConnResponse)
	d.dialRequests <- dialRequest{
		addr:     addr,
		response: response,
	}
	item := <-response
	return item.conn, item.err
}

// Release returns net.Conn back to the client. Second parameter is
// hostname which was used for Dial.
func (d *PooledDialer) Release(conn net.Conn, addr string) {
	d.releaseRequests <- releaseRequest{
		conn: conn,
		addr: addr,
	}
}

// NotifyClosed tells Dialer that net.Conn for address is broken.
func (d *PooledDialer) NotifyClosed(addr string) {
	d.closedRequests <- addr
}

// Run runs background worker.
func (d *PooledDialer) Run() {
	for {
		select {
		case req := <-d.dialRequests:
			addresses, ok := d.addresses[req.addr]
			if !ok {
				addresses = newConns(req.addr, d.base, d.timeout, d.limit, d.obsoleteRequests)
				d.addresses[req.addr] = addresses
				go addresses.run()
			}

			channel, err := addresses.getResponseChan(time.Second)
			if err != nil {
				req.response <- getConnResponse{err: err}
				continue
			}
			req.response <- <-channel

		case req := <-d.releaseRequests:
			if addresses, ok := d.addresses[req.addr]; ok {
				addresses.put(req.conn)
			}

		case req := <-d.closedRequests:
			if addresses, ok := d.addresses[req]; ok {
				addresses.notifyClosed()
			}

		case req := <-d.obsoleteRequests:
			if addresses, ok := d.addresses[req]; ok && addresses.isObsolete() {
				addresses.stop()
				delete(d.addresses, req)
			}
		}
	}
}

// NewPooledDialer creates new instance of PooledDialer.
func NewPooledDialer(dialer BaseDialer, timeout time.Duration, limit int) (*PooledDialer, error) {
	if timeout == time.Duration(0) {
		timeout = DefaultDialTimeout
	}
	if limit == 0 {
		limit = DefaultLimit
	} else if limit < 0 {
		return nil, xerrors.New("limit should be >= 0")
	}

	return &PooledDialer{
		base:             dialer,
		addresses:        make(map[string]*conns),
		timeout:          timeout,
		limit:            limit,
		dialRequests:     make(chan dialRequest),
		releaseRequests:  make(chan releaseRequest),
		closedRequests:   make(chan string),
		obsoleteRequests: make(chan string),
	}, nil
}
