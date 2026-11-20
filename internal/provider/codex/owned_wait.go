package codex

import (
	"fmt"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func (d *protocolDriver) trackOwnedWait(method codexMethod, item map[string]any) []agentbridge.Event {
	id := stringField(item, "id")
	if id == "" || stringField(item, "type") != "collabAgentToolCall" ||
		stringField(item, "tool") != "wait" || stringField(item, "senderThreadId") != d.threadID {
		return nil
	}
	_, active := d.owned.waits[id]
	if method == codexMethodItemCompleted {
		if !active {
			return nil
		}
		delete(d.owned.waits, id)
		return d.ownedWaitEvents(id, "completed", nil)
	}
	if active {
		return nil
	}

	receivers, _ := item["receiverThreadIds"].([]any)
	targets := make(map[string]string, len(receivers))
	var threads []string
	for _, receiver := range receivers {
		thread, _ := receiver.(string)
		if _, exists := targets[thread]; exists || thread == "" || thread == d.threadID {
			continue
		}
		targets[thread] = ""
		threads = append(threads, thread)
	}
	if len(targets) == 0 {
		return nil
	}
	if d.owned.waits == nil {
		d.owned.waits = make(map[string]map[string]string)
	}
	d.owned.waits[id] = targets
	return d.ownedWaitEvents(id, "started", threads)
}

func (d *protocolDriver) ownedWaitEvents(id, status string, threads []string) []agentbridge.Event {
	return []agentbridge.Event{
		{Kind: agentbridge.EventSemanticActivity},
		{Kind: agentbridge.EventLog, Text: fmt.Sprintf(
			"codex wait %s target_thread=%q target_turn=%q item=%q receiver_threads=%q",
			status, d.threadID, d.owned.id, id, threads)},
	}
}

func (o *ownedTurn) observeWaitTarget(method codexMethod, thread, turn string, p map[string]any) bool {
	if len(o.waits) == 0 || turn == "" {
		return false
	}
	switch method {
	case codexMethodTurnStarted, codexMethodTurnStartedSlash:
	case codexMethodItemAgentMessageDelta:
		if stringField(p, "itemId") == "" || stringField(p, "delta") == "" {
			return false
		}
	case codexMethodItemCompleted:
		item := mapField(p, "item")
		if stringField(item, "id") == "" || stringField(item, "type") == "" {
			return false
		}
	case codexMethodTurnCompleted, codexMethodTurnCompleteSlash:
		switch stringField(mapField(p, "turn"), "status") {
		case "completed", "failed", "interrupted":
		default:
			return false
		}
	default:
		return false
	}

	activity := false
	for _, targets := range o.waits {
		observedTurn, waiting := targets[thread]
		if !waiting || (observedTurn != "" && observedTurn != turn) {
			continue
		}
		switch method {
		case codexMethodTurnStarted, codexMethodTurnStartedSlash:
			if observedTurn != "" {
				continue
			}
			targets[thread] = turn
		case codexMethodTurnCompleted, codexMethodTurnCompleteSlash:
			delete(targets, thread)
		default:
			targets[thread] = turn
		}
		activity = true
	}
	return activity
}
