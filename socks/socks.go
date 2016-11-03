package socks

import (
	"io"
	"net"
)

type SocksError int

const (
	_                            = iota
	ErrInvalidVersion SocksError = iota
	ErrStringTooLong  SocksError = iota
)

func (e SocksError) Error() string {
	switch e {
	case ErrInvalidVersion:
		return "Invalid Socks Version"
	case ErrStringTooLong:
		return "Zero-terminated String in Socks 4A too long"
	}
	return "Unknown Socks error"
}

type Dialer interface {
	// Dial connects to the given address via the proxy.
	Dial(network, addr string) (c net.Conn, err error)
}

type Server struct {
	Dialer Dialer
}

func (s Server) ServeConn(conn net.Conn) error {
	defer conn.Close()
	var hdr [1]byte
	if _, err := io.ReadFull(conn, hdr[:]); nil != err {
		return err
	}
	switch hdr[0] {
	case 0x04:
		return s.serverConnSocks4(conn)
	case 0x05:
		return s.serverConnSocks5(conn)
	default:
		return ErrInvalidVersion
	}
}
