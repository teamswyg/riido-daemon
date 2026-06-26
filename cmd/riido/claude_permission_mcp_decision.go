package main

import (
	"encoding/json"
	"net"
	"strings"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/toolargs"
)

type claudePermissionDecision struct {
	Behavior     string         `json:"behavior"`
	UpdatedInput map[string]any `json:"updatedInput,omitempty"`
	Message      string         `json:"message,omitempty"`
}

type daemonToolApprovalResponse struct {
	Approved bool   `json:"approved"`
	Reason   string `json:"reason,omitempty"`
	Error    string `json:"error,omitempty"`
	Detail   string `json:"detail,omitempty"`
}

func resolveClaudePermissionViaDaemon(opts claudePermissionMCPOptions, raw json.RawMessage) (claudePermissionDecision, error) {
	args, err := parseClaudePermissionArguments(raw)
	if err != nil {
		return claudePermissionDecision{}, err
	}
	resp, err := requestDaemonToolApproval(opts, args)
	if err != nil {
		return claudePermissionDecision{}, err
	}
	if resp.Approved {
		return claudePermissionDecision{Behavior: "allow", UpdatedInput: map[string]any{}}, nil
	}
	return claudePermissionDecision{Behavior: "deny", Message: firstApprovalMessage(resp)}, nil
}

func parseClaudePermissionArguments(raw json.RawMessage) (claudePermissionArguments, error) {
	var args claudePermissionArguments
	if err := json.Unmarshal(raw, &args); err != nil {
		return args, err
	}
	return args, nil
}

func requestDaemonToolApproval(opts claudePermissionMCPOptions, args claudePermissionArguments) (daemonToolApprovalResponse, error) {
	conn, err := net.DialTimeout("unix", opts.socket, 5*time.Second)
	if err != nil {
		return daemonToolApprovalResponse{}, err
	}
	defer conn.Close()
	req := daemonRequest{
		Method:       daemonMethodToolApproval,
		AssignmentID: opts.assignmentID,
		TaskID:       opts.taskID,
		RuntimeID:    opts.runtimeID,
		Tool: agentbridge.ToolRef{
			ID:   strings.TrimSpace(args.ToolUseID),
			Name: strings.TrimSpace(args.ToolName),
			Kind: strings.TrimSpace(args.ToolName),
			Args: toolargs.FromValue(args.Input),
		},
	}
	if err := json.NewEncoder(conn).Encode(req); err != nil {
		return daemonToolApprovalResponse{}, err
	}
	var resp daemonToolApprovalResponse
	if err := json.NewDecoder(conn).Decode(&resp); err != nil {
		return daemonToolApprovalResponse{}, err
	}
	return resp, nil
}

func firstApprovalMessage(resp daemonToolApprovalResponse) string {
	for _, value := range []string{resp.Reason, resp.Detail, resp.Error, "Permission denied"} {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return "Permission denied"
}

func (d claudePermissionDecision) JSON() string {
	body, err := json.Marshal(d)
	if err != nil {
		return `{"behavior":"deny","message":"Permission denied"}`
	}
	return string(body)
}
