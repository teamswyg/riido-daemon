package claude

import (
	"encoding/json"
	"fmt"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

// BuildProviderInput serializes reducer approval commands into Claude's
// stream-json control_response shape.
func BuildProviderInput(cmd agentbridge.Command) ([]byte, error) {
	requestID := cmd.ProviderRequestID
	if requestID == "" {
		return nil, fmt.Errorf("claude: provider request id is required for %s", cmd.Kind)
	}
	response, err := providerInputResponse(cmd)
	if err != nil {
		return nil, err
	}
	body, err := json.Marshal(map[string]any{
		"type": "control_response",
		"response": map[string]any{
			"subtype":    "success",
			"request_id": requestID,
			"response":   response,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("claude: marshal control_response: %w", err)
	}
	return append(body, '\n'), nil
}

func providerInputResponse(cmd agentbridge.Command) (map[string]any, error) {
	switch cmd.Kind {
	case agentbridge.CommandApproveTool:
		return map[string]any{"behavior": "allow", "updatedInput": map[string]any{}}, nil
	case agentbridge.CommandRejectTool:
		reason := cmd.Reason
		if reason == "" {
			reason = "Permission denied"
		}
		return map[string]any{"behavior": "deny", "message": reason}, nil
	default:
		return nil, fmt.Errorf("claude: unsupported provider input command %s", cmd.Kind)
	}
}
