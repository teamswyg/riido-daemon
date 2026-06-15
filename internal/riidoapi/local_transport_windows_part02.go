//go:build windows

package riidoapi

import (
	"errors"
	"io/fs"
	"net"
	"os"
	"syscall"
	"time"
)

type namedPipeConn struct {
	file       *os.File
	path       string
	disconnect bool
}

func newNamedPipeConn(handle syscall.Handle, path string, disconnect bool) *namedPipeConn {
	return &namedPipeConn{
		file:       os.NewFile(uintptr(handle), path),
		path:       path,
		disconnect: disconnect,
	}
}

func (c *namedPipeConn) Read(p []byte) (int, error)  { return c.file.Read(p) }
func (c *namedPipeConn) Write(p []byte) (int, error) { return c.file.Write(p) }

func (c *namedPipeConn) Close() error {
	if c.disconnect {
		_, _, _ = kernel32ProcDisconnectNamedPipe.Call(c.file.Fd())
	}
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
