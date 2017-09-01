package connpeeker

import (
	"errors"
	"net"
)

var errListenerClosed = errors.New("Listener was closed")

type fakeListeningAddress struct{}

var fake_listening_address = fakeListeningAddress{}

func (_ fakeListeningAddress) Network() string {
	return "tcp"
}

func (_ fakeListeningAddress) String() string {
	return "fake-address"
}

type FakeListener struct {
	queueOut chan<- net.Conn
	queueIn  <-chan net.Conn
}

func NewFakeListener() *FakeListener {
	q := make(chan net.Conn, 16)
	return &FakeListener{
		queueOut: q,
		queueIn:  q,
	}
}

func (l *FakeListener) ServeConn(conn net.Conn) error {
	if nil == l.queueOut {
		return errListenerClosed
	} else {
		l.queueOut <- conn
		return nil
	}
}

func (l *FakeListener) Accept() (net.Conn, error) {
	if conn, ok := <-l.queueIn; ok {
		return conn, nil
	} else {
		return nil, errListenerClosed
	}
}

// Close closes the listener.
// Any blocked Accept operations will be unblocked and return errors.
func (l *FakeListener) Close() error {
	if nil == l.queueOut {
		return errListenerClosed
	}
	q := l.queueOut
	l.queueOut = nil
	close(q)
	return nil
}

// Addr returns the listener's network address.
func (l *FakeListener) Addr() net.Addr {
	return fake_listening_address
}
