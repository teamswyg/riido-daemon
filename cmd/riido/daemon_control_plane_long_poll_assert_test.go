package main

import (
	"testing"
	"time"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
)

func assertDefaultLongPollWait(t *testing.T, pollSeen <-chan assignmentcontract.PollRequest) {
	t.Helper()
	select {
	case poll := <-pollSeen:
		if poll.WaitMs != 30000 {
			t.Fatalf("wait_ms = %d, want 30000", poll.WaitMs)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for poll request")
	}
}
