package codex

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

func TestIntegration(t *testing.T) {
	if os.Getenv("AGENTBRIDGE_INTEGRATION") != "1" {
		t.Skip("AGENTBRIDGE_INTEGRATION not set")
	}
	if _, err := exec.LookPath(DefaultExecutable); err != nil {
		t.Skipf("codex not on $PATH: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Two-stage gate (audit M-8 / integration-matrix.md §0).
	det, err := Detect(ctx, agentbridge.DetectEnv{})
	if err != nil {
		t.Fatalf("Detect: %v", err)
	}
	if !det.Available {
		t.Skipf("codex Detect reported Available=false: %s", det.Reason)
	}

	req := agentbridge.StartRequest{
		Prompt: `Respond with exactly "ok".`,
		Cwd:    t.TempDir(),
	}
	spawn, err := BuildStart(req, StartOptions{CodexHome: t.TempDir()})
	if err != nil {
		t.Fatalf("BuildStart: %v", err)
	}
	driver, err := NewProtocolDriver(req)
	if err != nil {
		t.Fatalf("NewProtocolDriver: %v", err)
	}

	sess, err := session.Start(ctx, session.Config{
		TaskID:    "integration-codex",
		RuntimeID: "codex",
		Adapter:   integrationAdapter{},
		Process:   processexec.New(),
		Spawn: process.Command{
			Executable: spawn.Executable,
			Args:       spawn.Args,
			Env:        spawn.Env,
			Dir:        spawn.Dir,
		},
		Request:        req,
		HardTimeout:    45 * time.Second,
		ProtocolDriver: driver,
	})
	if err != nil {
		t.Fatalf("Start: %v", err)
	}

	go func() {
		for range sess.Events() {
		}
	}()

	res := <-sess.Result()
	if res.Status != agentbridge.ResultCompleted {
		t.Fatalf("codex integration did not complete: %+v", res)
	}
}

type integrationAdapter struct{}

func (integrationAdapter) Name() string { return Name }
func (integrationAdapter) Detect(_ context.Context, _ agentbridge.DetectEnv) (agentbridge.DetectResult, error) {
	return agentbridge.DetectResult{Available: true}, nil
}
func (integrationAdapter) BuildStart(req agentbridge.StartRequest) (agentbridge.StartCommand, error) {
	return BuildStart(req, StartOptions{})
}
func (integrationAdapter) NewParser() agentbridge.Parser { return NewParser() }
func (integrationAdapter) Translate(raw agentbridge.RawEvent) ([]agentbridge.Event, []agentbridge.Command, error) {
	return Translate(raw)
}
func (integrationAdapter) BlockedArgs() []string { return BlockedArgs() }
