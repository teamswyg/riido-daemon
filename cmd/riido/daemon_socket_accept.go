package main

import (
	"errors"
	"net"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/runtimeactor"
	"github.com/teamswyg/riido-daemon/internal/logging"
	"github.com/teamswyg/riido-daemon/pkg/lifecycle"
)

func acceptDaemonConnections(ln net.Listener, flags startFlags, settings daemonSettings, startedAt time.Time, runtimes []*runtimeactor.Actor, shutdownCh chan<- lifecycle.ShutdownLevel, done <-chan lifecycle.ShutdownLevel, shutdownLevel lifecycle.ShutdownLevel, log logging.Logger) (lifecycle.ShutdownLevel, error) {
	for {
		conn, err := ln.Accept()
		if err != nil {
			level, ok := daemonAcceptShutdownLevel(done)
			if ok {
				return level, nil
			}
			if errors.Is(err, net.ErrClosed) {
				return shutdownLevel, nil
			}
			log.Printf("accept error: %v", err)
			continue
		}
		go handleDaemonConn(conn, flags, settings, startedAt, runtimes, shutdownCh, log)
	}
}

func daemonAcceptShutdownLevel(done <-chan lifecycle.ShutdownLevel) (lifecycle.ShutdownLevel, bool) {
	select {
	case level := <-done:
		return level, true
	default:
		return lifecycle.ShutdownGraceful, false
	}
}
