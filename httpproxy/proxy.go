package httpproxy

import (
	"errors"
	"net"
	"net/http"
	"net/http/httputil"
	"time"

	"github.com/rus-cert/socks-router/log"
)

type httpProxy struct {
	dial         func(network, addr string) (c net.Conn, err error)
	reverseProxy *httputil.ReverseProxy
}

func HTTPProxy(dial func(network, addr string) (c net.Conn, err error)) http.Handler {
	return &httpProxy{
		dial: dial,
		reverseProxy: &httputil.ReverseProxy{
			Transport: &http.Transport{
				Dial:                  dial,
				TLSHandshakeTimeout:   10 * time.Second,
				ExpectContinueTimeout: 1 * time.Second,
			},
			Director: func(req *http.Request) {
				req.RequestURI = ""
				if _, ok := req.Header["User-Agent"]; !ok {
					// explicitly disable User-Agent so it's not set to default value
					req.Header.Set("User-Agent", "")
				}
			},
			ErrorLog:      log.Error,
			FlushInterval: time.Second,
		},
	}
}

var errNoRedirect = errors.New("Redirect disabled")

func noRedirect(req *http.Request, via []*http.Request) error {
	return errNoRedirect
}

func (p *httpProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if "CONNECT" == r.Method {
		hj, ok := w.(http.Hijacker)
		if !ok {
			log.Error.Printf("failed CONNECT to %q: webserver doesn't support hijacking", r.RequestURI)
			http.Error(w, "webserver doesn't support hijacking", http.StatusInternalServerError)
			return
		}

		log.Debug.Printf("http CONNECT: %q", r.RequestURI)
		backend, err := p.dial("tcp", r.RequestURI)
		if nil != err {
			http.Error(w, err.Error(), 503)
			return
		}
		defer backend.Close()

		conn, bufrw, err := hj.Hijack()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// Don't forget to close the connection:
		defer conn.Close()
		bufrw.WriteString(r.Proto + " 200 OK\r\n\r\n")
		bufrw.Flush()

		log.Debug.Printf("forwarding CONNECT to %q", r.RequestURI)
		if err := Forward(conn, bufrw, backend, backend); nil != err {
			log.Error.Printf("failed CONNECT to %q: %v", r.RequestURI, err)
		} else {
			log.Debug.Printf("done CONNECT: %q", r.RequestURI)
		}
	} else {
		log.Debug.Printf("forwarding HTTP request for %q", r.RequestURI)
		p.reverseProxy.ServeHTTP(w, r)
	}
}
