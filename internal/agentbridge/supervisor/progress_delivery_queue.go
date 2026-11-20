package supervisor

import "github.com/teamswyg/riido-daemon/internal/agentbridge"

func (d *progressDelivery) enqueue(ev agentbridge.Event) {
	d.mu.Lock()
	defer d.mu.Unlock()
	// Large single display frames also have an authoritative local copy. Avoid
	// allocating an unbounded remote queue for them.
	if len(ev.Text) > maxDeliveryGroupBytes {
		d.superseded++
		return
	}
	n := len(d.queue)
	separator := ""
	if ev.Kind == agentbridge.EventLog {
		separator = "\n"
	}
	if n > 0 && d.queue[n-1].Kind == ev.Kind && len(d.queue[n-1].Text)+len(separator)+len(ev.Text) <= maxDeliveryGroupBytes {
		d.queue[n-1].Text += separator + ev.Text
		d.bytes += len(separator) + len(ev.Text)
	} else {
		d.queue = append(d.queue, ev)
		d.bytes += len(ev.Text)
	}
	for len(d.queue) > maxDeliveryGroupEvents || d.bytes > maxProgressQueueBytes {
		d.bytes -= len(d.queue[0].Text)
		d.queue[0] = agentbridge.Event{}
		d.queue = d.queue[1:]
		d.superseded++
	}
	select {
	case d.wake <- struct{}{}:
	default:
	}
}

func (d *progressDelivery) next() (agentbridge.Event, bool) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if len(d.queue) == 0 {
		return agentbridge.Event{}, false
	}
	ev := d.queue[0]
	d.bytes -= len(ev.Text)
	d.queue[0] = agentbridge.Event{}
	d.queue = d.queue[1:]
	return ev, true
}
