package codex

import (
	"strings"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func codexTaskTitle(req agentbridge.StartRequest) string {
	title := []rune(strings.Join(strings.Fields(req.Metadata["task_title"]), " "))
	if len(title) > 256 {
		title = title[:256]
	}
	return string(title)
}
