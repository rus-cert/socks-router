package socks

import (
	"bytes"
	"encoding/binary"
	"io"
	"net"
	"strconv"

	"github.com/rus-cert/socks-router/httpproxy"
)

type socks4ResultCode byte

const (
	errCode4Granted        socks4ResultCode = 90
	errCode4Rejected       socks4ResultCode = 91
	errCode4MissingIdentd  socks4ResultCode = 92
	errCode4MismatchIdentd socks4ResultCode = 93
)

// make sure cap(buf) >= minLen
func resizeBuf(buf *[]byte, minLen int) {
	l := len(*buf)
	if cap(*buf) < minLen {
		tmp := make([]byte, minLen)
		copy(tmp, *buf)
		*buf = tmp[0:l]
	}
}

func readZeroTerminatedString(buf *[]byte, conn net.Conn, maxLen int) ([]byte, error) {
	resizeBuf(buf, maxLen)
	for {
		if zeroPos := bytes.IndexByte(*buf, 0); -1 != zeroPos {
			result := (*buf)[:zeroPos]
			*buf = (*buf)[zeroPos+1:]
			return result, nil
		}

		l := len(*buf)
		if l >= maxLen {
			return nil, ErrStringTooLong
		}

		n, err := conn.Read((*buf)[l:maxLen])
		if n > 0 {
			*buf = (*buf)[l : l+n]
		}
		if nil != err {
			return nil, err
		}
	}
}

func sendSocks4Reply(conn net.Conn, code socks4ResultCode) error {
	resp := [8]byte{0, byte(code), 0, 0, 0, 0, 0, 0}
	_, err := conn.Write(resp[:])
	return err
}

func (s Server) serverConnSocks4(conn net.Conn) error {
	var hdr [7]byte
	if _, err := io.ReadFull(conn, hdr[:]); nil != err {
		return err
	}

	switch hdr[0] {
	case 1: // CONNECT
		break
	default:
		// unsupported command
		sendSocks4Reply(conn, errCode4Rejected)
		return nil
	}

	buf := make([]byte, 0, 512)
	if _ /* username */, err := readZeroTerminatedString(&buf, conn, 128); nil != err {
		sendSocks4Reply(conn, errCode4Rejected)
		return err
	}
	port := binary.BigEndian.Uint16(hdr[1:3])
	var addr string
	if 0 == hdr[3] && 0 == hdr[4] && 0 == hdr[5] && 0 != hdr[6] {
		// SOCKS 4A
		if dest, err := readZeroTerminatedString(&buf, conn, 256); nil != err {
			sendSocks4Reply(conn, errCode4Rejected)
			return err
		} else {
			addr = string(dest)
		}
	} else {
		// SOCKS 4
		ip := net.IP(hdr[3:7])
		addr = ip.String()
	}

	if backend, err := s.Dialer.Dial("tcp", net.JoinHostPort(addr, strconv.Itoa(int(port)))); nil != err {
		sendSocks4Reply(conn, errCode4Rejected)
		return err
	} else {
		defer backend.Close()
		if err := sendSocks4Reply(conn, errCode4Granted); nil != err {
			return err
		}
		return httpproxy.Forward(conn, conn, backend, backend)
	}
}
