package executionproof

import (
	"sync"
	"testing"
)

func TestTrackerConcurrentCompletionCountsEachProviderOnce(t *testing.T) {
	tracker := NewTracker(true)
	var workers sync.WaitGroup
	for range 64 {
		attempt, err := tracker.BeginInvocation()
		if err != nil {
			t.Fatal(err)
		}
		work, err := tracker.BeginHostWork()
		if err != nil {
			t.Fatal(err)
		}
		attempt.Started()
		for range 2 {
			workers.Go(func() {
				attempt.Started()
				attempt.Exited(true)
				work.Finished()
			})
		}
	}
	workers.Wait()
	proof, ok := tracker.Seal()
	if !ok || proof.Kind != "local_invocations_quiescent" || proof.StartedInvocations != 64 {
		t.Fatalf("concurrent duplicate completions changed observation: %#v, %v", proof, ok)
	}
}

func TestTrackerAdmissionAndSealHaveOneOrdering(t *testing.T) {
	for range 100 {
		tracker := NewTracker(true)
		gate := make(chan struct{})
		attempts := make(chan *Attempt, 1)
		proofs := make(chan bool, 1)
		go func() {
			<-gate
			attempt, _ := tracker.BeginInvocation()
			attempts <- attempt
		}()
		go func() {
			<-gate
			_, ok := tracker.Seal()
			proofs <- ok
		}()
		close(gate)
		attempt, positive := <-attempts, <-proofs
		if (attempt != nil) == positive {
			t.Fatalf("admission and positive seal both won or both lost: admitted=%v positive=%v", attempt != nil, positive)
		}
		attempt.Started()
		attempt.Exited(true)
		if proof, ok := tracker.Seal(); ok != positive || (ok && proof.StartedInvocations != 0) {
			t.Fatalf("late invocation changed frozen observation: %#v, %v", proof, ok)
		}
	}
}

func TestTrackerConcurrentSealsReturnSameSnapshot(t *testing.T) {
	tracker := NewTracker(true)
	attempt, err := tracker.BeginInvocation()
	if err != nil {
		t.Fatal(err)
	}
	attempt.Started()
	attempt.Exited(true)
	var callers sync.WaitGroup
	for range 32 {
		callers.Go(func() {
			proof, ok := tracker.Seal()
			if !ok || proof.Kind != "local_invocations_quiescent" || proof.StartedInvocations != 1 {
				t.Errorf("concurrent seal changed snapshot: %#v, %v", proof, ok)
			}
		})
	}
	callers.Wait()
}
