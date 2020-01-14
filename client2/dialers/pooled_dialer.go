package dialers

import (
	"context"
	"net"
	"time"

	reuseport "github.com/libp2p/go-reuseport"
	"github.com/rs/dnscache"
)

const (
	PooledDialerTimeout = 30 * time.Second

	pooledDialerGCEvery    = time.Second
	pooledDialerStaleAfter = time.Minute
)

type pooledDialerResponse struct {
	conn *pooledConn
	err  error
}

type pooledDialerDialRequest struct {
	addr     string
	ctx      context.Context
	response chan<- pooledDialerResponse
}

type pooledDialer struct {
	baseDialer

	pools               map[string]*connectionPool
	channelDialRequests chan pooledDialerDialRequest
}

func (p *pooledDialer) Dial(ctx context.Context, addr string) (Conn, error) {
	resp := make(chan pooledDialerResponse)
	req := pooledDialerDialRequest{
		addr:     addr,
		ctx:      ctx,
		response: resp,
	}

	select {
	case <-ctx.Done():
		return nil, ErrPooledDialerCanceled
	case <-p.ctx.Done():
		return nil, ErrPooledDialerCanceled
	case p.channelDialRequests <- req:
		rr := <-resp
		return rr.conn, rr.err
	}
}

func (p *pooledDialer) run(gcEvery, staleAfter time.Duration) {
	go p.baseDialer.run()

	ticker := time.NewTicker(gcEvery)
	defer ticker.Stop()

	for {
		select {
		case <-p.ctx.Done():
			for _, v := range p.pools {
				v.shutdown()
			}

			return
		case req := <-p.channelDialRequests:
			pool, ok := p.pools[req.addr]
			if !ok {
				pool = newConnectionPool(p.ctx, p)
				p.pools[req.addr] = pool
			}

			pool.get(req)
		case <-ticker.C:
			for k, v := range p.pools {
				if time.Since(v.getLastRequestAt()) > staleAfter {
					v.shutdown()
					delete(p.pools, k)
				}
			}
		}
	}
}

func NewPooledDialer(ctx context.Context, timeout time.Duration) Dialer {
	if timeout == 0 {
		timeout = PooledDialerTimeout
	}

	ctx, cancel := context.WithCancel(ctx)
	rv := &pooledDialer{
		baseDialer: baseDialer{
			dialer: net.Dialer{
				Timeout: timeout,
				Control: reuseport.Control,
			},
			ctx:    ctx,
			cancel: cancel,
			dns:    dnscache.Resolver{},
		},
		pools:               make(map[string]*connectionPool),
		channelDialRequests: make(chan pooledDialerDialRequest),
	}

	go rv.run(pooledDialerGCEvery, pooledDialerStaleAfter)

	return rv
}
