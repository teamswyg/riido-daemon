package runtimeactor

import (
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func expectTaskStatus(
	t *testing.T,
	results <-chan agentbridge.Result,
	want agentbridge.ResultStatus,
	timeout string,
) {
	t.Helper()
	select {
	case res := <-results:
		if res.Status != want {
			t.Fatalf("status: %s", res.Status)
		}
	case <-time.After(2 * time.Second):
		t.Fatal(timeout)
	}
}
