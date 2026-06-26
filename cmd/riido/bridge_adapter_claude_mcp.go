package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane/saasplane"
	"github.com/teamswyg/riido-daemon/internal/provider/claude"
)

const claudePermissionPromptToolName = "mcp__riido__approval"

func (a bridgeClaudeAdapter) startOptions(req agentbridge.StartRequest) (claude.StartOptions, error) {
	opts := claude.StartOptions{PermissionMode: claude.PermissionModeApproval}
	if !a.approvalMCPEnabled(req) {
		return opts, nil
	}
	path, err := a.writeApprovalMCPConfig(req)
	if err != nil {
		return opts, err
	}
	opts.MCPConfigPath = path
	opts.PermissionPromptToolName = claudePermissionPromptToolName
	return opts, nil
}

func (a bridgeClaudeAdapter) approvalMCPEnabled(req agentbridge.StartRequest) bool {
	return strings.TrimSpace(a.approvalSocket) != "" &&
		strings.TrimSpace(req.Metadata[saasplane.MetadataAssignmentID]) != ""
}

func (a bridgeClaudeAdapter) writeApprovalMCPConfig(req agentbridge.StartRequest) (string, error) {
	executable, err := os.Executable()
	if err != nil {
		return "", err
	}
	args := []string{
		string(mainCommandClaudePermissionMCP),
	}
	args = appendMCPArg(args, "--socket", a.approvalSocket)
	args = appendMCPArg(args, "--assignment-id", req.Metadata[saasplane.MetadataAssignmentID])
	args = appendMCPArg(args, "--task-id", req.Metadata[controlplane.MetadataTaskID])
	args = appendMCPArg(args, "--runtime-id", req.Metadata[controlplane.MetadataRuntimeID])
	if err := requireMCPArgs(args); err != nil {
		return "", err
	}
	body, err := json.Marshal(claudePermissionMCPConfig{
		MCPServers: map[string]claudePermissionMCPServer{
			"riido": {Command: executable, Args: args},
		},
	})
	if err != nil {
		return "", err
	}
	file, err := os.CreateTemp("", "riido-claude-permission-mcp-*.json")
	if err != nil {
		return "", err
	}
	defer file.Close()
	if _, err := file.Write(body); err != nil {
		return "", err
	}
	return file.Name(), nil
}

func appendMCPArg(args []string, name, value string) []string {
	value = strings.TrimSpace(value)
	if value == "" {
		return args
	}
	return append(args, name, value)
}

func requireMCPArgs(args []string) error {
	required := map[string]bool{
		"--socket":        false,
		"--assignment-id": false,
		"--task-id":       false,
		"--runtime-id":    false,
	}
	for i := 0; i+1 < len(args); i++ {
		if _, ok := required[args[i]]; ok {
			required[args[i]] = strings.TrimSpace(args[i+1]) != ""
		}
	}
	for name, ok := range required {
		if !ok {
			return fmt.Errorf("claude permission MCP config missing %s", name)
		}
	}
	return nil
}

type claudePermissionMCPConfig struct {
	MCPServers map[string]claudePermissionMCPServer `json:"mcpServers"`
}

type claudePermissionMCPServer struct {
	Command string   `json:"command"`
	Args    []string `json:"args"`
}
