package runtimeactor

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/teamswyg/riido-contracts/hostintegration"
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
)

func TestUnknownAuthProbeCannotConcealBinaryChange(t *testing.T) {
	now := time.Date(2026, 10, 14, 0, 0, 0, 0, time.UTC)
	executable := filepath.Join(t.TempDir(), "claude")
	if err := os.WriteFile(executable, []byte("original provider"), 0o600); err != nil {
		t.Fatal(err)
	}
	verified := claudeCapabilityDetectResult(executable)
	adapter := &mutableDetectAdapter{stubAdapter: stubAdapter{name: "claude", detected: verified}}
	actor, proc := startActor(t, Config{Adapters: []agentbridge.Adapter{adapter}, CapabilityRefreshEvery: time.Second, Now: func() time.Time { return now }})
	handle, err := actor.Submit(context.Background(), bridge.TaskRequest{ID: "pinned", Provider: "claude"})
	if err != nil {
		t.Fatal(err)
	}
	running := waitForRunning(t, proc, 0, time.Second)
	if err := os.WriteFile(executable, []byte("replacement provider"), 0o600); err != nil {
		t.Fatal(err)
	}
	adapter.setDetected(agentbridge.DetectResult{Executable: executable, HealthStatus: hostintegration.ProviderHealthUnknown, DiagnosticCode: hostintegration.ProviderDiagnosticAuthProbeFailed})
	now = now.Add(2 * time.Second)
	_, _ = actor.Status(context.Background())
	expectFakeProcessKill(t, running, "unknown auth concealed actual binary replacement")
	if res := expectTaskResult(t, handle.Result(), "missing binary drift result"); res.Status != agentbridge.ResultBlocked {
		t.Fatalf("drift=%+v", res)
	}
}
