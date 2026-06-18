//go:build windows

package riidoapi

import (
	"context"
	"net"
	"syscall"
	"time"
)

func dialNamedPipe(ctx context.Context, path string) (net.Conn, error) {
	if err := validateWindowsNamedPipePath(path); err != nil {
		return nil, err
	}
	name, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return nil, err
	}
	for {
		conn, err := tryDialNamedPipe(name, path)
		if err == nil {
			return conn, nil
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

func tryDialNamedPipe(name *uint16, path string) (net.Conn, error) {
	handle, err := syscall.CreateFile(
		name,
		syscall.GENERIC_READ|syscall.GENERIC_WRITE,
		0,
		nil,
		syscall.OPEN_EXISTING,
		syscall.FILE_ATTRIBUTE_NORMAL,
		0,
	)
	if err != nil {
		return nil, err
	}
	return newNamedPipeConn(handle, path, false), nil
}
