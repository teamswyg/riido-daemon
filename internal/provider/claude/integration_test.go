package claude

import (
	"context"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/session"
	"github.com/teamswyg/riido-daemon/internal/process"
	"github.com/teamswyg/riido-daemon/internal/process/processexec"
)

// TestIntegration spawns the real Claude Code CLI and runs a trivial
// prompt. Skipped unless AGENTBRIDGE_INTEGRATION=1 is set AND the
// `claude` binary is on $PATH. This matches spec §10 Phase 6 / §6.X
// (Dev/Prod Parity).
func TestIntegration(t *testing.T) {
	if os.Getenv("AGENTBRIDGE_INTEGRATION") != "1" {
		t.Skip("AGENTBRIDGE_INTEGRATION not set")
	}
	if _, err := exec.LookPath(DefaultExecutable); err != nil {
		t.Skipf("claude not on $PATH: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Two-stage gate (audit M-8 / integration-matrix.md §0).
	det, err := Detect(ctx, agentbridge.DetectEnv{})
	if err != nil {
		t.Fatalf("Detect: %v", err)
	}
	if !det.Available {
		t.Skipf("claude Detect reported Available=false: %s", det.Reason)
	}

	spawn, err := BuildStart(agentbridge.StartRequest{
		Prompt: `Respond with exactly the single word "ok" and nothing else.`,
		Cwd:    t.TempDir(),
	}, StartOptions{PermissionMode: PermissionModeApproval})
	if err != nil {
		t.Fatalf("BuildStart: %v", err)
	}

	// Real claude requires the user-message stream-json frame on stdin
	// followed by EOF. Plumb the Claude ProtocolDriver so the session
	// actor writes the frame and closes stdin instead of leaving claude
	// blocked on read.
	driver, err := NewProtocolDriver(agentbridge.StartRequest{
		Prompt: `Respond with exactly the single word "ok" and nothing else.`,
		Cwd:    spawn.Dir,
	})
	if err != nil {
		t.Fatalf("NewProtocolDriver: %v", err)
	}

	sess, err := session.Start(ctx, session.Config{
		TaskID:    "integration-claude",
		RuntimeID: "claude",
		Adapter:   integrationAdapter{},
		Process:   processexec.New(),
		Spawn: process.Command{
			Executable: spawn.Executable,
			Args:       spawn.Args,
			Env:        spawn.Env,
			Dir:        spawn.Dir,
		},
		HardTimeout:    45 * time.Second,
		ProtocolDriver: driver,
	})
	if err != nil {
		t.Fatalf("Start: %v", err)
	}

	// Pump events to keep the channel from blocking the actor.
	go func() {
		for range sess.Events() {
		}
	}()

	res := <-sess.Result()
	if res.Status != agentbridge.ResultCompleted {
		t.Fatalf("claude integration did not complete: %+v", res)
	}
}

// integrationAdapter wraps the package-level helpers as an Adapter so
// session.Start can use it. We can't reach into the bridge package
// here without an import cycle, so we keep this duplicate small.
type integrationAdapter struct{}

func (integrationAdapter) Name() string { return Name }
func (integrationAdapter) Detect(_ context.Context, _ agentbridge.DetectEnv) (agentbridge.DetectResult, error) {
	return agentbridge.DetectResult{Available: true}, nil
}
func (integrationAdapter) BuildStart(req agentbridge.StartRequest) (agentbridge.StartCommand, error) {
	return BuildStart(req, StartOptions{PermissionMode: PermissionModeApproval})
}
func (integrationAdapter) NewParser() agentbridge.Parser { return NewParser() }
func (integrationAdapter) Translate(raw agentbridge.RawEvent) ([]agentbridge.Event, []agentbridge.Command, error) {
	return Translate(raw)
}
func (integrationAdapter) BlockedArgs() []string { return BlockedArgs() }

// We also need the integration adapter to opt into the ProtocolDriver
// path; without it claude -p sits on stdin and times out. But since
// the integration test uses session.Start directly (not RuntimeActor),
// we have to plumb the driver into session.Config.ProtocolDriver
// ourselves — see TestIntegration.
