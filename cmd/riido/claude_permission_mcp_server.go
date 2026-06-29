package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
)

func serveClaudePermissionMCP(in io.Reader, out io.Writer, opts claudePermissionMCPOptions) error {
	encoder := json.NewEncoder(out)
	scanner := bufio.NewScanner(in)
	scanner.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
	for scanner.Scan() {
		var req mcpRequest
		if err := json.Unmarshal(scanner.Bytes(), &req); err != nil {
			continue
		}
		if req.ID == nil {
			continue
		}
		if err := encoder.Encode(handleClaudePermissionMCPRequest(req, opts)); err != nil {
			return err
		}
	}
	return scanner.Err()
}

func handleClaudePermissionMCPRequest(req mcpRequest, opts claudePermissionMCPOptions) mcpResponse {
	switch req.Method {
	case "initialize":
		return mcpOK(req.ID, map[string]any{
			"protocolVersion": "2025-11-25",
			"capabilities":    map[string]any{"tools": map[string]any{}},
			"serverInfo":      map[string]string{"name": "riido", "version": versionLabel()},
		})
	case "tools/list":
		return mcpOK(req.ID, map[string]any{"tools": []map[string]any{claudePermissionToolSpec()}})
	case "tools/call":
		return handleClaudePermissionToolCall(req, opts)
	default:
		return mcpOK(req.ID, map[string]any{})
	}
}

func handleClaudePermissionToolCall(req mcpRequest, opts claudePermissionMCPOptions) mcpResponse {
	var params mcpToolCallParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return mcpErr(req.ID, -32602, err.Error())
	}
	if params.Name != "approval" {
		return mcpErr(req.ID, -32602, fmt.Sprintf("unknown tool %q", params.Name))
	}
	decision, err := resolveClaudePermissionViaDaemon(opts, params.Arguments)
	if err != nil {
		decision = claudePermissionDecision{Behavior: "deny", Message: err.Error()}
	}
	return mcpOK(req.ID, map[string]any{
		"content": []map[string]string{{"type": "text", "text": decision.JSON()}},
		"isError": false,
	})
}

func claudePermissionToolSpec() map[string]any {
	return map[string]any{
		"name":        "approval",
		"description": "Routes Claude Code tool permission prompts to Riido.",
		"inputSchema": map[string]any{
			"type":                 "object",
			"additionalProperties": true,
		},
	}
}

func mcpOK(id, result any) mcpResponse {
	return mcpResponse{JSONRPC: "2.0", ID: id, Result: result}
}

func mcpErr(id any, code int, message string) mcpResponse {
	return mcpResponse{JSONRPC: "2.0", ID: id, Error: &mcpErrorObj{Code: code, Message: message}}
}
