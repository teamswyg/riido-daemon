package saasplane

import (
	"context"
	"errors"
	"strings"
)

func sendAndCloseCancelWatcher(s *planeState, executionID string, cause error) {
	ch := s.cancelWatchers[executionID]
	if ch == nil {
		return
	}
	if cause != nil {
		select {
		case ch <- cause:
		default:
		}
	}
	closeCancelWatcher(s, executionID)
}

func closeCancelWatcher(s *planeState, executionID string) {
	ch := s.cancelWatchers[executionID]
	if ch == nil {
		return
	}
	close(ch)
	delete(s.cancelWatchers, executionID)
}

func closeCancelWatcherIfCurrent(s *planeState, executionID string, ch chan error) {
	if s.cancelWatchers[executionID] != ch {
		return
	}
	closeCancelWatcher(s, executionID)
}

func (p *Plane) closeCancelWatcherWhenContextDone(ctx context.Context, executionID string, ch chan error) {
	select {
	case <-ctx.Done():
	case <-p.done:
		return
	}
	_ = p.withState(context.Background(), func(s *planeState) {
		closeCancelWatcherIfCurrent(s, executionID, ch)
	})
}

func (p *Plane) WatchCancellation(ctx context.Context, executionID string) (<-chan error, error) {
	executionID = strings.TrimSpace(executionID)
	if executionID == "" {
		return nil, errors.New("saasplane: empty executionID")
	}
	ch := make(chan error, 1)
	err := p.withState(ctx, func(s *planeState) {
		closeCancelWatcher(s, executionID)
		s.cancelWatchers[executionID] = ch
	})
	if err != nil {
		return nil, err
	}
	go p.closeCancelWatcherWhenContextDone(ctx, executionID, ch)
	return ch, nil
}
