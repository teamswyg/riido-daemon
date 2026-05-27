//go:build !windows

package riidoapi

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
)

func listenLocalEndpoint(transport LocalTransport, path string) (net.Listener, func(), error) {
	if err := validateLocalTransportPath(transport, path); err != nil {
		return nil, nil, err
	}
	switch transport {
	case LocalTransportUnixSocket:
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return nil, nil, fmt.Errorf("create socket directory: %w", err)
		}
		_ = os.Remove(path)
		listener, err := net.Listen("unix", path)
		if err != nil {
			return nil, nil, err
		}
		_ = os.Chmod(path, 0o600)
		cleanup := func() {
			_ = listener.Close()
			_ = os.Remove(path)
		}
		return listener, cleanup, nil
	case LocalTransportWindowsNamedPipe:
		return nil, nil, errors.New("windows named pipe transport requires Windows")
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
		return (&net.Dialer{}).DialContext(ctx, "unix", path)
	case LocalTransportWindowsNamedPipe:
		return nil, errors.New("windows named pipe transport requires Windows")
	default:
		return nil, fmt.Errorf("unknown local transport %q", transport)
	}
}
