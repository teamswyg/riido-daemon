package controlplane

import (
	"context"
	"sort"
)

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
