package cpool

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
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
					fmt.Println("+ client disconnected")
					return
				} else if err != nil {
					log.Fatal(err)
				}
				c.Write(line)
			}
		}(c)
	}
}

func dial(t *testing.T) net.Conn {
	c, err := net.Dial("tcp", addr)
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func echo(t *testing.T, c net.Conn, v []byte) {
	c.Write(v)
	b := bufio.NewReader(c)
	line, err := b.ReadBytes('\n')
	if err != nil {
		t.Fatal(err)
	}
	st.Expect(t, line, v)
}

func TestEcho(t *testing.T) {
	c := dial(t)
	defer c.Close()
	echo(t, c, []byte("hello\n"))
}
