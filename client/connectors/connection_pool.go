package connectors

import (
	"context"
	"sync"
	"time"
)

type pooledConnectorConnectionPool struct {
	list                 pooledConnectorConnectionList
	connector            Connector
	ctx                  context.Context
	cancel               context.CancelFunc
	lastRequestAt        time.Time
	channelGetRequests   chan pooledConnectorGetRequest
	channelConnReturns   chan *pooledConn
	channelLastRequestAt chan chan<- time.Time
}

func (p *pooledConnectorConnectionPool) process(req pooledConnectorGetRequest) {
	select {
	case <-p.ctx.Done():
		req.response <- pooledConnectorGetResponse{err: ErrContextClosed}
	case p.channelGetRequests <- req:
	}
}

func (p *pooledConnectorConnectionPool) release(conn *pooledConn) bool {
	select {
	case <-p.ctx.Done():
		return false
	case p.channelConnReturns <- conn:
		return true
	}
}

func (p *pooledConnectorConnectionPool) getLastRequestAt() time.Time {
	response := make(chan time.Time)

	select {
	case <-p.ctx.Done():
		return time.Time{}
	case p.channelLastRequestAt <- response:
		return <-response
	}
}

func (p *pooledConnectorConnectionPool) run(gcEvery, staleAfter time.Duration) {
	ticker := time.NewTicker(gcEvery)
	defer ticker.Stop()

	for {
		select {
		case <-p.ctx.Done():
			for _, v := range p.list.conns {
				v.Close()
			}

			return
		case conn := <-p.channelConnReturns:
			conn.once = sync.Once{}
			conn.timestamp = time.Now()
			p.list.put(conn)
		case req := <-p.channelGetRequests:
			p.lastRequestAt = time.Now()
			if conn := p.list.get(); conn != nil {
				req.response <- pooledConnectorGetResponse{conn: conn}
			} else {
				conn, err := p.connector.Connect(req.ctx, req.addr)
				req.response <- pooledConnectorGetResponse{
					conn: &pooledConn{Conn: conn, pool: p},
					err:  err,
				}
			}

			close(req.response)
		case <-ticker.C:
			if conn := p.list.get(); conn != nil {
				if time.Since(conn.timestamp) > staleAfter {
					conn.Close()
				} else {
					p.list.put(conn)
				}
			}
		case response := <-p.channelLastRequestAt:
			response <- p.lastRequestAt
			close(response)
		}
	}
}

func (p *pooledConnectorConnectionPool) shutdown() {
	p.cancel()
}

func newConnectionPool(ctx context.Context, connector Connector) *pooledConnectorConnectionPool {
	ctx, cancel := context.WithCancel(ctx)
	rv := &pooledConnectorConnectionPool{
		list:                 pooledConnectorConnectionList{},
		connector:            connector,
		ctx:                  ctx,
		cancel:               cancel,
		channelGetRequests:   make(chan pooledConnectorGetRequest),
		channelConnReturns:   make(chan *pooledConn),
		channelLastRequestAt: make(chan chan<- time.Time),
	}

	go rv.run(PooledConnectorConnectionPoolGCEvery, PooledConnectorConnectionPoolStaleAfter)

	return rv
}
