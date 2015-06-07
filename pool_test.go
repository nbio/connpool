// +build !race
// This package cannot be tested with the Go race detector (-race)
// because sync.Pool is disabled (no-op) when race detection is enabled.

package connpool

import (
	"bufio"
	"io"
	"log"
	"net"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/nbio/st"
)

const addr = "127.0.0.1:3540"

func init() {
	go server()
}

func server() {
	srv, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}
	for {
		c, err := srv.Accept()
		if err != nil {
			log.Fatal(err)
		}
		go func(c net.Conn) {
			defer c.Close()
			b := bufio.NewReader(c)
			for {
				line, err := b.ReadBytes('\n')
				if err == io.EOF {
					return
				} else if err != nil {
					log.Fatal(err)
				}
				c.Write(line)
			}
		}(c)
	}
}

func echo(t *testing.T, c net.Conn, v []byte) {
	c.Write(v)
	b := bufio.NewReader(c)
	line, err := b.ReadBytes('\n')
	st.Expect(t, err, nil)
	st.Expect(t, line, v)
}

func TestEcho(t *testing.T) {
	c, err := net.Dial("tcp", addr)
	st.Assert(t, err, nil)
	defer c.Close()
	echo(t, c, []byte("hello\n"))
}

func TestPool(t *testing.T) {
	var dialCount int32
	p := Pool{New: func() (net.Conn, error) {
		atomic.AddInt32(&dialCount, 1)
		return net.Dial("tcp", addr)
	}}

	// Get and Put one
	for i := 0; i < 10; i++ {
		c, err := p.Get()
		st.Expect(t, err, nil)
		echo(t, c, []byte("alpha\n"))
		p.Put(c)
	}
	st.Expect(t, dialCount, int32(1))
	runtime.GC()

	// Get several
	dialCount = 0
	var cs []net.Conn
	for i := 0; i < 5; i++ {
		c, err := p.Get()
		st.Expect(t, err, nil)
		cs = append(cs, c)
	}
	for _, c := range cs {
		p.Put(c)
	}
	for i := 0; i < 5; i++ {
		_, err := p.Get()
		st.Expect(t, err, nil)
	}
	st.Expect(t, dialCount, int32(5))
	runtime.GC()

	// Test concurrency
	var wg sync.WaitGroup
	dialCount = 0
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := p.Get()
			st.Expect(t, err, nil)
		}()
	}
	wg.Wait()
	st.Expect(t, dialCount, int32(10))
}
