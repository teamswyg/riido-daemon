package executionproof

import "testing"

func TestTrackerCountsOnlyStartedAndObservedInvocations(t *testing.T) {
	tracker := NewTracker(true)
	for range 2 {
		attempt, err := tracker.BeginInvocation()
		if err != nil {
			t.Fatal(err)
		}
		attempt.Started()
		attempt.Exited(true)
	}
	proof, ok := tracker.Seal()
	if !ok || proof.Kind != "local_invocations_quiescent" || proof.StartedInvocations != 2 {
		t.Fatalf("fully observed invocations = %#v, %v; want quiescent with count 2", proof, ok)
	}
}

func TestTrackerUnknownNeverUpgradesAfterSeal(t *testing.T) {
	tracker := NewTracker(true)
	attempt, err := tracker.BeginInvocation()
	if err != nil {
		t.Fatal(err)
	}
	attempt.Started()
	if proof, ok := tracker.Seal(); ok || proof != (Proof{}) {
		t.Fatalf("pending invocation became positive: %#v, %v", proof, ok)
	}
	attempt.Exited(true)
	if proof, ok := tracker.Seal(); ok || proof != (Proof{}) {
		t.Fatalf("late exit upgraded frozen unknown: %#v, %v", proof, ok)
	}
	if _, err := tracker.BeginInvocation(); err == nil {
		t.Fatal("new invocation admitted after seal")
	}
}

func TestTrackerEarlierUnknownSurvivesLaterSuccess(t *testing.T) {
	tracker := NewTracker(true)
	first, err := tracker.BeginInvocation()
	if err != nil {
		t.Fatal(err)
	}
	first.StartUnknown()
	first.Exited(true)
	second, err := tracker.BeginInvocation()
	if err != nil {
		t.Fatal(err)
	}
	second.Started()
	second.Exited(true)
	if proof, ok := tracker.Seal(); ok || proof != (Proof{}) {
		t.Fatalf("later success erased unknown start: %#v, %v", proof, ok)
	}
}
