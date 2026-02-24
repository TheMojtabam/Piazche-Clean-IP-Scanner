package utils

import (
	"sync"
	"sync/atomic"
)

// PortPool manages a pool of available ports for SOCKS proxies
type PortPool struct {
	startPort int32
	endPort   int32
	current   atomic.Int32
	inUse     sync.Map
}

// NewPortPool creates a new port pool
func NewPortPool(start, end int) *PortPool {
	pp := &PortPool{
		startPort: int32(start),
		endPort:   int32(end),
	}
	pp.current.Store(int32(start))
	return pp
}

// Acquire gets an available port from the pool
func (p *PortPool) Acquire() int {
	for {
		port := p.current.Add(1) - 1
		if port > p.endPort {
			p.current.Store(p.startPort)
			port = p.current.Add(1) - 1
		}

		if _, loaded := p.inUse.LoadOrStore(port, true); !loaded {
			return int(port)
		}
	}
}

// Release returns a port to the pool
func (p *PortPool) Release(port int) {
	p.inUse.Delete(int32(port))
}

// DefaultPortPool is the default port pool instance
var DefaultPortPool = NewPortPool(10000, 60000)

// AcquirePort gets a port from the default pool
func AcquirePort() int {
	return DefaultPortPool.Acquire()
}

// ReleasePort returns a port to the default pool
func ReleasePort(port int) {
	DefaultPortPool.Release(port)
}
