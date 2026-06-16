package saasplane

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
