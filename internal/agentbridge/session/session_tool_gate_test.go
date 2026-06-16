package session

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/process"
)

func TestSessionToolStartGateBlocksStartedTool(t *testing.T) {
	fake := process.NewFake()
	running := process.NewFakeRunning()
	fake.NextRunning = running
	parser := &recordingParser{}
	adapter := &recordingAdapter{
		name:   "fake",
		parser: parser,
		translateFn: func(raw agentbridge.RawEvent) ([]agentbridge.Event, []agentbridge.Command, error) {
			if string(raw.Bytes) == "START" {
				return []agentbridge.Event{{
					Kind: agentbridge.EventToolCallStarted,
					Tool: agentbridge.ToolRef{ID: "tool-1", Kind: "shell", Args: map[string]string{"command": "rm -rf .git"}},
				}}, nil, nil
			}
			return nil, nil, nil
		},
	}

	sess, err := Start(context.Background(), Config{
		TaskID:    "task-tool-start-block",
		RuntimeID: "rt-1",
		Adapter:   adapter,
		Process:   fake,
		Spawn:     process.Command{Executable: "fake"},
		ToolStartGate: func(tool agentbridge.ToolRef) agentbridge.ToolStartDecision {
			if tool.ID != "tool-1" {
				t.Fatalf("unexpected tool: %+v", tool)
			}
			return agentbridge.ToolStartDecision{Block: true, Code: "TOOL_USE_NOT_IN_POLICY_BUNDLE", Reason: "blocked in test"}
		},
	})
	if err != nil {
		t.Fatalf("Start: %v", err)
	}

	running.EmitStdout([]byte("START"))
	select {
	case <-running.KillRecv():
	case <-time.After(time.Second):
		t.Fatal("expected provider kill after tool start block")
	}
	res := waitResult(t, sess, 2*time.Second)
	if res.Status != agentbridge.ResultBlocked {
		t.Fatalf("result: %+v", res)
	}
	if res.Error != "TOOL_USE_NOT_IN_POLICY_BUNDLE: blocked in test" {
		t.Fatalf("block error: %q", res.Error)
	}
	events := drainEvents(t, sess, time.Second)
	var sawWarning bool
	for _, ev := range events {
		if ev.Kind == agentbridge.EventWarning && ev.Text == "tool use blocked by policy" {
			sawWarning = true
		}
	}
	if !sawWarning {
		t.Fatalf("missing policy warning in events: %+v", events)
	}
}

func TestSessionToolApprovalGateBlocksHeadlessApproval(t *testing.T) {
	fake := process.NewFake()
	running := process.NewFakeRunning()
	fake.NextRunning = running
	parser := &recordingParser{}
	adapter := &recordingAdapter{
		name:   "fake",
		parser: parser,
		translateFn: func(raw agentbridge.RawEvent) ([]agentbridge.Event, []agentbridge.Command, error) {
			if string(raw.Bytes) == "APPROVAL" {
				return []agentbridge.Event{{
					Kind: agentbridge.EventToolApprovalNeeded,
					Tool: agentbridge.ToolRef{ID: "approval-1", Kind: "patch_apply"},
				}}, nil, nil
			}
			return nil, nil, nil
		},
	}

	sess, err := Start(context.Background(), Config{
		TaskID:    "task-tool-approval-block",
		RuntimeID: "rt-1",
		Adapter:   adapter,
		Process:   fake,
		Spawn:     process.Command{Executable: "fake"},
		ToolApprovalGate: func(tool agentbridge.ToolRef) agentbridge.ToolStartDecision {
			if tool.ID != "approval-1" {
				t.Fatalf("unexpected tool: %+v", tool)
			}
			return agentbridge.ToolStartDecision{Block: true, Code: "TOOL_USE_NOT_IN_POLICY_BUNDLE", Reason: "no headless approval path"}
		},
	})
	if err != nil {
		t.Fatalf("Start: %v", err)
	}

	running.EmitStdout([]byte("APPROVAL"))
	select {
	case <-running.KillRecv():
	case <-time.After(time.Second):
		t.Fatal("expected provider kill after headless approval block")
	}
	res := waitResult(t, sess, 2*time.Second)
	if res.Status != agentbridge.ResultBlocked {
		t.Fatalf("result: %+v", res)
	}
	if res.Error != "TOOL_USE_NOT_IN_POLICY_BUNDLE: no headless approval path" {
		t.Fatalf("block error: %q", res.Error)
	}
	events := drainEvents(t, sess, time.Second)
	var sawWarning bool
	for _, ev := range events {
		if ev.Kind == agentbridge.EventWarning && ev.Text == "tool approval unavailable in headless run" {
			sawWarning = true
		}
	}
	if !sawWarning {
		t.Fatalf("missing headless approval warning in events: %+v", events)
	}
}

func TestSessionRemovesTempFilesAfterProcessExit(t *testing.T) {
	tempFile, err := os.CreateTemp(t.TempDir(), "mcp-*.json")
	if err != nil {
		t.Fatalf("CreateTemp: %v", err)
	}
	if err := tempFile.Close(); err != nil {
		t.Fatalf("close temp file: %v", err)
	}

	fake := process.NewFake()
	fake.NextRunning = process.NewFakeRunning()
	adapter := &recordingAdapter{
		name:   "fake",
		parser: &recordingParser{},
		translateFn: func(raw agentbridge.RawEvent) ([]agentbridge.Event, []agentbridge.Command, error) {
			if string(raw.Bytes) == "DONE" {
				return []agentbridge.Event{{Kind: agentbridge.EventResult, Result: agentbridge.Result{Status: agentbridge.ResultCompleted}}}, nil, nil
			}
			return nil, nil, nil
		},
	}

	sess, err := Start(context.Background(), Config{
		TaskID:    "task-tempfile-exit",
		RuntimeID: "rt-1",
		Adapter:   adapter,
		Process:   fake,
		Spawn:     process.Command{Executable: "fake"},
		TempFiles: []string{tempFile.Name(), tempFile.Name()},
	})
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	go func() {
		sess.runningForTest().EmitStdout([]byte("DONE"))
		sess.runningForTest().EmitExit(0, nil)
	}()

	res := waitResult(t, sess, 2*time.Second)
	if res.Status != agentbridge.ResultCompleted {
		t.Fatalf("result: %+v", res)
	}
	if _, err := os.Stat(tempFile.Name()); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("temp file should be removed after process exit, stat err=%v", err)
	}
	_ = drainEvents(t, sess, time.Second)
}

func TestSessionRemovesTempFilesAfterCancellation(t *testing.T) {
	tempFile, err := os.CreateTemp(t.TempDir(), "mcp-*.json")
	if err != nil {
		t.Fatalf("CreateTemp: %v", err)
	}
	if err := tempFile.Close(); err != nil {
		t.Fatalf("close temp file: %v", err)
	}

	fake := process.NewFake()
	fake.NextRunning = process.NewFakeRunning()
	adapter := &recordingAdapter{
		name:        "fake",
		parser:      &recordingParser{},
		translateFn: func(agentbridge.RawEvent) ([]agentbridge.Event, []agentbridge.Command, error) { return nil, nil, nil },
	}
	sess, err := Start(context.Background(), Config{
		TaskID:    "task-tempfile-cancel",
		RuntimeID: "rt-1",
		Adapter:   adapter,
		Process:   fake,
		Spawn:     process.Command{Executable: "fake"},
		TempFiles: []string{tempFile.Name()},
	})
	if err != nil {
		t.Fatalf("Start: %v", err)
	}

	sess.Cancel(nil)
	res := waitResult(t, sess, time.Second)
	if res.Status != agentbridge.ResultCancelled {
		t.Fatalf("result: %+v", res)
	}
	if _, err := os.Stat(tempFile.Name()); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("temp file should be removed after cancellation, stat err=%v", err)
	}
	_ = drainEvents(t, sess, time.Second)
}
