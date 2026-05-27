package cursor

import (
	"context"
	"os"
	"os/exec"
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
	if _, err := exec.LookPath(DefaultExecutable); err != nil {
		t.Skipf("cursor-agent not on $PATH: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Two-stage gate (audit M-8 / integration-matrix.md §0).
	det, err := Detect(ctx, agentbridge.DetectEnv{})
	if err != nil {
		t.Fatalf("Detect: %v", err)
	}
	if !det.Available {
		t.Skipf("cursor Detect reported Available=false: %s", det.Reason)
	}
	if ok, reason := cursorAccountAvailable(ctx); !ok {
		t.Skip(reason)
	}
	profile := ProfileRootPrint
	if det.Metadata != nil && det.Metadata["profile"] != "" {
		profile = Profile(det.Metadata["profile"])
	}
	req := agentbridge.StartRequest{
		Prompt: `Respond with exactly "ok".`,
		Cwd:    t.TempDir(),
	}
	spawn, err := BuildStart(req, StartOptions{Profile: profile})
	if err != nil {
		t.Fatalf("BuildStart: %v", err)
	}

	sess, err := session.Start(ctx, session.Config{
		TaskID:    "integration-cursor",
		RuntimeID: "cursor",
		Adapter:   integrationAdapter{},
		Process:   processexec.New(),
		Spawn: process.Command{
			Executable: spawn.Executable,
			Args:       spawn.Args,
			Env:        spawn.Env,
			Dir:        spawn.Dir,
		},
		Request:     req,
		HardTimeout: 45 * time.Second,
	})
	if err != nil {
		t.Fatalf("Start: %v", err)
	}

	eventsDone := make(chan []agentbridge.Event, 1)
	go func() {
		var events []agentbridge.Event
		for ev := range sess.Events() {
			events = append(events, ev)
		}
		eventsDone <- events
	}()

	res := <-sess.Result()
	events := <-eventsDone
	if res.Status != agentbridge.ResultCompleted {
		if cursorAuthMissing(res, events) {
			t.Skip("cursor-agent authentication missing; run cursor-agent login or set CURSOR_API_KEY")
		}
		t.Fatalf("cursor integration did not complete: %+v", res)
	}
}

func cursorAuthMissing(res agentbridge.Result, events []agentbridge.Event) bool {
	var b strings.Builder
	b.WriteString(res.Error)
	b.WriteByte(' ')
	b.WriteString(res.Output)
	for _, ev := range events {
		b.WriteByte(' ')
		b.WriteString(ev.Text)
		b.WriteByte(' ')
		b.WriteString(ev.Err)
	}
	haystack := strings.ToLower(b.String())
	return strings.Contains(haystack, "authentication required") ||
		strings.Contains(haystack, "cursor_api_key")
}

func cursorAccountAvailable(ctx context.Context) (bool, string) {
	probeCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	cmd := exec.CommandContext(probeCtx, DefaultExecutable, "about")
	out, err := cmd.CombinedOutput()
	text := strings.ToLower(string(out))
	if strings.Contains(text, "not logged in") {
		return false, "cursor-agent account missing; run cursor-agent login or set CURSOR_API_KEY"
	}
	if err != nil {
		return true, ""
	}
	return true, ""
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
