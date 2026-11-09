package codex

import (
	"strings"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func (d *protocolDriver) observeAssistantMessage(method codexMethod, p map[string]any) {
	if method != codexMethodItemCompleted {
		return
	}
	item := mapField(p, "item")
	if stringField(item, "type") != "agentMessage" {
		return
	}
	if strings.TrimSpace(stringField(item, "text")) != "" {
		d.hasAssistantText = true
	}
	switch stringField(item, "phase") {
	case "final_answer":
		d.finalOutput = stringField(item, "text")
		d.hasFinalAnswer = true
	case "":
		// Older runtimes do not classify message phases. The last completed
		// message is the compatibility answer, never a concatenated transcript.
		if !d.hasFinalAnswer {
			d.finalOutput = stringField(item, "text")
		}
	}
}

func (d *protocolDriver) turnResultEvents(method codexMethod, p map[string]any) []agentbridge.Event {
	if stringField(mapField(p, "turn"), "status") == "" && d.lastRuntimeError != "" && !d.sawAssistantOutput && stringField(p, "output") == "" && d.finalOutput == "" {
		return d.failedEvents(d.lastRuntimeError)
	}
	events := codexTurnCompletedEvent(p)
	result := &events[0].Result
	if result.Status != agentbridge.ResultCompleted {
		return events
	}
	if result.Output == "" {
		result.Output = d.finalOutput
	}
	if method == codexMethodTurnCompleteSlash && strings.TrimSpace(result.Output) == "" && d.hasAssistantText {
		return d.failedEvents("codex: completed turn has no final assistant message")
	}
	d.turnStarted = false
	return events
}
