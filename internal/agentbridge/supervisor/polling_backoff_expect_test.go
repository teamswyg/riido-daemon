package supervisor

import (
	"testing"
	"time"
)

func expectIdlePoll(t *testing.T, source *idlePollSource) {
	t.Helper()
	select {
	case <-source.claims:
	case <-time.After(time.Second):
		t.Fatal("first poll did not happen")
	}
}

func assertNoIdlePollBeforeBackoff(t *testing.T, source *idlePollSource) {
	t.Helper()
	select {
	case runtimeID := <-source.claims:
		t.Fatalf("idle poll happened before backoff elapsed: %s", runtimeID)
	case <-time.After(50 * time.Millisecond):
	}
}

func expectIdlePollResumes(t *testing.T, source *idlePollSource) {
	t.Helper()
	select {
	case <-source.claims:
	case <-time.After(250 * time.Millisecond):
		t.Fatal("idle poll did not resume after backoff interval")
	}
}
