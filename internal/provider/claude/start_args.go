package claude

import (
	"encoding/json"
	"fmt"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func buildStartArgs(req agentbridge.StartRequest, opts StartOptions) ([]string, []string, []string) {
	args := protocolArgs(opts.PermissionMode)
	args = appendContextCLIReadArgs(args, opts.ContextCLIReads)
	if len(req.ReferenceDirectories) > 0 {
		settings, _ := json.Marshal(map[string]any{"permissions": map[string]any{"additionalDirectories": req.ReferenceDirectories}})
		args = append(args, "--settings", string(settings))
	}
	tempFiles := []string{}
	args, tempFiles = appendMCPArgs(args, tempFiles, opts.MCPConfigPath)
	if opts.PermissionMode != PermissionModeBypassDangerous || opts.PermissionPromptToolName != "" {
		args = appendPermissionPromptTool(args, opts.PermissionPromptToolName)
	}
	args = appendRequestArgs(args, req)
	blocked := startBlockedArgs(opts.ContextCLIReads)
	if len(req.ReferenceDirectories) > 0 {
		blocked = append(blocked, "--settings", "--add-dir")
	}
	kept, dropped := agentbridge.FilterBlockedArgs(req.CustomArgs, blocked)
	return append(args, kept...), dropped, tempFiles
}

func appendPermissionPromptTool(args []string, name string) []string {
	if name == "" {
		name = "stdio"
	}
	return appendStringArg(args, "--permission-prompt-tool", name)
}

func protocolArgs(mode PermissionMode) []string {
	return []string{
		"-p",
		"--output-format", "stream-json",
		"--input-format", "stream-json",
		"--verbose",
		"--include-partial-messages",
		"--permission-mode", string(mode),
	}
}

func appendMCPArgs(args, tempFiles []string, path string) ([]string, []string) {
	if path == "" {
		return args, tempFiles
	}
	args = append(args, "--strict-mcp-config", "--mcp-config", path)
	return args, append(tempFiles, path)
}

func appendRequestArgs(args []string, req agentbridge.StartRequest) []string {
	args = appendStringArg(args, "--model", req.Model)
	args = appendStringArg(args, "--append-system-prompt", req.SystemPrompt)
	if req.MaxTurns > 0 {
		args = append(args, "--max-turns", fmt.Sprintf("%d", req.MaxTurns))
	}
	return appendStringArg(args, "--resume", req.ResumeSessionID)
}

func appendStringArg(args []string, name, value string) []string {
	if value == "" {
		return args
	}
	return append(args, name, value)
}
