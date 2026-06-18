package cursor

import (
	"strconv"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func appendUnsupported(dropped []string, req agentbridge.StartRequest) []string {
	if req.SystemPrompt != "" {
		dropped = append(dropped, "unsupported:system_prompt")
	}
	if req.MaxTurns > 0 {
		dropped = append(dropped, "unsupported:max_turns="+strconv.Itoa(req.MaxTurns))
	}
	return dropped
}
