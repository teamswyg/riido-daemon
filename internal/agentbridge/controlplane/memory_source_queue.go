package controlplane

import (
	"context"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
)

// Enqueue appends a task to the internal queue (test/daemon helper).
func (s *MemorySource) Enqueue(req bridge.TaskRequest) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.queue = append(s.queue, req)
}

func (s *MemorySource) ClaimTask(_ context.Context, _ string) (*bridge.TaskRequest, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.queue) == 0 {
		return nil, nil
	}
	req := s.queue[0]
	s.queue = s.queue[1:]
	return &req, nil
}
