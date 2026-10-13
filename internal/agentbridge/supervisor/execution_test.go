package supervisor

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/execution"
	"github.com/teamswyg/riido-daemon/internal/workdir"
)

func TestProviderSuccessRequiresExecutionEvidenceForBoundTask(t *testing.T) {
	repo := t.TempDir()
	if err := os.WriteFile(filepath.Join(repo, "result.txt"), []byte("old"), 0o600); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{{"init", "-q", repo}, {"-C", repo, "add", "."}, {"-C", repo, "-c", "user.name=test", "-c", "user.email=test@example.invalid", "commit", "-qm", "fixture"}} {
		if out, err := exec.Command("git", args...).CombinedOutput(); err != nil {
			t.Fatalf("%s: %v", out, err)
		}
	}
	prepared, err := execution.Prepare(context.Background(), execution.Binding{Repository: repo, ChangedFiles: []string{"result.txt"}, Checks: [][]string{{"git", "diff", "--check"}}}, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	actor := &Actor{}
	task := &runningTask{taskID: "task", execution: prepared, workspace: &workdir.Workspace{Workdir: prepared.Directory, Artifacts: t.TempDir()}}
	result := actor.recordTerminalResult(context.Background(), task, agentbridge.Result{Status: agentbridge.ResultCompleted, Output: "I will fix this next."})
	if result.Status != agentbridge.ResultFailed || !strings.Contains(result.Error, "not changed") {
		t.Fatalf("summary accepted as completed: %+v", result)
	}
	if _, err := os.Stat(filepath.Join(task.workspace.Artifacts, "verification.json")); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(prepared.Directory, "result.txt"), []byte("fixed\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	result = actor.recordTerminalResult(context.Background(), task, agentbridge.Result{Status: agentbridge.ResultCompleted, Output: "Fixed."})
	if result.Status != agentbridge.ResultCompleted || !strings.Contains(result.Output, "checks passed: 1") {
		t.Fatalf("verified work rejected: %+v", result)
	}
}
