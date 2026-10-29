package main

import (
	"errors"
	"net"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/runtimeactor"
	"github.com/teamswyg/riido-daemon/internal/contextcli"
	"github.com/teamswyg/riido-daemon/internal/logging"
	"github.com/teamswyg/riido-daemon/pkg/lifecycle"
)

func handleDaemonConn(conn net.Conn, flags startFlags, settings daemonSettings, startedAt time.Time, runtimes []*runtimeactor.Actor, resolver agentbridge.ToolApprovalResolver, authorizer agentbridge.ToolApprovalAuthorizer, shutdownCh chan<- lifecycle.ShutdownLevel, log logging.Logger) {
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(5 * time.Second))

	req, ok, err := readDaemonRequest(conn)
	if err != nil {
		log.Printf("decode request: %v", err)
		var linkErr *contextcli.CLIInstallationLinkError
		if errors.As(err, &linkErr) {
			_ = writeDaemonJSON(conn, contextcli.CLIInstallationLinkReply{SchemaVersion: contextcli.CLIInstallationLinkIPCSchema, Error: linkErr})
		} else {
			_ = writeDaemonDecodeError(conn, err)
		}
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
	case daemonMethodExecutionBindingUpdate, daemonMethodExecutionBindingRemove:
		writeExecutionBindingUpdate(conn, req, settings)
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
	case daemonMethodContextRead:
		// A control-plane fetch may take up to its own timeout; the default
		// 5s connection deadline would drop a slower successful answer.
		_ = conn.SetDeadline(time.Now().Add(contextReadConnDeadline))
		writeContextRead(conn, req, settings, runtimes, log)
	case daemonMethodCLIInstallationLinkInfo, daemonMethodCLIInstallationLinkConfirm:
		_ = conn.SetDeadline(time.Now().Add(12 * time.Second))
		ctx, cancel := daemonConnClosedContext(conn)
		defer cancel()
		writeCLIInstallationLink(ctx, conn, req.linkRequest, settings)
	case daemonMethodToolApproval:
		_ = conn.SetDeadline(time.Time{})
		ctx, cancel := daemonConnClosedContext(conn)
		defer cancel()
		writeToolApprovalResolution(ctx, conn, req, resolver, authorizer, log)
	default:
		if err := writeUnknownDaemonMethod(conn, req.Method); err != nil {
			log.Printf("write unknown-method response: %v", err)
		}
	}
}
