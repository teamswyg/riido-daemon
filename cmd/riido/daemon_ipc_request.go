package main

import (
	"encoding/json"
	"errors"
	"io"
	"net"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/pkg/lifecycle"
)

// daemonRequest is the JSON envelope read off the socket.
type daemonRequest struct {
	Method        daemonMethod        `json:"method"`
	ShutdownLevel string              `json:"shutdown_level,omitempty"`
	Force         bool                `json:"force,omitempty"`
	AssignmentID  string              `json:"assignment_id,omitempty"`
	TaskID        string              `json:"task_id,omitempty"`
	RuntimeID     string              `json:"runtime_id,omitempty"`
	Tool          agentbridge.ToolRef `json:"tool"`
}

func readDaemonRequest(conn net.Conn) (daemonRequest, bool, error) {
	var req daemonRequest
	if err := json.NewDecoder(conn).Decode(&req); err != nil {
		if errors.Is(err, io.EOF) {
			return daemonRequest{}, false, nil
		}
		return daemonRequest{}, false, err
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
