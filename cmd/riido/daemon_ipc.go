package main

import (
	"net"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/runtimeactor"
	"github.com/teamswyg/riido-daemon/internal/logging"
	"github.com/teamswyg/riido-daemon/pkg/lifecycle"
)

func handleDaemonConn(conn net.Conn, flags startFlags, settings daemonSettings, startedAt time.Time, runtimes []*runtimeactor.Actor, resolver agentbridge.ToolApprovalResolver, authorizer agentbridge.ToolApprovalAuthorizer, shutdownCh chan<- lifecycle.ShutdownLevel, log logging.Logger) {
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(5 * time.Second))

	req, ok, err := readDaemonRequest(conn)
	if err != nil {
		log.Printf("decode request: %v", err)
		_ = writeDaemonDecodeError(conn, err)
	}
	if !ok {
		return
	}
	log.Printf("%s request received", req.Method)
	dispatchDaemonRequest(conn, req, flags, settings, startedAt, runtimes, resolver, authorizer, shutdownCh, log)
}

func dispatchDaemonRequest(conn net.Conn, req daemonRequest, flags startFlags, settings daemonSettings, startedAt time.Time, runtimes []*runtimeactor.Actor, resolver agentbridge.ToolApprovalResolver, authorizer agentbridge.ToolApprovalAuthorizer, shutdownCh chan<- lifecycle.ShutdownLevel, log logging.Logger) {
	switch req.Method {
	case daemonMethodStatus, daemonMethodDefault:
		writeStatus(conn, flags, settings, startedAt, runtimes)
	case daemonMethodHealth:
		writeHealth(conn)
	case daemonMethodReady:
		writeReady(conn, runtimes)
	case daemonMethodMetrics:
		writeMetrics(conn, runtimes)
	case daemonMethodShutdown:
		level := req.lifecycleShutdownLevel()
		writeShutdownAck(conn, level)
		// Non-blocking signal — repeated shutdown requests are harmless.
		select {
		case shutdownCh <- level:
		default:
		}
		log.Printf("shutdown request received level=%s", level)
	case daemonMethodToolApproval:
		_ = conn.SetDeadline(time.Time{})
		writeToolApprovalResolution(conn, req, resolver, authorizer, log)
	default:
		if err := writeUnknownDaemonMethod(conn, req.Method); err != nil {
			log.Printf("write unknown-method response: %v", err)
		}
	}
}
