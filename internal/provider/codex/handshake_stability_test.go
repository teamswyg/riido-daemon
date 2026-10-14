package codex

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestCodexThreadSetupPreservesWorkspaceAndInstructions(t *testing.T) {
	for _, resume := range []string{"", "existing-thread"} {
		d, _ := NewProtocolDriver(agentbridge.StartRequest{Cwd: "/authorized/repository", SystemPrompt: "Perform and verify this task", ResumeSessionID: resume})
		io := newRecordingIO()
		_ = d.OnStart(context.Background(), io)
		_ = io.next(t, time.Second)
		_, _, err := d.OnRaw(context.Background(), makeResponse(1, nil), io)
		if err != nil {
			t.Fatal(err)
		}
		_ = io.next(t, time.Second)
		var frame map[string]any
		if err := json.Unmarshal(io.next(t, time.Second), &frame); err != nil {
			t.Fatal(err)
		}
		params := mapField(frame, "params")
		if params["cwd"] != "/authorized/repository" || params["developerInstructions"] != "Perform and verify this task" || params["sandbox"] != FullAccessSandboxMode {
			t.Fatalf("thread setup: %+v", frame)
		}
		if resume != "" && (frame["method"] != "thread/resume" || params["threadId"] != resume) {
			t.Fatalf("resume: %+v", frame)
		}
	}
}

func TestCodexInvalidThreadIdentityFailsBeforeTurnStart(t *testing.T) {
	for _, id := range []string{"", "different-thread"} {
		d, _ := NewProtocolDriver(agentbridge.StartRequest{ResumeSessionID: "requested-thread"})
		io := newRecordingIO()
		_ = d.OnStart(context.Background(), io)
		_ = io.next(t, time.Second)
		_, _, _ = d.OnRaw(context.Background(), makeResponse(1, nil), io)
		_ = io.next(t, time.Second)
		_ = io.next(t, time.Second)
		events, _, err := d.OnRaw(context.Background(), makeResponse(2, map[string]any{"thread": map[string]any{"id": id}}), io)
		if err != nil {
			t.Fatal(err)
		}
		if len(events) == 0 || events[len(events)-1].Result.Status != agentbridge.ResultFailed {
			t.Fatalf("invalid identity: %+v", events)
		}
		if len(io.writes) != 0 {
			t.Fatal("turn/start sent for invalid identity")
		}
	}
}
