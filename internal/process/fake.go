package process

import "context"

// Fake is a deterministic Process used by tests. It spawns no real OS
// process; the test drives stdout / stderr / exit signals through the
// FakeRunning's Emit* methods.
//
// The fake intentionally uses channels for all state to serve as a reference
// for the actor-style discipline used by daemon runtime packages.
type Fake struct {
	// NextRunning lets a test pre-construct the FakeRunning so it can
	// drive stdout/stderr/exit before Start is even called. If nil,
	// Start allocates a fresh FakeRunning.
	NextRunning *FakeRunning
}

func NewFake() *Fake { return &Fake{} }

func (f *Fake) Start(_ context.Context, cmd Command) (RunningProcess, error) {
	r := f.NextRunning
	if r == nil {
		r = NewFakeRunning()
	}
	f.NextRunning = nil
	r.cmd = cmd
	select {
	case r.started <- cmd:
	default:
	}
	return r, nil
}
