package executionproof

// HostWork reserves daemon-owned preparation or verification before its worker
// starts. It does not contribute to the successful provider Start counter.
type HostWork struct {
	tracker  *Tracker
	finished bool
}

func (t *Tracker) BeginHostWork() (*HostWork, error) {
	if t == nil {
		return nil, errAdmission
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	if !t.reserveLocked() {
		return nil, errAdmission
	}
	return &HostWork{tracker: t}, nil
}

// Finished must be called only after the worker's synchronous commands have
// returned from Wait, and before notifying the supervisor of completion.
func (w *HostWork) Finished() {
	if w == nil || w.tracker == nil {
		return
	}
	t := w.tracker
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.sealed || w.finished {
		return
	}
	w.finished = true
	t.pending--
}
