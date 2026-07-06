package codex

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

// BuildProviderInput serializes reducer approval commands into Codex
// app-server JSON-RPC responses.
func BuildProviderInput(cmd agentbridge.Command) ([]byte, error) {
	if cmd.ProviderRequestID == "" {
		return nil, fmt.Errorf("codex: provider request id is required for %s", cmd.Kind)
	}
	decision, err := approvalDecision(cmd)
	if err != nil {
		return nil, err
	}
	body, err := json.Marshal(map[string]any{
		"jsonrpc": "2.0",
		"id":      jsonRPCRequestID(cmd.ProviderRequestID),
		"result":  map[string]any{"decision": decision},
	})
	if err != nil {
		return nil, fmt.Errorf("codex: marshal approval response: %w", err)
	}
	return append(body, '\n'), nil
}

func approvalDecision(cmd agentbridge.Command) (string, error) {
	switch cmd.Kind {
	case agentbridge.CommandApproveTool:
		return "accept", nil
	case agentbridge.CommandRejectTool:
		return "cancel", nil
	default:
		return "", fmt.Errorf("codex: unsupported provider input command %s", cmd.Kind)
	}
}

func jsonRPCRequestID(value string) any {
	if n, err := strconv.ParseInt(value, 10, 64); err == nil {
		return n
	}
	return value
}
