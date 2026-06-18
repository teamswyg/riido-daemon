//go:build windows

package riidoapi

import (
	"context"
	"errors"
	"fmt"
	"net"
)

func listenLocalEndpoint(transport LocalTransport, path string) (net.Listener, func(), error) {
	if err := validateLocalTransportPath(transport, path); err != nil {
		return nil, nil, err
	}
	switch transport {
	case LocalTransportUnixSocket:
		return nil, nil, errors.New("unix socket transport is not supported on Windows")
	case LocalTransportWindowsNamedPipe:
		if err := validateWindowsNamedPipePath(path); err != nil {
			return nil, nil, err
		}
		listener := newNamedPipeListener(path)
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
