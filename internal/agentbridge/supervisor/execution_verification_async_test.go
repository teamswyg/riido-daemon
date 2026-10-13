package supervisor

import (
	"context"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
)

func TestExecutionBlockingCheck(t *testing.T) {
	marker := os.Args[len(os.Args)-1]
	if !strings.Contains(marker, "riido-check-started") {
		return
	}
	if err := os.WriteFile(marker, []byte("started"), 0o600); err != nil {
		t.Fatal(err)
	}
	time.Sleep(15 * time.Second)
}

func TestExecutionVerificationDoesNotBlockOtherTaskOrCancellation(t *testing.T) {
	marker := filepath.Join(t.TempDir(), "riido-check-started")
	adapter, proc, reporter, source, _ := startCorrectionActor(t, 1, []string{os.Args[0], "-test.run=^TestExecutionBlockingCheck$", "--", marker})
	req, p := awaitCorrectionRequest(t, adapter, proc)
	if err := os.WriteFile(filepath.Join(req.Cwd, "result.txt"), []byte("fixed\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	p.EmitStdout([]byte("fixed"))
	deadline := time.Now().Add(3 * time.Second)
	for {
		if _, err := os.Stat(marker); err == nil {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("host check not started")
		}
		time.Sleep(5 * time.Millisecond)
	}
	deadline = time.Now().Add(time.Second)
	for {
		rows := source.Registered()
		if len(rows) > 0 && rows[0].SlotsInUse == 0 {
			if !slices.Contains(rows[0].RunningTaskIDs, "assignment") {
				t.Fatalf("verification lease disappeared after provider exit: %+v", rows[0].RunningTaskIDs)
			}
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("native session slot not released during verification")
		}
		time.Sleep(5 * time.Millisecond)
	}
	source.Enqueue(bridge.TaskRequest{ID: "other-assignment", Provider: "fake", Prompt: "unrelated advice", Metadata: map[string]string{"workspace_id": "workspace", "task_id": "other"}})
	next, other := awaitCorrectionRequest(t, adapter, proc)
	if next.TaskID != "other-assignment" {
		t.Fatalf("unrelated assignment blocked: %+v", next)
	}
	source.Cancel("assignment", context.Canceled)
	other.EmitStdout([]byte("done"))
	for i := 0; i < 2; i++ {
		select {
		case res := <-reporter.results:
			if res.Status == agentbridge.ResultCancelled {
				return
			}
		case <-time.After(3 * time.Second):
			t.Fatal("verification held cancellation")
		}
	}
	t.Fatal("cancelled verification was accepted")
}
