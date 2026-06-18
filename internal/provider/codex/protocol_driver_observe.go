package codex

import (
	"strings"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func (d *protocolDriver) observeEvents(events []agentbridge.Event) {
	for _, ev := range events {
		switch ev.Kind {
		case agentbridge.EventTextDelta:
			if strings.TrimSpace(ev.Text) != "" {
				d.sawAssistantOutput = true
			}
		case agentbridge.EventResult:
			if strings.TrimSpace(ev.Result.Output) != "" {
				d.sawAssistantOutput = true
			}
		default:
		}
	}
}

func (d *protocolDriver) recordRuntimeError(message string) {
	if message == "" {
		message = "codex runtime error"
	}
	d.lastRuntimeError = message
}
