// Package executionproof tracks private, local invocation observations for one
// claim. It does not prove remote runtime or descendant termination.
package executionproof

import (
	"errors"
	"math"
	"sync"
)

const QuiescentKind = "local_invocations_quiescent"

var errAdmission = errors.New("local execution observation admission is closed")

type Proof struct {
	Kind               string
	StartedInvocations uint64
}

// Tracker has complete coverage only when its source owns every admission for
// the claim. A zero value or an uncovered tracker never produces a proof.
type Tracker struct {
	mu                       sync.Mutex
	covered, unknown, sealed bool
	pending, started         uint64
	proof                    Proof
	known                    bool
}

// Attempt reserves one provider Start before the call. Its methods must not be
// driven by provider output: only the process boundary can supply observations.
type Attempt struct {
	tracker           *Tracker
	started, finished bool
}

func NewTracker(covered bool) *Tracker { return &Tracker{covered: covered} }

func (t *Tracker) BeginInvocation() (*Attempt, error) {
	if t == nil {
		return nil, errAdmission
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	if !t.reserveLocked() {
		return nil, errAdmission
	}
	return &Attempt{tracker: t}, nil
}

func (t *Tracker) reserveLocked() bool {
	if t.sealed {
		return false
	}
	if t.pending == math.MaxUint64 {
		t.unknown = true
		return false
	}
	t.pending++
	return true
}

// Started records a successful Process.Start, once per invocation.
func (a *Attempt) Started() {
	if a == nil || a.tracker == nil {
		return
	}
	t := a.tracker
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.sealed || a.finished || a.started {
		return
	}
	a.started = true
	if t.started == math.MaxUint64 {
		t.unknown = true
	} else {
		t.started++
	}
}

// StartUnknown closes the admission conservatively. A generic Start error is
// not proof that no process was created, even if a later invocation succeeds.
func (a *Attempt) StartUnknown() {
	if a == nil || a.tracker == nil {
		return
	}
	t := a.tracker
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.sealed {
		return
	}
	t.unknown = true
	a.finishLocked()
}

// Exited closes an invocation only with the core's direct-child Wait
// observation. Missing evidence remains unknown, including an exit before Start.
func (a *Attempt) Exited(observed bool) {
	if a == nil || a.tracker == nil {
		return
	}
	t := a.tracker
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.sealed {
		return
	}
	if !observed || !a.started {
		t.unknown = true
	}
	a.finishLocked()
}

func (a *Attempt) finishLocked() {
	if a.finished {
		return
	}
	a.finished = true
	a.tracker.pending--
}

// Seal rejects future admissions and freezes the first verdict. Completing a
// pending invocation later cannot upgrade that terminal unknown observation.
func (t *Tracker) Seal() (Proof, bool) {
	if t == nil {
		return Proof{}, false
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	if !t.sealed {
		t.sealed = true
		t.known = t.covered && !t.unknown && t.pending == 0
		if t.known {
			t.proof = Proof{Kind: QuiescentKind, StartedInvocations: t.started}
		}
	}
	return t.proof, t.known
}

func (*Tracker) String() string   { return "[redacted execution tracker]" }
func (*Tracker) GoString() string { return "[redacted execution tracker]" }
