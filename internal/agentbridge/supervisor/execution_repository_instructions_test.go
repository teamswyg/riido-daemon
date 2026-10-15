package supervisor

import (
	"context"
	"strings"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/execution"
)

func TestRepositoryOnlyInstructionsDoNotPromiseHostChecks(t *testing.T) {
	actor, status := historyActor(t.TempDir())
	actor.cfg.ExecutionBindings = []execution.Binding{{WorkspaceID: "workspace", TaskID: "task", Repository: correctionRepository(t)}}
	req := historyRequest("first", "")
	if _, err := actor.prepareWorkspace(context.Background(), status, req); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(req.Prompt, "host will require changed files and passing checks") {
		t.Fatal("repository-only binding promised unconfigured host checks")
	}
	if !strings.Contains(req.Prompt, "Host verification is not configured") {
		t.Fatal("missing honest host verification boundary")
	}
}
