package controlplane

import "context"

func (s *FileQueueSource) WatchCancellation(_ context.Context, _ string) (<-chan error, error) {
	// File queue has no out-of-band cancel channel; return a closed
	// channel so the caller can range over it without blocking.
	ch := make(chan error)
	close(ch)
	return ch, nil
}
