package codex

import (
	"encoding/json"
	"fmt"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func BuildProviderInput(cmd agentbridge.Command) ([]byte, error) {
	var id any
	if err := json.Unmarshal([]byte(cmd.ProviderRequestID), &id); err != nil || id == nil {
		return nil, fmt.Errorf("codex: invalid approval request id")
	}
	decision := "decline"
	switch cmd.Kind {
	case agentbridge.CommandApproveTool:
		decision = "accept"
	case agentbridge.CommandRejectTool:
	default:
		return nil, fmt.Errorf("codex: unsupported provider input command %s", cmd.Kind)
	}
	body, err := json.Marshal(map[string]any{"jsonrpc": "2.0", "id": id, "result": map[string]any{"decision": decision}})
	return append(body, '\n'), err
}
