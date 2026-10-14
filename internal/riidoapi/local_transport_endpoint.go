package riidoapi

import (
	"context"
	"net"
)

// ListenLocalEndpoint opens the host-specific local IPC transport.
func ListenLocalEndpoint(transport LocalTransport, path string) (net.Listener, func(), error) {
	return listenLocalEndpoint(transport, path)
}

// DialLocalEndpoint connects to the host-specific local IPC transport.
func DialLocalEndpoint(ctx context.Context, transport LocalTransport, path string) (net.Conn, error) {
	return dialLocalEndpoint(ctx, transport, path)
}
