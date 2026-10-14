package codex

import "github.com/teamswyg/riido-daemon/internal/agentbridge"

func params(raw agentbridge.RawEvent) map[string]any {
	p, _ := raw.Payload["params"].(map[string]any)
	return p
}
