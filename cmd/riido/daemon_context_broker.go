package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane/saasplane"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/runtimeactor"
	"github.com/teamswyg/riido-daemon/internal/contextbroker"
	"github.com/teamswyg/riido-daemon/internal/contextcli"
	"github.com/teamswyg/riido-daemon/internal/logging"
)

const contextRunActiveTimeout = time.Second

func parseDaemonContextCLI(raw string) (bool, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return false, nil
	}
	switch strings.ToLower(value) {
	case "on":
		return true, nil
	case "off":
		return false, nil
	}
	enabled, err := strconv.ParseBool(value)
	if err != nil {
		return false, daemonErrorf(ErrDaemonConfig, "settings.context-cli", "%s must be on/off or a boolean: %q", envContextCLI, raw)
	}
	return enabled, nil
}

// newDaemonContextBroker puts the directory of this binary first on each
// run's PATH so the agent's `riido` is the same build as the daemon.
func newDaemonContextBroker(socket string) *contextbroker.Broker {
	binDir := ""
	if exe, err := os.Executable(); err == nil {
		binDir = filepath.Dir(exe)
	}
	return contextbroker.New(socket, binDir)
}

// daemonContextLaunchEnv returns the runtimeactor hook, or nil when the
// work-context CLI is off so launches stay exactly as before.
func daemonContextLaunchEnv(broker *contextbroker.Broker) func(bridge.TaskRequest, string, map[string]string) {
	if broker == nil {
		return nil
	}
	return func(req bridge.TaskRequest, runtimeID string, env map[string]string) {
		_ = broker.LaunchEnv(contextBindingForRun(req, runtimeID), env)
	}
}

func contextBindingForRun(req bridge.TaskRequest, runtimeID string) contextbroker.Binding {
	taskID := strings.TrimSpace(req.Metadata[controlplane.MetadataTaskID])
	if taskID == "" {
		taskID = req.ID
	}
	return contextbroker.Binding{
		RunKey:       req.ID,
		TaskID:       taskID,
		WorkspaceID:  strings.TrimSpace(req.Metadata["workspace_id"]),
		RuntimeID:    runtimeID,
		AgentID:      strings.TrimSpace(req.Metadata[saasplane.MetadataAgentID]),
		AssignmentID: strings.TrimSpace(req.Metadata[saasplane.MetadataAssignmentID]),
	}
}

// writeContextRead answers the CLI broker request. With the feature off the
// daemon behaves like older builds and reports an unknown method.
func writeContextRead(conn net.Conn, req daemonRequest, settings daemonSettings, runtimes []*runtimeactor.Actor, log logging.Logger) {
	if settings.contextBroker == nil {
		_ = writeUnknownDaemonMethod(conn, req.Method)
		return
	}
	reply := settings.contextBroker.Handle(contextcli.ContextReadRequest{
		Method:        string(req.Method),
		SchemaVersion: req.ContextSchemaVersion,
		Command:       req.Command,
		Args:          req.Args,
		RunCapability: req.RunCapability,
		TaskID:        req.TaskID,
		WorkspaceID:   req.WorkspaceID,
		RequestID:     req.RequestID,
	}, runtimeRunActive(runtimes))
	log.Printf("%s", contextReadAuditLine(req, reply))
	_ = writeDaemonJSON(conn, reply)
}

// contextReadAuditLine records command, target, result, and request id.
// It never includes the run capability or response bodies.
func contextReadAuditLine(req daemonRequest, reply contextcli.Envelope) string {
	result := "ok"
	if reply.Error != nil {
		result = string(reply.Error.Code)
	}
	target := req.Args["work"]
	if target == "" {
		target = "-"
	}
	return fmt.Sprintf("context_read command=%q target=%q result=%s request_id=%q sources=%d",
		req.Command, target, result, reply.RequestID, len(reply.Sources))
}

// runtimeRunActive reports whether any runtime still has the run in flight.
func runtimeRunActive(runtimes []*runtimeactor.Actor) contextbroker.RunActive {
	return func(runKey string) bool {
		ctx, cancel := context.WithTimeout(context.Background(), contextRunActiveTimeout)
		defer cancel()
		for _, rt := range runtimes {
			beat, err := rt.HeartbeatPayload(ctx)
			if err == nil && slices.Contains(beat.RunningTaskIDs, runKey) {
				return true
			}
		}
		return false
	}
}
