package codex

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

const (
	codexMCPApprovalMethod = "mcpServer/elicitation/request"
	codexMCPApprovalKind   = "mcp_confirmation"
)

func (d *protocolDriver) handleMCPApproval(ctx context.Context, raw agentbridge.RawEvent, io agentbridge.ProtocolIO) ([]agentbridge.Event, []agentbridge.Command, error) {
	id, err := mcpApprovalRequestID(raw.Payload["id"])
	if err != nil {
		return nil, nil, err
	}
	if _, seen := d.mcpApprovals[id]; seen {
		return nil, nil, nil
	}
	p := params(raw)
	message := strings.TrimSpace(stringField(p, "message"))
	if d.threadID == "" || d.owned.id == "" || d.owned.done ||
		stringField(p, "threadId") != d.threadID || stringField(p, "turnId") != d.owned.id ||
		!isMCPToolConfirmation(p) || message == "" {
		frame, err := mcpApprovalResponse(id, "cancel")
		if err == nil {
			err = io.WriteStdin(ctx, frame)
		}
		return []agentbridge.Event{{Kind: agentbridge.EventWarning, Text: "Codex MCP confirmation cancelled: unsupported form or unmatched active turn"}}, nil, err
	}
	sum := sha256.Sum256([]byte(d.threadID + "\x00" + d.owned.id + "\x00" + id))
	toolID := fmt.Sprintf("mcp-%x", sum[:16])
	if d.mcpApprovals == nil {
		d.mcpApprovals = make(map[string]string)
	}
	d.mcpApprovals[id] = toolID
	return []agentbridge.Event{{
		Kind: agentbridge.EventToolApprovalNeeded,
		Tool: agentbridge.ToolRef{ID: toolID, Name: message, Kind: codexMCPApprovalKind, ProviderRequestID: id, RequireHuman: true},
	}}, nil, nil
}

func isMCPToolConfirmation(p map[string]any) bool {
	meta := mapField(p, "_meta")
	schema := mapField(p, "requestedSchema")
	properties, ok := schema["properties"].(map[string]any)
	required, hasRequired := schema["required"]
	if entries, ok := required.([]any); ok && len(entries) == 0 {
		hasRequired = false
	}
	return stringField(p, "serverName") != "" && stringField(p, "mode") == "form" &&
		stringField(meta, "codex_approval_kind") == "mcp_tool_call" &&
		meta["codex_requires_user_input"] != true &&
		stringField(schema, "type") == "object" && ok && len(properties) == 0 && !hasRequired
}

func mcpApprovalRequestID(value any) (string, error) {
	switch id := value.(type) {
	case string:
		if id == "" {
			return "", fmt.Errorf("codex: empty MCP confirmation request ID")
		}
	case int, int64, float64, json.Number:
	default:
		return "", fmt.Errorf("codex: invalid MCP confirmation request ID")
	}
	encoded, err := json.Marshal(value)
	return string(encoded), err
}

func (d *protocolDriver) BuildProviderInput(cmd agentbridge.Command) ([]byte, error) {
	if cmd.ToolKind != codexMCPApprovalKind {
		return BuildProviderInput(cmd)
	}
	toolID := d.mcpApprovals[cmd.ProviderRequestID]
	if toolID == "" || toolID != cmd.ToolID {
		return nil, fmt.Errorf("codex: MCP confirmation is not pending for this tool")
	}
	action := "decline"
	switch cmd.Kind {
	case agentbridge.CommandApproveTool:
		action = "accept"
	case agentbridge.CommandRejectTool:
	default:
		return nil, fmt.Errorf("codex: unsupported MCP confirmation command %s", cmd.Kind)
	}
	frame, err := mcpApprovalResponse(cmd.ProviderRequestID, action)
	if err == nil {
		d.mcpApprovals[cmd.ProviderRequestID] = ""
	}
	return frame, err
}

func (d *protocolDriver) cancelMCPApprovals(ctx context.Context, io agentbridge.ProtocolIO) error {
	for id, toolID := range d.mcpApprovals {
		if toolID == "" {
			continue
		}
		frame, err := mcpApprovalResponse(id, "cancel")
		if err != nil {
			return err
		}
		if err := io.WriteStdin(ctx, frame); err != nil {
			return err
		}
		d.mcpApprovals[id] = ""
	}
	return nil
}

func mcpApprovalResponse(id, action string) ([]byte, error) {
	body, err := json.Marshal(map[string]any{
		"jsonrpc": "2.0",
		"id":      json.RawMessage(id),
		"result":  map[string]string{"action": action},
	})
	return append(body, '\n'), err
}
