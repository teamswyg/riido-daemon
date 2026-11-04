package codex

import (
	"context"
	"strings"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/provider/history"
)

func (d *protocolDriver) finishSharedHistory(ctx context.Context, turns []any, io agentbridge.ProtocolIO) ([]agentbridge.Event, []agentbridge.Command, error) {
	var matched map[string]any
	var completed []any
	beforeMatched := -1
	unsettled := false
	for _, value := range turns {
		turn, ok := value.(map[string]any)
		if !ok || stringField(turn, "id") == "" {
			return d.failedEvents("codex: invalid shared history turn"), nil, nil
		}
		items, ok := turn["items"].([]any)
		if !ok {
			return d.failedEvents("codex: incomplete shared history items"), nil, nil
		}
		for _, value := range items {
			item, _ := value.(map[string]any)
			if stringField(item, "type") == "userMessage" && stringField(item, "clientId") == d.shared.clientID {
				if matched != nil && stringField(matched, "id") != stringField(turn, "id") {
					return d.failedEvents("codex: duplicate queued turn identity"), nil, nil
				}
				matched = turn
				beforeMatched = len(completed)
			}
		}
		if sharedTurnFinished(turn) {
			completed = append(completed, turn)
		} else {
			unsettled = true
		}
	}
	var events []agentbridge.Event
	if beforeMatched >= 0 {
		// Messages authored after our turn belong to the next Riido pull. Do not
		// project them ahead of the response that this request is still publishing.
		completed = completed[:beforeMatched]
	}
	if unsettled && matched == nil && !d.shared.submitted {
		// The first task anchor may still be in flight in the desktop app.
		return nil, nil, nil
	}
	if history.Enabled(d.req) {
		messages, err := codexHistoryMessages(d.threadID, completed)
		if err != nil {
			return d.failedEvents(err.Error()), nil, nil
		}
		imported, err := history.External(d.req, messages)
		if err != nil {
			return d.failedEvents(err.Error()), nil, nil
		}
		for _, event := range imported {
			id := event.NativeHistory.MessageID
			if !d.shared.imported[id] {
				events = append(events, event)
				d.shared.imported[id] = true
			}
		}
	}
	if matched != nil {
		d.shared.submitted = true
		return append(events, d.sharedTurnEvents(matched)...), nil, nil
	}
	if !d.shared.submitted && !unsettled && !d.shared.stopping {
		d.shared.queueCursors = map[string]bool{}
		return events, nil, d.requestSharedQueue(ctx, "", io)
	}
	return events, nil, nil
}

func sharedTurnFinished(turn map[string]any) bool {
	// A detached reader reconstructs a still-running turn as interrupted. The
	// persisted completion timestamp, not that provisional status, is decisive.
	if turn["completedAt"] == nil {
		return false
	}
	switch stringField(turn, "status") {
	case "completed", "interrupted", "failed":
		return true
	default:
		return false
	}
}

func (d *protocolDriver) sharedTurnEvents(turn map[string]any) []agentbridge.Event {
	var events []agentbridge.Event
	if d.shared.turnID == "" {
		d.shared.turnID = stringField(turn, "id")
		d.turnStarted = true
		events = append(events, agentbridge.Event{Kind: agentbridge.EventLifecycle, Phase: agentbridge.StateRunning})
	}
	var final, all []string
	items, _ := turn["items"].([]any)
	for _, value := range items {
		item, _ := value.(map[string]any)
		if stringField(item, "type") != "agentMessage" {
			continue
		}
		id, text := stringField(item, "id"), stringField(item, "text")
		if id == "" {
			continue
		}
		previous := d.shared.output[id]
		if strings.HasPrefix(text, previous) && len(text) > len(previous) {
			events = append(events, agentbridge.Event{Kind: agentbridge.EventTextDelta, Text: strings.TrimPrefix(text, previous)})
		}
		d.shared.output[id] = text
		all = append(all, text)
		if stringField(item, "phase") == "final_answer" {
			final = append(final, text)
		}
	}
	if sharedTurnFinished(turn) {
		d.shared.done = true
		if len(final) == 0 {
			final = all
		}
		events = append(events, codexTurnCompletedEvent(map[string]any{"turn": turn, "output": strings.Join(final, "\n\n")})...)
	}
	return events
}
