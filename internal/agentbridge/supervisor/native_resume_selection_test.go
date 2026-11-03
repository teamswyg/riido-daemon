package supervisor

import (
	"context"
	"errors"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/runtimeactor"
	"github.com/teamswyg/riido-daemon/internal/workdir"
)

func TestNativeResumeSelectionChecksScopeBeforeSearchingCLI(t *testing.T) {
	a, status := historyActor(t.TempDir())
	status.Capabilities[0].Provider = "codex"
	req := historyRequest("first", "")
	req.Provider = "codex"
	record, err := a.nativeSessionRecord(req, status.Capabilities[0], "session")
	if err != nil {
		t.Fatal(err)
	}
	if err := a.cfg.Workdir.(workdir.NativeSessions).PinNativeSession(record); err != nil {
		t.Fatal(err)
	}
	req = historyRequest("retry", "session")
	req.Provider = "codex"
	req.Metadata["task_id"] = "other-task"
	status.Capabilities[0].CapabilityFingerprint = "different-cli"
	// An uninitialized runtime would fail if the executable search was reached.
	if _, err := a.selectNativeResume(context.Background(), &runtimeactor.Actor{}, &status, req); !errors.Is(err, errNativeSessionBinding) {
		t.Fatalf("scope check=%v", err)
	}
}
