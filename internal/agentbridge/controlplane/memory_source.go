package controlplane

import (
	"context"
	"sort"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
)

// MemorySource is the simplest TaskSourcePort: tasks live in a FIFO,
// runtimes in a map. Intended for tests, offline mode, and bootstrap.
//
// All state is owned by the calling goroutine -- the source itself is
// NOT a separate actor. Callers (daemon main goroutine or a
// SupervisorActor) serialize access. We do not use sync.Mutex here.
type MemorySource struct {
	queue     []bridge.TaskRequest
	runtimes  map[string]*RegisteredRuntime
	cancelChs map[string]chan error
	now       func() time.Time
}

func NewMemorySource() *MemorySource {
	return &MemorySource{
		runtimes:  map[string]*RegisteredRuntime{},
		cancelChs: map[string]chan error{},
		now:       time.Now,
	}
}

// Enqueue appends a task to the internal queue (test/daemon helper).
func (s *MemorySource) Enqueue(req bridge.TaskRequest) {
	s.queue = append(s.queue, req)
}

// Registered returns a snapshot of registered runtimes sorted by id.
func (s *MemorySource) Registered() []RegisteredRuntime {
	out := make([]RegisteredRuntime, 0, len(s.runtimes))
	for _, r := range s.runtimes {
		out = append(out, *r)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].RuntimeID < out[j].RuntimeID })
	return out
}

func (s *MemorySource) RegisterRuntime(_ context.Context, rt RuntimeRegistration) error {
	if rt.RuntimeID == "" {
		return controlPlaneErrorf(ErrControlPlaneRuntime, "memory.register-runtime", "empty RuntimeID")
	}
	s.runtimes[rt.RuntimeID] = &RegisteredRuntime{
		RuntimeRegistration: rt,
		LastHeartbeat:       s.now(),
	}
	return nil
}

func (s *MemorySource) DeregisterRuntime(_ context.Context, runtimeID string) error {
	if _, ok := s.runtimes[runtimeID]; !ok {
		return controlPlaneErrorf(ErrControlPlaneRuntime, "memory.deregister-runtime", "unknown runtime %q", runtimeID)
	}
	delete(s.runtimes, runtimeID)
	return nil
}

func (s *MemorySource) Heartbeat(_ context.Context, hb RuntimeHeartbeat) error {
	r, ok := s.runtimes[hb.RuntimeID]
	if !ok {
		return controlPlaneErrorf(ErrControlPlaneRuntime, "memory.heartbeat", "heartbeat for unknown runtime %q", hb.RuntimeID)
	}
	r.LastHeartbeat = s.now()
	applyHeartbeat(&r.RuntimeRegistration, hb)
	return nil
}

func (s *MemorySource) ClaimTask(_ context.Context, _ string) (*bridge.TaskRequest, error) {
	if len(s.queue) == 0 {
		return nil, nil
	}
	req := s.queue[0]
	s.queue = s.queue[1:]
	return &req, nil
}

func (s *MemorySource) WatchCancellation(_ context.Context, taskID string) (<-chan error, error) {
	if taskID == "" {
		return nil, controlPlaneErrorf(ErrControlPlaneInput, "memory.watch-cancellation", "empty taskID")
	}
	ch := make(chan error, 1)
	s.cancelChs[taskID] = ch
	return ch, nil
}

// Cancel delivers a cancellation cause to a previously watched task.
// If no watcher exists, the cause is dropped (no-op).
func (s *MemorySource) Cancel(taskID string, cause error) {
	ch, ok := s.cancelChs[taskID]
	if !ok {
		return
	}
	select {
	case ch <- cause:
	default:
	}
}
