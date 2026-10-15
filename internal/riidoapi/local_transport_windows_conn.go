//go:build windows

package riidoapi

import (
	"errors"
	"io"
	"io/fs"
	"net"
	"os"
	"syscall"
	"time"
)

type namedPipeConn struct {
	file *os.File
	path string
}

func newNamedPipeConn(handle syscall.Handle, path string) *namedPipeConn {
	return &namedPipeConn{
		file: os.NewFile(uintptr(handle), path),
		path: path,
	}
}

func (c *namedPipeConn) Read(p []byte) (int, error) {
	n, err := c.file.Read(p)
	if errors.Is(err, errorBrokenPipe) || errors.Is(err, errorNoData) || errors.Is(err, errorPipeNotConn) {
		return n, io.EOF
	}
	return n, err
}
func (c *namedPipeConn) Write(p []byte) (int, error) { return c.file.Write(p) }

func (c *namedPipeConn) Close() error {
	// Every connection owns a fresh handle, so it is never reconnected.
	// DisconnectNamedPipe would discard response bytes a client has not read.
	// Closing the handle leaves the pipe in CLOSING until buffered data drains.
	return c.file.Close()
}

func (c *namedPipeConn) LocalAddr() net.Addr  { return namedPipeAddr(c.path) }
func (c *namedPipeConn) RemoteAddr() net.Addr { return namedPipeAddr(c.path) }

func (c *namedPipeConn) SetDeadline(t time.Time) error {
	return ignoreUnsupportedDeadline(c.file.SetDeadline(t))
}

func (c *namedPipeConn) SetReadDeadline(t time.Time) error {
	return ignoreUnsupportedDeadline(c.file.SetReadDeadline(t))
}

func (c *namedPipeConn) SetWriteDeadline(t time.Time) error {
	return ignoreUnsupportedDeadline(c.file.SetWriteDeadline(t))
}

func ignoreUnsupportedDeadline(err error) error {
	if errors.Is(err, fs.ErrInvalid) {
		return nil
	}
	return err
}
