package client

import (
	"crypto/tls"
	"net"
	"time"

	"github.com/juju/errors"
	"github.com/valyala/fasthttp"
)

const (
	connsGCTick        = time.Second
	connsGCAfter       = 3
	connsObsoleteAfter = time.Minute
)

type getConnResponse struct {
	conn net.Conn
	err  error
}

type conns struct {
	dialer       fasthttp.DialFunc
	lastUsed     time.Time
	free         []net.Conn
	toCreate     int
	gcCounter    int
	addr         string
	isTLS        bool
	tlsConfig    *tls.Config
	getChan      chan chan<- getConnResponse
	releasedChan chan net.Conn
	closedChan   chan struct{}
	done         chan struct{}
}

func (c *conns) get(timeout time.Duration) (net.Conn, error) {
	respChan := make(chan getConnResponse)

	select {
	case <-c.done:
		return nil, errors.New("Connections are closing")
	case <-time.After(timeout):
		return nil, errors.New("Timed out")
	case c.getChan <- respChan:
		response := <-respChan
		return response.conn, response.err
	}
}

func (c *conns) put(conn net.Conn) {
	select {
	case <-c.done:
		conn.Close()
		return
	case c.releasedChan <- conn:
		return
	}
}

func (c *conns) use() {
	c.lastUsed = time.Now()
}

func (c *conns) notifyClosed() {
	select {
	case <-c.done:
		return
	case c.closedChan <- struct{}{}:
		return
	}
}

func (c *conns) isObsolete() bool {
	return time.Since(c.lastUsed) > connsObsoleteAfter
}

func (c *conns) stop() {
	select {
	case <-c.done:
		return
	default:
		close(c.done)
	}
}

func (c *conns) run() {
	ticker := time.NewTicker(connsGCTick)
	defer ticker.Stop()

	getChan := c.getChan
	for {
		select {
		case <-c.done:
			for _, v := range c.free {
				v.Close()
			}
			return

		case request := <-getChan:
			if len(c.free) > 0 {
				request <- getConnResponse{conn: c.free[len(c.free)-1]}
				c.free = c.free[:len(c.free)-1]
				continue
			}

			newConn, err := c.dialer(c.addr)
			if err != nil {
				if newConn != nil {
					newConn.Close()
				}
				request <- getConnResponse{err: err}
			}
			if c.isTLS {
				newConn = tls.Client(newConn, c.tlsConfig)
			}
			request <- getConnResponse{conn: newConn}
			c.toCreate--

			if c.toCreate == 0 {
				getChan = nil
			}

		case conn := <-c.releasedChan:
			c.free = append(c.free, conn)
			getChan = c.getChan

		case <-c.closedChan:
			c.toCreate++
			getChan = c.getChan

		case <-ticker.C:
			if len(c.free) == 0 {
				c.gcCounter = 0
				continue
			}

			c.gcCounter++
			if c.gcCounter == connsGCAfter {
				c.gcCounter = 0
				c.free[len(c.free)-1].Close()
				c.free = c.free[:len(c.free)-1]
				c.toCreate++
				getChan = c.getChan
			}
		}
	}
}

func newConns(addr string, isTLS bool, freeSlots int, tlsConfig *tls.Config, dialer fasthttp.DialFunc) (*conns, error) {
	if freeSlots <= 0 {
		return nil, errors.New("FreeSlots should be >= 1")
	}

	return &conns{
		dialer:       dialer,
		addr:         addr,
		isTLS:        isTLS,
		getChan:      make(chan chan<- getConnResponse),
		releasedChan: make(chan net.Conn),
		closedChan:   make(chan struct{}),
		done:         make(chan struct{}),
		free:         make([]net.Conn, 0, freeSlots),
		toCreate:     freeSlots,
	}, nil
}
