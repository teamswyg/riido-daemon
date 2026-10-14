package supervisor

import (
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/toolargs"
)

func toolPayload(tool agentbridge.ToolRef) map[string]any {
	payload := map[string]any{
		"toolID":   tool.ID,
		"toolName": tool.Name,
		"toolKind": tool.Kind,
		"args":     map[string]string{},
	}
	if len(tool.Args) > 0 {
		payload["args"] = toolargs.Clone(tool.Args)
	}
	return payload
}
