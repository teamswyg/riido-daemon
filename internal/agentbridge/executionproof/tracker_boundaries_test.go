package executionproof

import (
	"math"
	"testing"
)

func TestTrackerHostWorkPreventsPrematureSeal(t *testing.T) {
	tracker := NewTracker(true)
	work, err := tracker.BeginHostWork()
	if err != nil {
		t.Fatal(err)
	}
	if proof, ok := tracker.Seal(); ok || proof != (Proof{}) {
		t.Fatalf("pending host work became positive: %#v, %v", proof, ok)
	}
	work.Finished()
	if proof, ok := tracker.Seal(); ok || proof != (Proof{}) {
		t.Fatalf("late host completion upgraded unknown: %#v, %v", proof, ok)
	}
	if _, err := tracker.BeginHostWork(); err == nil {
		t.Fatal("host work admitted after seal")
	}
}

func TestTrackerHostWorkDoesNotCountProviderStarts(t *testing.T) {
	tracker := NewTracker(true)
	work, err := tracker.BeginHostWork()
	if err != nil {
		t.Fatal(err)
	}
	work.Finished()
	work.Finished()
	attempt, err := tracker.BeginInvocation()
	if err != nil {
		t.Fatal(err)
	}
	attempt.Started()
	attempt.Started()
	attempt.Exited(true)
	attempt.Exited(true)
	proof, ok := tracker.Seal()
	if !ok || proof.Kind != "local_invocations_quiescent" || proof.StartedInvocations != 1 {
		t.Fatalf("host or duplicate notifications changed count: %#v, %v", proof, ok)
	}
}

func TestTrackerUncoveredAndMissingEvidenceRemainUnknown(t *testing.T) {
	for _, name := range []string{"uncovered", "pending-start", "missing-exit", "exit-before-start", "duplicate-false-exit"} {
		t.Run(name, func(t *testing.T) {
			tracker := NewTracker(name != "uncovered")
			attempt, err := tracker.BeginInvocation()
			if err != nil {
				t.Fatal(err)
			}
			switch name {
			case "uncovered":
				attempt.Started()
				attempt.Exited(true)
			case "missing-exit":
				attempt.Started()
				attempt.Exited(false)
				attempt.Exited(true)
			case "exit-before-start":
				attempt.Exited(true)
				attempt.Started()
			case "duplicate-false-exit":
				attempt.Started()
				attempt.Exited(true)
				attempt.Exited(false)
			}
			if proof, ok := tracker.Seal(); ok || proof != (Proof{}) {
				t.Fatalf("missing coverage became positive: %#v, %v", proof, ok)
			}
		})
	}
}

func TestTrackerCounterOverflowCannotProduceProof(t *testing.T) {
	for _, name := range []string{"invocation-admission", "host-admission", "started"} {
		t.Run(name, func(t *testing.T) {
			tracker := NewTracker(true)
			if name == "started" {
				tracker.started = math.MaxUint64
				attempt, err := tracker.BeginInvocation()
				if err != nil {
					t.Fatal(err)
				}
				attempt.Started()
				attempt.Exited(true)
			} else {
				tracker.pending = math.MaxUint64
				if name == "invocation-admission" {
					if _, err := tracker.BeginInvocation(); err == nil {
						t.Fatal("overflowing invocation admitted")
					}
				} else if _, err := tracker.BeginHostWork(); err == nil {
					t.Fatal("overflowing host work admitted")
				}
			}
			if proof, ok := tracker.Seal(); ok || proof != (Proof{}) {
				t.Fatalf("overflow became positive: %#v, %v", proof, ok)
			}
		})
	}
}

func TestTrackerNilAndZeroValueNeverProveCoverage(t *testing.T) {
	for _, tracker := range []*Tracker{nil, {}} {
		if proof, ok := tracker.Seal(); ok || proof != (Proof{}) {
			t.Fatalf("missing tracker coverage: %#v, %v", proof, ok)
		}
		if _, err := tracker.BeginInvocation(); err == nil {
			t.Fatal("nil/sealed tracker admitted invocation")
		}
		if _, err := tracker.BeginHostWork(); err == nil {
			t.Fatal("nil/sealed tracker admitted host work")
		}
	}
	var attempt *Attempt
	attempt.Started()
	attempt.StartUnknown()
	attempt.Exited(true)
	var work *HostWork
	work.Finished()
}
