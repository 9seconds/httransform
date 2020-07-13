package connectors

import "container/heap"

type pooledConnectorConnectionList struct {
	conns []*pooledConn
}

func (p *pooledConnectorConnectionList) Len() int {
	return len(p.conns)
}

func (p *pooledConnectorConnectionList) Less(i, j int) bool {
	return p.conns[i].timestamp.Sub(p.conns[j].timestamp) < 0
}

func (p *pooledConnectorConnectionList) Swap(i, j int) {
	p.conns[i], p.conns[j] = p.conns[j], p.conns[i]
}

func (p *pooledConnectorConnectionList) Push(x interface{}) {
	p.conns = append(p.conns, x.(*pooledConn))
}

func (p *pooledConnectorConnectionList) Pop() interface{} {
	item := p.conns[len(p.conns)-1]
	p.conns = p.conns[:len(p.conns)-1]

	return item
}

func (p *pooledConnectorConnectionList) put(conn *pooledConn) {
	heap.Push(p, conn)
}

func (p *pooledConnectorConnectionList) get() *pooledConn {
	if len(p.conns) == 0 {
		return nil
	}

	return heap.Pop(p).(*pooledConn)
}
