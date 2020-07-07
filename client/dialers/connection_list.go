package dialers

import "container/heap"

type connectionList struct {
	conns []*pooledConn
}

func (c *connectionList) Len() int {
	return len(c.conns)
}

func (c *connectionList) Less(i, j int) bool {
	return c.conns[i].timestamp.Sub(c.conns[j].timestamp) < 0
}

func (c *connectionList) Swap(i, j int) {
	c.conns[i], c.conns[j] = c.conns[j], c.conns[i]
}

func (c *connectionList) Push(x interface{}) {
	c.conns = append(c.conns, x.(*pooledConn))
}

func (c *connectionList) Pop() interface{} {
	item := c.conns[len(c.conns)-1]
	c.conns = c.conns[:len(c.conns)-1]

	return item
}

func (c *connectionList) put(conn *pooledConn) {
	heap.Push(c, conn)
}

func (c *connectionList) get() *pooledConn {
	if len(c.conns) == 0 {
		return nil
	}

	return heap.Pop(c).(*pooledConn)
}
