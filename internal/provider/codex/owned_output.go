package codex

import (
	"strings"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func (d *protocolDriver) rememberMessage(item map[string]any) {
	id := stringField(item, "id")
	if id == "" || stringField(item, "type") != "agentMessage" {
		return
	}
	if d.owned.messages == nil {
		d.owned.messages = map[string]map[string]any{}
	}
	if _, exists := d.owned.messages[id]; !exists {
		d.owned.order = append(d.owned.order, id)
	}
	d.owned.messages[id] = item
}

func (d *protocolDriver) ownedMessageEvents(method codexMethod, p map[string]any) ([]agentbridge.Event, bool) {
	switch method {
	case codexMethodItemStarted, codexMethodItemUpdated, codexMethodItemCompleted:
		item := mapField(p, "item")
		if stringField(item, "type") != "agentMessage" {
			return nil, false
		}
		d.rememberMessage(item)
		if method == codexMethodItemCompleted {
			return d.ownedMessageDelta(stringField(item, "id"), stringField(item, "text"), false), true
		}
		return nil, true
	case codexMethodItemAgentMessageDelta:
		return d.ownedMessageDelta(stringField(p, "itemId"), stringField(p, "delta"), true), true
	case codexMethodAgentMessage:
		return []agentbridge.Event{{Kind: agentbridge.EventProgress, Text: stringField(p, "text")}}, true
	default:
		return nil, false
	}
}

func (d *protocolDriver) ownedMessageDelta(id, text string, delta bool) []agentbridge.Event {
	if d.owned.streamed == nil {
		d.owned.streamed = map[string]string{}
	}
	// Only explicitly final text can enter the reducer's output buffer. The
	// terminal result supplies the selected complete item for phase-less builds.
	if stringField(d.owned.messages[id], "phase") != "final_answer" {
		if text == "" {
			return nil
		}
		return []agentbridge.Event{{Kind: agentbridge.EventProgress, Text: text}}
	}
	previous := d.owned.streamed[id]
	if delta {
		d.owned.streamed[id] = previous + text
	} else {
		d.owned.streamed[id] = text
		if !strings.HasPrefix(text, previous) {
			return nil
		}
		text = strings.TrimPrefix(text, previous)
	}
	if text == "" {
		return nil
	}
	return []agentbridge.Event{{Kind: agentbridge.EventTextDelta, Text: text}}
}

// Explicit commentary is never a result. Older phase-less item streams use
// their last message rather than concatenating every intermediate response.
func finalAnswer(items []any) string {
	var final []string
	var legacy string
	hasPhase := false
	for _, value := range items {
		item, _ := value.(map[string]any)
		if stringField(item, "type") != "agentMessage" {
			continue
		}
		text := stringField(item, "text")
		hasPhase = hasPhase || stringField(item, "phase") != ""
		switch stringField(item, "phase") {
		case "final_answer":
			final = append(final, text)
		case "":
			legacy = text
		}
	}
	if len(final) > 0 {
		return strings.Join(final, "\n\n")
	}
	if hasPhase {
		return ""
	}
	return legacy
}

func (d *protocolDriver) ownedResultEvents(raw agentbridge.RawEvent, events []agentbridge.Event) []agentbridge.Event {
	if d.owned.id == "" {
		return events
	}
	for i := range events {
		if events[i].Kind != agentbridge.EventResult {
			continue
		}
		items, _ := mapField(params(raw), "turn")["items"].([]any)
		for _, value := range items {
			item, _ := value.(map[string]any)
			d.rememberMessage(item)
		}
		items = nil
		for _, id := range d.owned.order {
			items = append(items, d.owned.messages[id])
		}
		events[i].Result.Output = finalAnswer(items)
		d.owned.done = true
		d.turnStarted = false
		return append(d.ownedDecision(raw, "terminal", string(events[i].Result.Status)), events...)
	}
	return events
}
