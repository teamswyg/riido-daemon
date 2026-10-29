package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net"

	"github.com/teamswyg/riido-daemon/internal/contextcli"
	"github.com/teamswyg/riido-daemon/internal/execution"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/pkg/lifecycle"
)

// daemonRequest is the JSON envelope read off the socket.
type daemonRequest struct {
	linkRequest      *contextcli.CLIInstallationLinkRequest
	WorkspaceID      string             `json:"workspace_id,omitempty"`
	ExecutionBinding *execution.Binding `json:"execution_binding,omitempty"`

	Method        daemonMethod        `json:"method"`
	ShutdownLevel string              `json:"shutdown_level,omitempty"`
	Force         bool                `json:"force,omitempty"`
	AssignmentID  string              `json:"assignment_id,omitempty"`
	TaskID        string              `json:"task_id,omitempty"`
	RuntimeID     string              `json:"runtime_id,omitempty"`
	Tool          agentbridge.ToolRef `json:"tool"`

	// context_read fields (work-context CLI). RunCapability is never logged.
	ContextSchemaVersion string            `json:"context_schema_version,omitempty"`
	Command              string            `json:"command,omitempty"`
	Args                 map[string]string `json:"args,omitempty"`
	RunCapability        string            `json:"run_capability,omitempty"`
	RequestID            string            `json:"request_id,omitempty"`
}

func readDaemonRequest(conn net.Conn) (daemonRequest, bool, error) {
	var raw json.RawMessage
	var req daemonRequest
	decoder := json.NewDecoder(io.LimitReader(conn, 1024*1024))
	if err := decoder.Decode(&raw); err != nil {
		if errors.Is(err, io.EOF) {
			return daemonRequest{}, false, nil
		}
		return daemonRequest{}, false, err
	}
	if err := json.Unmarshal(raw, &req); err != nil {
		return daemonRequest{}, false, err
	}
	if req.Method == daemonMethodCLIInstallationLinkInfo || req.Method == daemonMethodCLIInstallationLinkConfirm {
		trailing, err := io.ReadAll(decoder.Buffered())
		if err != nil || len(bytes.TrimSpace(trailing)) != 0 {
			return daemonRequest{}, false, &contextcli.CLIInstallationLinkError{Code: "LINK_INVALID_REQUEST"}
		}
		link, err := contextcli.DecodeCLIInstallationLinkRequest(raw)
		if err != nil {
			return daemonRequest{}, false, err
		}
		req.linkRequest = &link
	}
	return req, true, nil
}

func (r daemonRequest) lifecycleShutdownLevel() lifecycle.ShutdownLevel {
	if r.Force {
		return lifecycle.ShutdownForced
	}
	if level, ok := lifecycle.ParseShutdownLevel(r.ShutdownLevel); ok {
		return lifecycle.NormalizeShutdownLevel(level)
	}
	return lifecycle.ShutdownGraceful
}
