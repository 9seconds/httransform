package connectors

import (
	"net"
	"sync"
	"time"
)

type pooledConn struct {
	net.Conn

	timestamp time.Time
	once      sync.Once
	pool      *pooledConnectorConnectionPool
}

func (p *pooledConn) Release() {
	p.once.Do(func() {
		go func() {
			if !p.pool.release(p) {
				p.Conn.Close()
			}
		}()
	})
}

func (p *pooledConn) Close() error {
	p.once.Do(func() {
		p.Conn.Close()
	})

	return nil
}
