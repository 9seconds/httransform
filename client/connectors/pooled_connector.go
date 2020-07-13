package connectors

import (
	"context"
	"time"
)

type pooledConnectorGetRequest struct {
	addr     string
	ctx      context.Context
	response chan<- pooledConnectorGetResponse
}

type pooledConnectorGetResponse struct {
	conn *pooledConn
	err  error
}

type pooledConnector struct {
	baseConnector

	pools              map[string]*pooledConnectorConnectionPool
	channelGetRequests chan pooledConnectorGetRequest
}

func (p *pooledConnector) Connect(ctx context.Context, addr string) (Conn, error) {
	response := make(chan pooledConnectorGetResponse)
	req := pooledConnectorGetRequest{
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

func (p *pooledConnector) run(gcEvery, staleAfter time.Duration) {
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

func NewPooledConnector(ctx context.Context, dialer Dialer) Connector {
	ctx, cancel := context.WithCancel(ctx)
	rv := &pooledConnector{
		baseConnector: baseConnector{
			dialer: dialer,
			ctx:    ctx,
			cancel: cancel,
		},
		pools:              map[string]*pooledConnectorConnectionPool{},
		channelGetRequests: make(chan pooledConnectorGetRequest),
	}

	go rv.run(PooledConnectorConnectionPoolGCEvery, PooledConnectorConnectionPoolStaleAfter)

	return rv
}
