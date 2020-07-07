package dialers

import (
	"context"
	"net"
	"time"

	"github.com/libp2p/go-reuseport"
)

type pooledDialerGetRequest struct {
	addr     string
	ctx      context.Context
	response chan<- pooledDialerGetResponse
}

type pooledDialerGetResponse struct {
	conn *pooledConn
	err  error
}

type pooledDialer struct {
	baseDialer

	pools              map[string]*connectionPool
	channelGetRequests chan pooledDialerGetRequest
}

func (p *pooledDialer) Dial(ctx context.Context, addr string) (Conn, error) {
	response := make(chan pooledDialerGetResponse)
	req := pooledDialerGetRequest{
		addr:     addr,
		ctx:      ctx,
		response: response,
	}

	select {
	case <-ctx.Done():
		return nil, ErrContextClosed
	case <-p.ctx.Done():
		return nil, ErrContextClosed
	case p.channelGetRequests <- req:
		resp := <-response
		return resp.conn, resp.err
	}
}

func (p *pooledDialer) run(gcEvery, staleAfter time.Duration) {
	ticker := time.NewTicker(gcEvery)
	defer ticker.Stop()

	for {
		select {
		case <-p.ctx.Done():
			for k, v := range p.pools {
				v.shutdown()
				delete(p.pools, k)
			}
			return
		case req := <-p.channelGetRequests:
			pool, ok := p.pools[req.addr]
			if !ok {
				pool = newConnectionPool(p.ctx, p)
				p.pools[req.addr] = pool
			}
			pool.process(req)
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
		},
		pools:              map[string]*connectionPool{},
		channelGetRequests: make(chan pooledDialerGetRequest),
	}

	go rv.run(PooledDialerGCEvery, PooledDialerStaleAfter)

	return rv
}
