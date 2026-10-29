package supervisor

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/executionproof"
	"github.com/teamswyg/riido-daemon/internal/execution"
)

func TestClaimProofVerificationReservedBeforeWorkerAndWait(t *testing.T) {
	root := t.TempDir()
	marker := filepath.Join(root, "riido-check-started")
	if err := os.WriteFile(filepath.Join(root, "result.txt"), []byte("changed"), 0o600); err != nil {
		t.Fatal(err)
	}
	scope := executionproof.NewScope(true, nil)
	ctx, cancel := context.WithCancel(executionproof.ContextWithScope(context.Background(), scope))
	defer cancel()
	task := &runningTask{ctx: ctx, execution: &execution.Prepared{Directory: root, Binding: execution.Binding{ChangedFiles: []string{"result.txt"}, Checks: [][]string{{os.Args[0], "-test.run=^TestExecutionBlockingCheck$", "--", marker}}}}}
	actor := &Actor{mailbox: make(chan envelope, 1), stoppedCh: make(chan struct{})}
	actor.beginExecutionVerification(task, agentbridge.Result{Status: agentbridge.ResultCompleted})
	deadline := time.Now().Add(3 * time.Second)
	for {
		if _, err := os.Stat(marker); err == nil {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("verification child did not start")
		}
		time.Sleep(5 * time.Millisecond)
	}
	if _, ok := scope.Tracker().Seal(); ok {
		t.Fatal("running host verification omitted from claim")
	}
	cancel()
	select {
	case msg := <-actor.mailbox:
		if msg.taskVerification == nil {
			t.Fatal("wrong notification")
		}
	case <-time.After(3 * time.Second):
		t.Fatal("verification did not finish after cancellation")
	}
	if _, ok := scope.Tracker().Seal(); ok {
		t.Fatal("late check Wait upgraded sealed unknown")
	}
}

func TestClaimProofVerificationFinishedBeforeNotification(t *testing.T) {
	scope := executionproof.NewScope(true, nil)
	ctx := executionproof.ContextWithScope(context.Background(), scope)
	task := &runningTask{ctx: ctx, execution: &execution.Prepared{Directory: t.TempDir()}}
	actor := &Actor{mailbox: make(chan envelope, 1), stoppedCh: make(chan struct{})}
	actor.beginExecutionVerification(task, agentbridge.Result{Status: agentbridge.ResultCompleted})
	select {
	case <-actor.mailbox:
	case <-time.After(time.Second):
		t.Fatal("verification notification missing")
	}
	proof, ok := scope.Tracker().Seal()
	if !ok || proof.StartedInvocations != 0 {
		t.Fatal("completed host worker remained pending or counted as provider")
	}
}
