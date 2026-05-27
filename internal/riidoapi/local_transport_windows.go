//go:build windows

package riidoapi

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"net"
	"os"
	"strings"
	"sync"
	"syscall"
	"time"
	"unsafe"
)

const (
	namedPipeAccessDuplex = 0x00000003
	namedPipeTypeByte     = 0x00000000
	namedPipeReadModeByte = 0x00000000
	namedPipeWait         = 0x00000000
	namedPipeInstances    = 255
	namedPipeBufferSize   = 64 * 1024

	errorPipeBusy      syscall.Errno = 231
	errorPipeConnected syscall.Errno = 535
)

var (
	kernel32ProcConnectNamedPipe    = syscall.NewLazyDLL("kernel32.dll").NewProc("ConnectNamedPipe")
	kernel32ProcCreateNamedPipe     = syscall.NewLazyDLL("kernel32.dll").NewProc("CreateNamedPipeW")
	kernel32ProcDisconnectNamedPipe = syscall.NewLazyDLL("kernel32.dll").NewProc("DisconnectNamedPipe")
)

type namedPipeListener struct {
	path   string
	done   chan struct{}
	closed sync.Once
}

type namedPipeAddr string

func listenLocalEndpoint(transport LocalTransport, path string) (net.Listener, func(), error) {
	if err := validateLocalTransportPath(transport, path); err != nil {
		return nil, nil, err
	}
	switch transport {
	case LocalTransportUnixSocket:
		return nil, nil, errors.New("unix socket transport is not supported on Windows")
	case LocalTransportWindowsNamedPipe:
		if !strings.HasPrefix(strings.ToLower(path), `\\.\pipe\`) {
			return nil, nil, errors.New(`windows named pipe path must start with \\.\pipe\`)
		}
		listener := &namedPipeListener{
			path: path,
			done: make(chan struct{}),
		}
		return listener, func() { _ = listener.Close() }, nil
	default:
		return nil, nil, fmt.Errorf("unknown local transport %q", transport)
	}
}

func dialLocalEndpoint(ctx context.Context, transport LocalTransport, path string) (net.Conn, error) {
	if err := validateLocalTransportPath(transport, path); err != nil {
		return nil, err
	}
	switch transport {
	case LocalTransportUnixSocket:
		return nil, errors.New("unix socket transport is not supported on Windows")
	case LocalTransportWindowsNamedPipe:
		return dialNamedPipe(ctx, path)
	default:
		return nil, fmt.Errorf("unknown local transport %q", transport)
	}
}

func (l *namedPipeListener) Accept() (net.Conn, error) {
	select {
	case <-l.done:
		return nil, net.ErrClosed
	default:
	}

	handle, err := createNamedPipe(l.path)
	if err != nil {
		return nil, err
	}
	connected, err := connectNamedPipe(handle)
	if err != nil {
		_ = syscall.CloseHandle(handle)
		return nil, err
	}
	if !connected {
		_ = syscall.CloseHandle(handle)
		return nil, net.ErrClosed
	}
	return newNamedPipeConn(handle, l.path, true), nil
}

func (l *namedPipeListener) Close() error {
	l.closed.Do(func() {
		close(l.done)
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()
		conn, err := dialNamedPipe(ctx, l.path)
		if err == nil {
			_ = conn.Close()
		}
	})
	return nil
}

func (l *namedPipeListener) Addr() net.Addr {
	return namedPipeAddr(l.path)
}

func (a namedPipeAddr) Network() string { return string(LocalTransportWindowsNamedPipe) }
func (a namedPipeAddr) String() string  { return string(a) }

func createNamedPipe(path string) (syscall.Handle, error) {
	name, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return syscall.InvalidHandle, err
	}
	handle, _, callErr := kernel32ProcCreateNamedPipe.Call(
		uintptr(unsafe.Pointer(name)),
		uintptr(namedPipeAccessDuplex),
		uintptr(namedPipeTypeByte|namedPipeReadModeByte|namedPipeWait),
		uintptr(namedPipeInstances),
		uintptr(namedPipeBufferSize),
		uintptr(namedPipeBufferSize),
		0,
		0,
	)
	if syscall.Handle(handle) == syscall.InvalidHandle {
		return syscall.InvalidHandle, callErr
	}
	return syscall.Handle(handle), nil
}

func connectNamedPipe(handle syscall.Handle) (bool, error) {
	ok, _, callErr := kernel32ProcConnectNamedPipe.Call(uintptr(handle), 0)
	if ok != 0 {
		return true, nil
	}
	errno, isErrno := callErr.(syscall.Errno)
	if isErrno && errno == errorPipeConnected {
		return true, nil
	}
	return false, callErr
}

func dialNamedPipe(ctx context.Context, path string) (net.Conn, error) {
	if !strings.HasPrefix(strings.ToLower(path), `\\.\pipe\`) {
		return nil, errors.New(`windows named pipe path must start with \\.\pipe\`)
	}
	name, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return nil, err
	}
	for {
		handle, err := syscall.CreateFile(
			name,
			syscall.GENERIC_READ|syscall.GENERIC_WRITE,
			0,
			nil,
			syscall.OPEN_EXISTING,
			syscall.FILE_ATTRIBUTE_NORMAL,
			0,
		)
		if err == nil {
			return newNamedPipeConn(handle, path, false), nil
		}
		errno, ok := err.(syscall.Errno)
		if !ok || errno != errorPipeBusy {
			return nil, err
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(20 * time.Millisecond):
		}
	}
}

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
