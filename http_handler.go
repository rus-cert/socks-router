package main

import (
	"bytes"
	"net"
	"net/http"
	"strings"

	"github.com/rus-cert/socks-router/connpeeker"
	"github.com/rus-cert/socks-router/httpproxy"
	"github.com/rus-cert/socks-router/log"
)

type httpHandler struct {
	httpListener *connpeeker.FakeListener
	dialer       Dialer
	server       *http.Server
}

type httpConnectHandler struct {
	dialer  Dialer
	address string
}

func (h httpConnectHandler) ServeConn(conn net.Conn) error {
	defer conn.Close()
	pc := conn.(*connpeeker.PeekTCPConn)
	// drop *all* read bytes so far; continue with underlying connection
	pc.ReadBuffer = nil

	log.Debug.Printf("(bad http) CONNECT: %q", h.address)
	if backend, err := h.dialer.Dial("tcp", h.address); nil != err {
		pc.Conn.Write([]byte("HTTP/1.0 503 OK\r\n\r\n"))
		return err
	} else {
		defer backend.Close()
		if _, err := pc.Conn.Write([]byte("HTTP/1.0 200 OK\r\n\r\n")); nil != err {
			return err
		}
		return httpproxy.Forward(pc.Conn, pc.Conn, backend, backend)
	}
}

func (h httpHandler) Close() error {
	return h.httpListener.Close()
}

func (h httpHandler) Detect(peek []byte) (ConnHandler, error) {
	// find request line
	if eol := bytes.IndexAny(peek, "\r\n"); -1 != eol {
		fields := strings.Split(string(peek[:eol]), " ")
		if 3 == len(fields) && strings.HasPrefix(fields[2], "HTTP/") {
			// "good" HTTP request
			return h.httpListener, nil
		} else if 2 == len(fields) && "CONNECT" == fields[0] {
			// "bad" CONNECT (not really HTTP) request
			return httpConnectHandler{
				dialer:  h.dialer,
				address: fields[1],
			}, nil
		}
	}
	return nil, nil
}

// CreateHTTPHandler returns a ProtocolHandler to detect and handle HTTP
// and CONNECT requests
func CreateHTTPHandler(dialer Dialer) (ProtocolHandler, error) {
	// http.Server doesn't have a "ServeConn" method; it only supports
	// the Listener interface, so pass connections through a "fake
	// listener" (a simple queue)
	httpListener := connpeeker.NewFakeListener()
	server := &http.Server{
		Handler:        httpproxy.HTTPProxy(dialer.Dial),
		MaxHeaderBytes: 1 << 20,
	}

	// needs to poll listener in a separate thread; should exit on
	// httpHandler.Close()
	go server.Serve(httpListener)

	return httpHandler{
		httpListener: httpListener,
		dialer:       dialer,
		server:       server,
	}, nil
}
