//go:build windows

package riidoapi

import (
	"context"
	"net"
	"sync"
	"syscall"
	"time"
)

type namedPipeListener struct {
	path   string
	done   chan struct{}
	closed sync.Once
}

func newNamedPipeListener(path string) *namedPipeListener {
	return &namedPipeListener{path: path, done: make(chan struct{})}
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
	return l.acceptHandle(handle)
}

func (l *namedPipeListener) acceptHandle(handle syscall.Handle) (net.Conn, error) {
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
