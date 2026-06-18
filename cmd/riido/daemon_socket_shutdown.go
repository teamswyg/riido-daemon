package main

import (
	"net"

	"github.com/teamswyg/riido-daemon/internal/logging"
	"github.com/teamswyg/riido-daemon/pkg/lifecycle"
)

func watchDaemonSocketShutdown(ctx lifecycle.Context, ln net.Listener, shutdownCh <-chan lifecycle.ShutdownLevel, shutdownLevel lifecycle.ShutdownLevel, log logging.Logger) <-chan lifecycle.ShutdownLevel {
	signalCtx, stop := lifecycle.Notify(ctx.WithShutdownLevel(shutdownLevel), daemonInterruptSignals()...)
	done := make(chan lifecycle.ShutdownLevel, 1)
	go func() {
		defer stop()
		level := daemonSocketShutdownLevel(signalCtx, shutdownCh, log)
		done <- level
		_ = ln.Close()
	}()
	return done
}

func daemonSocketShutdownLevel(signalCtx lifecycle.Context, shutdownCh <-chan lifecycle.ShutdownLevel, log logging.Logger) lifecycle.ShutdownLevel {
	var level lifecycle.ShutdownLevel
	select {
	case <-signalCtx.Done():
		level = lifecycle.NormalizeShutdownLevel(signalCtx.ShutdownLevel())
		log.Printf("daemon shutdown requested level=%s source=signal", level)
	case level = <-shutdownCh:
		level = lifecycle.NormalizeShutdownLevel(level)
		log.Printf("daemon shutdown requested level=%s source=socket", level)
	}
	return level
}
