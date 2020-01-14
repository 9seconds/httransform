package dialers

import (
	"context"
	"sync"
	"time"
)

const (
	connectionPoolGCEvery    = time.Second
	connectionPoolStaleAfter = 30 * time.Second
)

type connectionPool struct {
	list                 connectionList
	dialer               Dialer
	ctx                  context.Context
	cancel               context.CancelFunc
	lastRequestAt        time.Time
	channelRequests      chan pooledDialerDialRequest
	channelReturn        chan *pooledConn
	channelLastRequestAt chan chan<- time.Time
}

func (c *connectionPool) get(req pooledDialerDialRequest) {
	select {
	case <-c.ctx.Done():
		req.response <- pooledDialerResponse{err: ErrPooledDialerCanceled}
	case c.channelRequests <- req:
	}
}

func (c *connectionPool) release(conn *pooledConn) bool {
	select {
	case <-c.ctx.Done():
		return false
	case c.channelReturn <- conn:
		return true
	}
}

func (c *connectionPool) getLastRequestAt() time.Time {
	response := make(chan time.Time)

	select {
	case <-c.ctx.Done():
		return time.Time{}
	case c.channelLastRequestAt <- response:
		return <-response
	}
}

func (c *connectionPool) run(gcEvery, staleAfter time.Duration) {
	ticker := time.NewTicker(gcEvery)
	defer ticker.Stop()

	for {
		select {
		case <-c.ctx.Done():
			for _, v := range c.list.conns {
				v.Close()
			}

			c.list.conns = []*pooledConn{}

			return
		case conn := <-c.channelReturn:
			conn.once = sync.Once{}
			conn.timestamp = time.Now()
			c.list.put(conn)
		case req := <-c.channelRequests:
			c.lastRequestAt = time.Now()

			if conn := c.list.get(); conn != nil {
				req.response <- pooledDialerResponse{conn: conn}
				close(req.response)

				continue
			}

			conn, err := c.dialer.Dial(req.ctx, req.addr)
			req.response <- pooledDialerResponse{
				conn: &pooledConn{Conn: conn, pool: c},
				err:  err,
			}
			close(req.response)
		case <-ticker.C:
			if conn := c.list.get(); conn != nil {
				if time.Since(conn.timestamp) > staleAfter {
					conn.Close()
				} else {
					c.list.put(conn)
				}
			}
		case response := <-c.channelLastRequestAt:
			response <- c.lastRequestAt
			close(response)
		}
	}
}

func (c *connectionPool) shutdown() {
	c.cancel()
}

func newConnectionPool(ctx context.Context, dialer Dialer) *connectionPool {
	ctx, cancel := context.WithCancel(ctx)
	rv := &connectionPool{
		list:                 connectionList{},
		dialer:               dialer,
		ctx:                  ctx,
		cancel:               cancel,
		channelRequests:      make(chan pooledDialerDialRequest),
		channelReturn:        make(chan *pooledConn),
		channelLastRequestAt: make(chan chan<- time.Time),
	}

	go rv.run(connectionPoolGCEvery, connectionPoolStaleAfter)

	return rv
}
