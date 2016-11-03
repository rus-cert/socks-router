package main

import (
	"fmt"
	"io"
	"net"
	"time"

	"github.com/rus-cert/socks-router/connpeeker"
	"github.com/rus-cert/socks-router/log"
	"github.com/rus-cert/socks-router/routing"
)

// Dialer is imported from routing/
type Dialer routing.Dialer

// ConnHandler describe anything capable of handling connections
type ConnHandler interface {
	ServeConn(conn net.Conn) error
}

// ProtocolHandler detects certain protocols and returns a handler if detected
type ProtocolHandler interface {
	Detect(peek []byte) (ConnHandler, error)
}

// ProtocolMultiplexer uses a list of ProtocolHandlers to handle many
// protocols on the same TCP socket
type ProtocolMultiplexer struct {
	MaxPeek  int
	Handlers []ProtocolHandler
}

// Close closes all ProtocolHandlers which implement io.Closer
func (pm ProtocolMultiplexer) Close() error {
	// return first error or nil
	var err error
	for _, h := range pm.Handlers {
		if c, ok := h.(io.Closer); ok {
			if herr := c.Close(); nil != herr && nil == err {
				err = herr
			}
		}
	}
	return err
}

// ListenTCP accepts and handles TCP connections
func (pm ProtocolMultiplexer) ListenTCP(l *net.TCPListener) error {
	var tempDelay time.Duration
	for {
		if conn, err := l.AcceptTCP(); err != nil {
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				if tempDelay == 0 {
					tempDelay = 5 * time.Millisecond
				} else {
					tempDelay *= 2
				}
				if max := 1 * time.Second; tempDelay > max {
					tempDelay = max
				}
				log.Error.Printf("Accept error: %v; retrying in %v", err, tempDelay)
				time.Sleep(tempDelay)
				continue
			}
			return err
		} else {
			tempDelay = 0
			go func() {
				pm.LogError(conn, pm.ServeTCP(conn))
			}()
		}
	}
}

// LogError logs errors in connection context
func (pm ProtocolMultiplexer) LogError(conn net.Conn, err error) {
	if nil != err {
		log.Error.Printf("Error in %v -> %v: %v", conn.RemoteAddr(), conn.LocalAddr(), err)
	}
}

// ServeTCP handle a single TCP connection
func (pm ProtocolMultiplexer) ServeTCP(conn *net.TCPConn) error {
	pConn := &connpeeker.PeekTCPConn{Conn: conn}
	maxPeek := pm.MaxPeek
	if 0 == maxPeek {
		maxPeek = 1024
	}
	for {
		if err := pConn.Peek(maxPeek); nil != err {
			conn.Close()
			return err
		}
		for _, h := range pm.Handlers {
			if connHandler, err := h.Detect(pConn.ReadBuffer); nil != err {
				conn.Close()
				return err
			} else if nil != connHandler {
				return connHandler.ServeConn(pConn)
			}
			// otherwise continue search for handler
		}
		if maxPeek == len(pConn.ReadBuffer) {
			conn.Close()
			return fmt.Errorf("Couldn't find handler within %v bytes", maxPeek)
		}
	}
}
