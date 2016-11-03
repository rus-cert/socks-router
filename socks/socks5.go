package socks

import (
	"bytes"
	"encoding/binary"
	"io"
	"net"
	"strconv"
	"strings"

	"github.com/rus-cert/socks-router/httpproxy"
)

type socks5ResultCode byte

const (
	errCode5Succeeded               byte             = 0x00
	errCode5Failure                 socks5ResultCode = 0x01
	errCode5Forbidden               socks5ResultCode = 0x02
	errCode5NetworkUnreachable      socks5ResultCode = 0x03
	errCode5HostUnreachable         socks5ResultCode = 0x04
	errCode5ConnectionRefused       socks5ResultCode = 0x05
	errCode5TTLExpired              socks5ResultCode = 0x06
	errCode5CommandNotSupported     socks5ResultCode = 0x07
	errCode5AddressTypeNotSupported socks5ResultCode = 0x08
)

func sendSocks5ReplyAddr(conn net.Conn, code byte, ip []byte, port uint16) error {
	l := 6
	iplen := 4
	var atyp byte = 1
	if 16 == len(ip) {
		iplen = 16
		atyp = 4
	}
	resp := make([]byte, l+iplen)
	resp[0] = 0x05
	resp[1] = code
	resp[2] = 0
	resp[3] = atyp
	copy(resp[4:4+iplen], ip)
	binary.BigEndian.PutUint16(resp[4+iplen:], port)

	_, err := conn.Write(resp)
	return err
}

func sendSocks5ReplyConn(conn net.Conn, target net.Conn) error {
	if nil != target {
		if local, ok := target.LocalAddr().(*net.TCPAddr); ok {
			return sendSocks5ReplyAddr(conn, errCode5Succeeded, local.IP, uint16(local.Port))
		}
	}
	return sendSocks5ReplyAddr(conn, errCode5Succeeded, nil, 0)
}

func sendSocks5Error(conn net.Conn, code socks5ResultCode) error {
	return sendSocks5ReplyAddr(conn, byte(code), nil, 0)
}

func socks5MapDialError(err error) socks5ResultCode {
	msg := err.Error()
	if strings.Contains(msg, "refused") {
		return errCode5ConnectionRefused
	} else if strings.Contains(msg, "network is unreachable") {
		return errCode5NetworkUnreachable
	}
	return errCode5HostUnreachable
}

func (s Server) serverConnSocks5(conn net.Conn) error {
	{
		var methods []byte
		var nmethods [1]byte
		if _, err := io.ReadFull(conn, nmethods[:]); nil != err {
			return err
		} else if nmethods[0] > 0 {
			methods = make([]byte, nmethods[0])
			if _, err := io.ReadFull(conn, methods); nil != err {
				return err
			}
		}

		if -1 == bytes.IndexByte(methods, 0) {
			// client doesn't support "no authentication"
			conn.Write([]byte{0x05, 0xff}) // no acceptable method
			return nil
		} else {
			conn.Write([]byte{0x05, 0x00}) // select "no authentication"
		}
	}

	var hdr [4]byte
	if _, err := io.ReadFull(conn, hdr[:]); nil != err {
		sendSocks5Error(conn, errCode5Failure)
		return err
	}

	if 0x05 != hdr[0] {
		sendSocks5Error(conn, errCode5Failure)
		return ErrInvalidVersion
	}

	switch hdr[1] {
	case 1: // CONNECT
		break
	default:
		// unsupported command
		return sendSocks5Error(conn, errCode5CommandNotSupported)
	}

	var addr string
	var port uint16
	switch hdr[3] {
	case 0x01: // IPv4
		var ipAndPort [6]byte
		if _, err := io.ReadFull(conn, ipAndPort[:]); nil != err {
			sendSocks5Error(conn, errCode5Failure)
			return err
		}
		addr = net.IP(ipAndPort[:4]).String()
		port = binary.BigEndian.Uint16(ipAndPort[4:])
	case 0x03: // FQDN
		var fqdnLen [1]byte
		if _, err := io.ReadFull(conn, fqdnLen[:]); nil != err {
			sendSocks5Error(conn, errCode5Failure)
			return err
		}
		fqdnAndPort := make([]byte, fqdnLen[0]+2)
		if _, err := io.ReadFull(conn, fqdnAndPort[:]); nil != err {
			sendSocks5Error(conn, errCode5Failure)
			return err
		}
		addr = string(fqdnAndPort[:fqdnLen[0]])
		port = binary.BigEndian.Uint16(fqdnAndPort[fqdnLen[0]:])
	case 0x04: // IPv6
		var ipAndPort [18]byte
		if _, err := io.ReadFull(conn, ipAndPort[:]); nil != err {
			sendSocks5Error(conn, errCode5Failure)
			return err
		}
		addr = net.IP(ipAndPort[:16]).String()
		port = binary.BigEndian.Uint16(ipAndPort[16:])
	default:
		// unsupported address type
		return sendSocks5Error(conn, errCode5AddressTypeNotSupported)
	}

	if backend, err := s.Dialer.Dial("tcp", net.JoinHostPort(addr, strconv.Itoa(int(port)))); nil != err {
		return sendSocks5Error(conn, socks5MapDialError(err))
	} else {
		defer backend.Close()
		if err := sendSocks5ReplyConn(conn, backend); nil != err {
			return err
		}
		return httpproxy.Forward(conn, conn, backend, backend)
	}
}
