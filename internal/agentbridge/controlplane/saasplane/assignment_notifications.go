package saasplane

import (
	"context"
	"log"
	"math/rand/v2"
	"sync"
	"time"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
)

const assignmentEventsSchema = "riido-daemon-assignment-events.v1"

type assignmentNotification struct {
	SchemaVersion           string                                   `json:"schema_version"`
	Bindings                []assignmentcontract.AgentRuntimeBinding `json:"bindings"`
	ConnectionRevision      string                                   `json:"connection_revision,omitempty"`
	ConnectedPrincipalCount int                                      `json:"connected_principal_count,omitempty"`
	AgentIDs                []string                                 `json:"agent_ids"`
}

type pendingAssignmentNotice struct {
	generation uint64
	retryAfter time.Time
}
type assignmentNotifications struct {
	mu              sync.Mutex
	started, legacy bool
	cancel          context.CancelFunc
	wake            chan struct{}
	done            chan struct{}
	bindings        []assignmentcontract.AgentRuntimeBinding
	pending         map[string]pendingAssignmentNotice
	generation      uint64
}

// AssignmentNotifications starts exactly one stream per Plane, independently of
// provider sessions. The wake channel is coalesced and stays open until shutdown.
func (p *Plane) AssignmentNotifications(ctx context.Context) <-chan struct{} {
	if !p.cfg.AssignmentEvents || p.cfg.DeviceSecret == "" || !p.dynamicBindingsEnabled() {
		return nil
	}
	p.noticeOnce.Do(func() {
		ctx, cancel := context.WithCancel(ctx)
		p.notices.mu.Lock()
		p.notices.started = true
		p.notices.cancel = cancel
		p.notices.wake = make(chan struct{}, 1)
		p.notices.done = make(chan struct{})
		p.notices.pending = map[string]pendingAssignmentNotice{}
		p.notices.mu.Unlock()
		go p.assignmentNotificationLoop(ctx)
	})
	return p.notices.wake
}

func (p *Plane) stopAssignmentNotifications() {
	p.notices.mu.Lock()
	cancel, done := p.notices.cancel, p.notices.done
	p.notices.mu.Unlock()
	if cancel != nil {
		cancel()
		<-done
	}
}

func (p *Plane) assignmentNotificationLoop(ctx context.Context) {
	defer close(p.notices.done)
	delay := time.Second
	for ctx.Err() == nil {
		received, unsupported, err := p.readAssignmentNotifications(ctx)
		if ctx.Err() != nil {
			return
		}
		if unsupported {
			p.notices.mu.Lock()
			p.notices.legacy = true
			p.notices.mu.Unlock()
			p.wakeAssignmentClaims()
			log.Printf("event=daemon_assignment_stream_legacy_fallback")
			return
		}
		if received {
			delay = time.Second
		}
		if err != nil {
			log.Printf("event=daemon_assignment_stream_reconnect retry_seconds=%d", int(delay.Seconds()))
		}
		timer := time.NewTimer(delay + time.Duration(rand.Int64N(int64(delay/4))))
		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-timer.C:
		}
		delay = min(30*time.Second, delay*2)
	}
}

func (p *Plane) acceptAssignmentNotification(ctx context.Context, event assignmentNotification) {
	p.observeDeviceConnections(ctx, AgentRuntimeBindingListResponse{Bindings: event.Bindings, ConnectionRevision: event.ConnectionRevision, ConnectedPrincipalCount: event.ConnectedPrincipalCount})
	p.notices.mu.Lock()
	defer p.notices.mu.Unlock()
	p.notices.bindings = cloneAgentRuntimeBindings(event.Bindings)
	allowed := map[string]bool{}
	for _, b := range event.Bindings {
		allowed[b.AgentID] = true
	}
	for id := range p.notices.pending {
		if !allowed[id] {
			delete(p.notices.pending, id)
		}
	}
	for _, id := range event.AgentIDs {
		if !allowed[id] {
			continue
		}
		p.notices.generation++
		p.notices.pending[id] = pendingAssignmentNotice{generation: p.notices.generation}
	}
	p.wakeAssignmentClaims()
}

func (p *Plane) wakeAssignmentClaims() {
	select {
	case p.notices.wake <- struct{}{}:
	default:
	}
}
