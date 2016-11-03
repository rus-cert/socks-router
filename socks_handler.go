package main

import (
	"github.com/rus-cert/socks-router/socks"
)

type socksHandler struct {
	handler ConnHandler
}

func (h socksHandler) Detect(peek []byte) (ConnHandler, error) {
	// first byte "0x04" or "0x05" -> SOCKS4 (or 4A) or SOCKS5
	if len(peek) > 0 && (4 == peek[0] || 5 == peek[0]) {
		return h.handler, nil
	}
	return nil, nil
}

// CreateSocksHandler returns a ProtocolHandler to detect and handle
// SOCKS[4,4a,5] requests
func CreateSocksHandler(dialer Dialer) (ProtocolHandler, error) {
	server := socks.Server{
		Dialer: dialer,
	}

	return socksHandler{server}, nil
}
