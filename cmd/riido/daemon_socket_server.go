package main

import (
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/runtimeactor"
	"github.com/teamswyg/riido-daemon/internal/logging"
	"github.com/teamswyg/riido-daemon/pkg/lifecycle"
)

func serveDaemonSocket(ctx lifecycle.Context, flags startFlags, settings daemonSettings, startedAt time.Time, runtimes []*runtimeactor.Actor, resolver agentbridge.ToolApprovalResolver, authorizer agentbridge.ToolApprovalAuthorizer, shutdownLevel lifecycle.ShutdownLevel, log logging.Logger) (lifecycle.ShutdownLevel, error) {
	ln, cleanup, err := listenDaemonSocket(flags.socket)
	if err != nil {
		return shutdownLevel, daemonWrapf(ErrDaemonSocket, "serve.listen", err, "listen %s", flags.socket)
	}
	defer cleanup()

	shutdownCh := make(chan lifecycle.ShutdownLevel, 8)
	done := watchDaemonSocketShutdown(ctx, ln, shutdownCh, shutdownLevel, log)

	log.Printf("daemon listening on %s", flags.socket)
	return acceptDaemonConnections(ln, flags, settings, startedAt, runtimes, resolver, authorizer, shutdownCh, done, shutdownLevel, log)
}
