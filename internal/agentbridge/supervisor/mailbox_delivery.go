package supervisor

import (
	"context"
	"reflect"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

const (
	maxDeliveryGroupEvents = 64
	maxDeliveryGroupBytes  = 64 * 1024
)

// Only consume events already waiting behind the current delivery. A lone first
// token never waits for a flush timer. At most one boundary envelope is read
// ahead; it is processed immediately after the group, without recursive draining.
func (a *Actor) handleMailboxDelivery(ctx context.Context, msg envelope, inFlight map[string]*runningTask, detached *[]detachedReport) bool {
	var boundary *envelope
	if msg.taskEvent != nil && plainDeliveryEvent(msg.taskEvent.event) {
		msg, boundary = a.coalesceQueuedDelivery(msg)
	}
	changed := a.handleMailboxMessage(ctx, msg, inFlight, detached)
	if boundary != nil {
		changed = a.handleMailboxMessage(ctx, *boundary, inFlight, detached) || changed
	}
	return changed
}

func (a *Actor) coalesceQueuedDelivery(msg envelope) (envelope, *envelope) {
	group := *msg.taskEvent
	group.originalEvents = []agentbridge.Event{group.event}
	group.originalQueuedAt = []time.Time{group.queuedAt}
	for len(group.originalEvents) < maxDeliveryGroupEvents && len(group.event.Text) < maxDeliveryGroupBytes {
		select {
		case next := <-a.mailbox:
			item := next.taskEvent
			if item == nil || item.taskID != group.taskID || item.event.Kind != group.event.Kind || !plainDeliveryEvent(item.event) || len(group.event.Text)+len(item.event.Text)+1 > maxDeliveryGroupBytes {
				msg.taskEvent = &group
				return msg, &next
			}
			if group.event.Kind == agentbridge.EventLog {
				group.event.Text += "\n"
			}
			group.event.Text += item.event.Text
			group.originalEvents = append(group.originalEvents, item.event)
			group.originalQueuedAt = append(group.originalQueuedAt, item.queuedAt)
		default:
			msg.taskEvent = &group
			return msg, nil
		}
	}
	msg.taskEvent = &group
	return msg, nil
}

// Structured events retain their identity and ordering, including future fields.
func plainDeliveryEvent(ev agentbridge.Event) bool {
	if ev.Kind != agentbridge.EventTextDelta && ev.Kind != agentbridge.EventLog {
		return false
	}
	ev.Text = ""
	ev.At = time.Time{}
	return reflect.DeepEqual(ev, agentbridge.Event{Kind: ev.Kind})
}
