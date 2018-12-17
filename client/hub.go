package client

import (
	"crypto/tls"
	"sync"
	"time"

	"github.com/juju/errors"
	"github.com/valyala/fasthttp"
)

const connsHubGCTick = time.Second

type connectionHub struct {
	mutex           sync.RWMutex
	connections     map[string]*conns
	dialer          fasthttp.DialFunc
	tlsConfig       *tls.Config
	connectionLimit int
	isTLS           bool
	done            chan struct{}
}

func (c *connectionHub) get(addr string, timeout time.Duration) (releasableConnection, error) {
	c.mutex.RLock()
	connections, ok := c.connections[addr]
	if ok {
		connections.use()
	}
	c.mutex.RUnlock()

	if !ok {
		c.mutex.Lock()
		connections, ok = c.connections[addr]
		if !ok {
			connections, _ = newConns(addr, c.isTLS, c.connectionLimit, c.tlsConfig, c.dialer)
			c.connections[addr] = connections
			go connections.run()
		}
		connections.use()
		c.mutex.Unlock()
	}

	conn, err := connections.get(timeout)
	if err != nil {
		return releasableConnection{}, err
	}
	return releasableConnection{conn: conn, addr: addr}, nil
}

func (c *connectionHub) put(conn releasableConnection) {
	connections, ok := c.connections[conn.addr]
	if !ok {
		conn.conn.Close()
		return
	}
	connections.put(conn.conn)
}

func (c *connectionHub) notifyClosed(conn releasableConnection) {
	connections, ok := c.connections[conn.addr]
	if !ok {
		conn.conn.Close()
		return
	}
	connections.notifyClosed()
}

func (c *connectionHub) stop() {
	select {
	case <-c.done:
		return
	default:
		close(c.done)
	}
}

func (c *connectionHub) run() {
	ticker := time.NewTicker(connsHubGCTick)
	defer ticker.Stop()

	for {
		select {
		case <-c.done:
			c.gc()
			return
		case <-ticker.C:
			c.gc()
		}
	}
}

func (c *connectionHub) gc() {
	keys := []string{}

	c.mutex.Lock()
	for k, v := range c.connections {
		if v.isObsolete() {
			keys = append(keys, k)
			v.stop()
		}
	}
	for _, k := range keys {
		delete(c.connections, k)
	}
	c.mutex.Unlock()
}

func newConnectionHub(dialer fasthttp.DialFunc, isTLS bool, config *tls.Config, limit int) (*connectionHub, error) {
	if limit <= 0 {
		return nil, errors.New("FreeSlots should be >= 1")
	}

	return &connectionHub{
		dialer:          dialer,
		connectionLimit: limit,
		tlsConfig:       config,
		connections:     make(map[string]*conns),
		mutex:           sync.RWMutex{},
	}, nil
}
