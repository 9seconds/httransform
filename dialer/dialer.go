package dialer

import (
	"net"
	"time"

	"github.com/juju/errors"
	"github.com/valyala/fasthttp"
)

const (
	DefaultDialTimeout = 20 * time.Second
	DefaultLimit       = 1024
)

type BaseDialer func(string, time.Duration) (net.Conn, error)

type releaseRequest struct {
	conn net.Conn
	addr string
}

type dialRequest struct {
	response chan<- getConnResponse
	addr     string
}

type Dialer struct {
	base             BaseDialer
	addresses        map[string]*conns
	timeout          time.Duration
	limit            int
	dialRequests     chan dialRequest
	releaseRequests  chan releaseRequest
	closedRequests   chan string
	obsoleteRequests chan string
}

func (d *Dialer) Dial(addr string) (net.Conn, error) {
	response := make(chan getConnResponse)
	d.dialRequests <- dialRequest{
		addr:     addr,
		response: response,
	}
	item := <-response
	return item.conn, item.err
}

func (d *Dialer) Release(conn net.Conn, addr string) {
	d.releaseRequests <- releaseRequest{
		conn: conn,
		addr: addr,
	}
}

func (d *Dialer) NotifyClosed(addr string) {
	d.closedRequests <- addr
}

func (d *Dialer) run() {
	for {
		select {
		case req := <-d.dialRequests:
			connections := d.getConnections(req.addr)
			channel, err := connections.getResponseChan(time.Second)
			if err != nil {
				req.response <- getConnResponse{err: err}
				continue
			}
			req.response <- <-channel

		case req := <-d.releaseRequests:
			d.getConnections(req.addr).put(req.conn)

		case req := <-d.closedRequests:
			d.getConnections(req).notifyClosed()

		case req := <-d.obsoleteRequests:
			connections := d.getConnections(req)
			if connections.isObsolete() {
				connections.stop()
				delete(d.addresses, req)
			}
		}
	}
}

func (d *Dialer) getConnections(addr string) *conns {
	addresses, ok := d.addresses[addr]
	if !ok {
		addresses = newConns(addr, d.base, d.timeout, d.limit, d.obsoleteRequests)
		d.addresses[addr] = addresses
	}
	return addresses
}

func NewDialer(dialer BaseDialer, timeout time.Duration, limit int) (*Dialer, error) {
	if timeout == time.Duration(0) {
		timeout = DefaultDialTimeout
	}
	if limit == 0 {
		limit = DefaultLimit
	} else if limit < 0 {
		return nil, errors.New("Limit should be >= 0")
	}

	return &Dialer{
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

func FastHTTPBaseDialer(addr string, dialTimeout time.Duration) (net.Conn, error) {
	return fasthttp.DialDualStackTimeout(addr, dialTimeout)
}

func StdBaseDialer(addr string, dialTimeout time.Duration) (net.Conn, error) {
	return (&net.Dialer{
		Timeout:   dialTimeout,
		DualStack: true,
	}).Dial("tcp", addr)
}
