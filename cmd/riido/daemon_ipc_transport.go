package main

import (
	"context"
	"net"
	"runtime"
	"time"

	"github.com/teamswyg/riido-daemon/internal/riidoapi"
)

func daemonLocalTransportForOS(goos string) riidoapi.LocalTransport {
	if goos == "windows" {
		return riidoapi.LocalTransportWindowsNamedPipe
	}
	return riidoapi.LocalTransportUnixSocket
}

func listenDaemonSocket(path string) (net.Listener, func(), error) {
	return riidoapi.ListenLocalEndpoint(daemonLocalTransportForOS(runtime.GOOS), path)
}

func dialDaemonSocket(path string, timeout time.Duration) (net.Conn, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return riidoapi.DialLocalEndpoint(ctx, daemonLocalTransportForOS(runtime.GOOS), path)
}
