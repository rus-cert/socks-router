package connpeeker

import (
	"fmt"
	"io"
	"net"
	"time"
)

type PeekTCPConn struct {
	Conn       *net.TCPConn
	ReadBuffer []byte
}

// fill up to @num bytes into ReadBuffer (appending to existing data)
// @num limits the total number of bytes in ReadBuffer, not the new data
func (c *PeekTCPConn) Peek(num int) error {
	if num <= 0 {
		return fmt.Errorf("Invalid Peek size: %v", num)
	}
	l := len(c.ReadBuffer)
	if cap(c.ReadBuffer) < num {
		tmp := make([]byte, num)
		copy(tmp, c.ReadBuffer)
		c.ReadBuffer = tmp[0:l]
	}
	toread := num - l
	if toread == 0 {
		return nil
	}
	r, err := c.Conn.Read(c.ReadBuffer[l:num])
	if r > 0 {
		c.ReadBuffer = c.ReadBuffer[:l+r]
	}
	return err
}

func (c *PeekTCPConn) Close() error {
	c.ReadBuffer = nil
	return c.Conn.Close()
}

func (c *PeekTCPConn) CloseRead() error {
	c.ReadBuffer = nil
	return c.Conn.CloseRead()
}

func (c *PeekTCPConn) CloseWrite() error {
	return c.Conn.CloseWrite()
}

// File() could be used to read directly without checking the ReadBuffer
//func (c *PeekTCPConn) File() (f *os.File, err error) {
//	return c.Conn.File()
//}

func (c *PeekTCPConn) LocalAddr() net.Addr {
	return c.Conn.LocalAddr()
}

func (c *PeekTCPConn) Read(b []byte) (int, error) {
	if len(c.ReadBuffer) > 0 {
		copy(b, c.ReadBuffer)
		if len(c.ReadBuffer) > len(b) {
			c.ReadBuffer = c.ReadBuffer[len(b):]
			return len(b), nil
		} else {
			l := len(c.ReadBuffer)
			c.ReadBuffer = nil
			return l, nil
		}
	}
	return c.Conn.Read(b)
}

func (c *PeekTCPConn) ReadFrom(r io.Reader) (int64, error) {
	return c.Conn.ReadFrom(r)
}

func (c *PeekTCPConn) RemoteAddr() net.Addr {
	return c.Conn.RemoteAddr()
}

func (c *PeekTCPConn) SetDeadline(t time.Time) error {
	return c.Conn.SetDeadline(t)
}

func (c *PeekTCPConn) SetKeepAlive(keepalive bool) error {
	return c.Conn.SetKeepAlive(keepalive)
}

func (c *PeekTCPConn) SetKeepAlivePeriod(d time.Duration) error {
	return c.Conn.SetKeepAlivePeriod(d)
}

func (c *PeekTCPConn) SetLinger(sec int) error {
	return c.Conn.SetLinger(sec)
}

func (c *PeekTCPConn) SetNoDelay(noDelay bool) error {
	return c.Conn.SetNoDelay(noDelay)
}

func (c *PeekTCPConn) SetReadBuffer(bytes int) error {
	return c.Conn.SetReadBuffer(bytes)
}

func (c *PeekTCPConn) SetReadDeadline(t time.Time) error {
	return c.Conn.SetReadDeadline(t)
}

func (c *PeekTCPConn) SetWriteBuffer(bytes int) error {
	return c.Conn.SetWriteBuffer(bytes)
}

func (c *PeekTCPConn) SetWriteDeadline(t time.Time) error {
	return c.Conn.SetWriteDeadline(t)
}

func (c *PeekTCPConn) Write(b []byte) (int, error) {
	return c.Conn.Write(b)
}
