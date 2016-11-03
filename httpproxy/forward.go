package httpproxy

import (
	"io"
)

type closeWriter interface {
	CloseWrite() error
}

func Forward(conn1W io.Writer, conn1R io.Reader, conn2W io.Writer, conn2R io.Reader) error {
	// Start forwarding
	errCh := make(chan error, 2)
	go forwardSingle(conn1W, conn2R, errCh)
	go forwardSingle(conn2W, conn1R, errCh)

	// Wait
	for i := 0; i < 2; i++ {
		e := <-errCh
		if e != nil {
			return e
		}
	}
	return nil
}

// forwardSingle is used to suffle data from src to destination, and
// sends errors down a dedicated channel
func forwardSingle(dst io.Writer, src io.Reader, errCh chan error) {
	_, err := io.Copy(dst, src)
	if cw, ok := dst.(closeWriter); ok {
		cw.CloseWrite()
	}
	errCh <- err
}
