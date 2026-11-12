package codex

import (
	"fmt"
	"strings"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

type ownedTurn struct {
	id         string
	done       bool
	early      []agentbridge.RawEvent
	earlyBytes int
	messages   map[string]map[string]any
	order      []string
	streamed   map[string]string
}

func notificationTurnID(p map[string]any) string {
	for _, id := range []string{stringField(p, "turnId"), stringField(p, "turn_id"), stringField(mapField(p, "turn"), "id")} {
		if id != "" {
			return id
		}
	}
	return ""
}

// Correlation belongs before Translate: a child completion otherwise terminates
// the session reducer and its process, even while the parent is still running.
func (d *protocolDriver) guardOwnedNotification(raw agentbridge.RawEvent) ([]agentbridge.Event, bool) {
	p := params(raw)
	thread, turn := threadIDFromParams(p), notificationTurnID(p)
	method := codexMethod(strings.TrimPrefix(raw.Type, rawFrameNotificationPrefix))
	if d.threadID == "" && thread != "" && (d.awaitingReply(codexMethodThreadStart) || d.awaitingReply(codexMethodThreadResume)) {
		// Only our correlated response may establish the root thread identity.
		return nil, true
	}
	if d.threadID != "" && thread != "" && thread != d.threadID {
		return d.ownedDecision(raw, "ignored", "thread_mismatch"), true
	}
	if d.owned.id == "" {
		if d.threadID != "" && d.awaitingReply(codexMethodTurnStart) && runScopedNotification(method) {
			if len(d.owned.early) >= 256 || d.owned.earlyBytes+len(raw.Bytes) > 2<<20 {
				return d.failedEvents("codex: turn identity reply exceeded notification buffer"), true
			}
			d.owned.early = append(d.owned.early, raw)
			d.owned.earlyBytes += len(raw.Bytes)
			return nil, true
		}
		if d.awaitingReply(codexMethodTurnStart) && (method == codexMethodThreadStatusChanged || method == codexMethodThreadStatusAlt) {
			return nil, true
		}
		return nil, false // Legacy protocol without a correlated turn identity.
	}
	if turn != "" && turn != d.owned.id {
		return d.ownedDecision(raw, "ignored", "turn_mismatch"), true
	}
	if runScopedNotification(method) && (thread != d.threadID || turn != d.owned.id) {
		return d.ownedDecision(raw, "ignored", "missing_identity"), true
	}
	if method == codexMethodThreadStatusChanged || method == codexMethodThreadStatusAlt {
		return d.ownedDecision(raw, "observed", "await_owned_turn_terminal"), true
	}
	if d.owned.done {
		return nil, true
	}
	return nil, false
}

func (d *protocolDriver) awaitingReply(method codexMethod) bool {
	for _, request := range d.pending {
		if request.method == method {
			return true
		}
	}
	return false
}

func runScopedNotification(method codexMethod) bool {
	switch method {
	case codexMethodTurnStarted, codexMethodTurnStartedSlash,
		codexMethodTurnCompleted, codexMethodTurnCompleteSlash,
		codexMethodTurnError, codexMethodTurnErrorSlash, codexMethodTurnFailedSlash,
		codexMethodError, codexMethodAgentMessage, codexMethodItemAgentMessageDelta,
		codexMethodReasoning, codexMethodItemStarted, codexMethodItemUpdated, codexMethodItemCompleted:
		return true
	default:
		return false
	}
}

func (d *protocolDriver) ownedDecision(raw agentbridge.RawEvent, decision, reason string) []agentbridge.Event {
	p := params(raw)
	return []agentbridge.Event{{Kind: agentbridge.EventLog, Text: fmt.Sprintf(
		"codex event decision=%s reason=%s target_thread=%q target_turn=%q event=%q thread=%q turn=%q",
		decision, reason, d.threadID, d.owned.id, raw.Type, threadIDFromParams(p), notificationTurnID(p))}}
}
