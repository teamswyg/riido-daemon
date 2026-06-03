package openclaw

import (
	"context"
	"os"
	"strconv"
	"strings"
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

	ctx, cancel := context.WithTimeout(context.Background(), 210*time.Second)
	defer cancel()

	// Two-stage gate (audit M-8): Detect must also report Available
	// before we attempt a real prompt. This handles the "binary present
	// but unusable" case (e.g. Node version too old). See
	// docs/30-architecture/integration-matrix.md §0.
	det, err := Detect(ctx, agentbridge.DetectEnv{})
	if err != nil {
		t.Fatalf("Detect: %v", err)
	}
	if !det.Available {
		t.Skipf("openclaw Detect reported Available=false: %s", det.Reason)
	}

	sessionID := "integration-openclaw-" + strconv.FormatInt(time.Now().UnixNano(), 36)
	spawn, err := BuildStart(agentbridge.StartRequest{
		Prompt:     `Respond with exactly "ok".`,
		Cwd:        t.TempDir(),
		Executable: det.Executable,
	}, StartOptions{SessionID: sessionID})
	if err != nil {
		t.Fatalf("BuildStart: %v", err)
	}

	sess, err := session.Start(ctx, session.Config{
		TaskID:    "integration-openclaw",
		RuntimeID: "openclaw",
		Adapter:   integrationAdapter{sessionID: sessionID},
		Process:   processexec.New(),
		Spawn: process.Command{
			Executable: spawn.Executable,
			Args:       spawn.Args,
			Env:        spawn.Env,
			Dir:        spawn.Dir,
		},
		HardTimeout: 180 * time.Second,
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
		t.Fatalf("openclaw integration did not complete: %+v", res)
	}
	if !strings.Contains(strings.ToLower(res.Output), "ok") {
		t.Fatalf("openclaw integration output = %q", res.Output)
	}
}

type integrationAdapter struct {
	sessionID string
}

func (integrationAdapter) Name() string { return Name }
func (integrationAdapter) Detect(_ context.Context, _ agentbridge.DetectEnv) (agentbridge.DetectResult, error) {
	return agentbridge.DetectResult{Available: true}, nil
}
func (a integrationAdapter) BuildStart(req agentbridge.StartRequest) (agentbridge.StartCommand, error) {
	return BuildStart(req, StartOptions{SessionID: a.sessionID})
}
func (integrationAdapter) NewParser() agentbridge.Parser { return NewParser() }
func (integrationAdapter) Translate(raw agentbridge.RawEvent) ([]agentbridge.Event, []agentbridge.Command, error) {
	return Translate(raw)
}
func (integrationAdapter) BlockedArgs() []string { return BlockedArgs() }
