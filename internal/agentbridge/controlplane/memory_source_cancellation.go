package controlplane

import "context"

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
