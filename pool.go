package connpool

import (
	"fmt"
	"net"
	"sync"
)

// Pool maintains a pool of net.Conn created with the New func.
// Internally it wraps a sync.Pool, and shares its runtime characteristics.
// Any Conn stored in the Pool may be removed automatically at any time without notification.
// If the Pool holds the only reference when this happens, the Conn might be deallocated.
//
// Note: Pool assumes the underlying net.Conn implementation will automatically close
// (typically with a Finalizer). It makes no attempts to close the underlying Conn.
//
// Pool is safe for use by multiple goroutines.
type Pool struct {
	New  func() (net.Conn, error)
	pool sync.Pool
}

// Get returns a net.Conn or an error if unable to create a new one
// with the New func. It will panic if p has a nil New func.
func (p *Pool) Get() (net.Conn, error) {
	i := p.pool.Get()
	if c, ok := i.(net.Conn); ok && c != nil {
		return c, nil
	}
	c, err := p.New()
	if err != nil {
		return nil, err
	}
	return c, nil
}

// Put releases the Conn to the pool.
func (p *Pool) Put(c net.Conn) {
	p.pool.Put(c)
}

func close(c net.Conn) {
	fmt.Printf("closing %+v\n", c)
	c.Close()
}
