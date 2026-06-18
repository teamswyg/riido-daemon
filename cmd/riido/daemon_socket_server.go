package main

import (
	"net"
	"os"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/runtimeactor"
	"github.com/teamswyg/riido-daemon/internal/logging"
	"github.com/teamswyg/riido-daemon/pkg/lifecycle"
)

func serveDaemonSocket(ctx lifecycle.Context, flags startFlags, settings daemonSettings, startedAt time.Time, runtimes []*runtimeactor.Actor, shutdownLevel lifecycle.ShutdownLevel, log logging.Logger) (lifecycle.ShutdownLevel, error) {
	ln, err := net.Listen("unix", flags.socket)
	if err != nil {
		return shutdownLevel, daemonWrapf(ErrDaemonSocket, "serve.listen", err, "listen %s", flags.socket)
	}
	defer cleanupDaemonSocket(ln, flags.socket)

	shutdownCh := make(chan lifecycle.ShutdownLevel, 8)
	done := watchDaemonSocketShutdown(ctx, ln, shutdownCh, shutdownLevel, log)

	log.Printf("daemon listening on %s", flags.socket)
	return acceptDaemonConnections(ln, flags, settings, startedAt, runtimes, shutdownCh, done, shutdownLevel, log)
}

func cleanupDaemonSocket(ln net.Listener, socketPath string) {
	_ = ln.Close()
	_ = os.Remove(socketPath)
}
