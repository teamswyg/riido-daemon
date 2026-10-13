package cursor

import (
	"strings"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

// The CLI has no separate system surface. Preserve supplied instructions while
// keeping the actual task first in the conversation.
func cursorMessage(req agentbridge.StartRequest) string {
	system := strings.TrimSpace(req.SystemPrompt)
	if system == "" {
		return req.Prompt
	}
	return req.Prompt + "\n\nRiido runtime instructions:\n" + system
}
