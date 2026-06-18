package cursor

import "github.com/teamswyg/riido-daemon/internal/agentbridge"

func translateStepFinish(p map[string]any) []agentbridge.Event {
	if usage, ok := p["usage"].(map[string]any); ok {
		return []agentbridge.Event{{Kind: agentbridge.EventUsageDelta, Usage: parseUsage(usage)}}
	}
	return []agentbridge.Event{{Kind: agentbridge.EventLog, Text: "cursor step_finish without usage"}}
}
