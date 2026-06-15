package session

import (
	"context"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/process"
)

// Normal completion: stdout chunk → translated TextDelta → Result.
func TestSessionNormalCompletion(t *testing.T) {
	fake := process.NewFake()
	fake.NextRunning = process.NewFakeRunning()
	parser := &recordingParser{}
	adapter := &recordingAdapter{
		name:   "fake",
		parser: parser,
		translateFn: func(raw agentbridge.RawEvent) ([]agentbridge.Event, []agentbridge.Command, error) {
			if raw.Type != "chunk" {
				return nil, nil, nil
			}
			if string(raw.Bytes) == "DONE" {
				return []agentbridge.Event{{Kind: agentbridge.EventResult, Result: agentbridge.Result{Status: agentbridge.ResultCompleted, Output: "hello"}}}, nil, nil
			}
			return []agentbridge.Event{
				{Kind: agentbridge.EventLifecycle, Phase: agentbridge.StateRunning},
				{Kind: agentbridge.EventTextDelta, Text: string(raw.Bytes)},
			}, nil, nil
		},
	}

	cfg := Config{
		TaskID:    "task-1",
		RuntimeID: "rt-1",
		Adapter:   adapter,
		Process:   fake,
		Spawn:     process.Command{Executable: "fake"},
	}
	sess, err := Start(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	running := sess.runningForTest()

	// Drive output: a normal chunk, then DONE, then exit.
	go func() {
		running.EmitStdout([]byte("hello"))
		running.EmitStdout([]byte("DONE"))
		running.EmitExit(0, nil)
	}()

	res := waitResult(t, sess, 2*time.Second)
	if res.Status != agentbridge.ResultCompleted {
		t.Fatalf("status: %s", res.Status)
	}
	if res.Output != "hello" {
		t.Fatalf("output: %q", res.Output)
	}

	// After Result, Events channel must be closed.
	events := drainEvents(t, sess, time.Second)
	sawText := false
	for _, ev := range events {
		if ev.Kind == agentbridge.EventTextDelta && ev.Text == "hello" {
			sawText = true
		}
	}
	if !sawText {
		t.Fatalf("expected to see TextDelta in event stream, got %+v", events)
	}
}

func TestSessionExtractsRiidoTelemetryAboveAdapters(t *testing.T) {
	fake := process.NewFake()
	fake.NextRunning = process.NewFakeRunning()
	adapter := &recordingAdapter{
		name:   "fake",
		parser: &recordingParser{},
		translateFn: func(raw agentbridge.RawEvent) ([]agentbridge.Event, []agentbridge.Command, error) {
			if string(raw.Bytes) == "DONE" {
				return []agentbridge.Event{{Kind: agentbridge.EventResult, Result: agentbridge.Result{Status: agentbridge.ResultCompleted}}}, nil, nil
			}
			return []agentbridge.Event{{Kind: agentbridge.EventTextDelta, Text: string(raw.Bytes)}}, nil, nil
		},
	}
	sess, err := Start(context.Background(), Config{
		TaskID:    "task-telemetry",
		RuntimeID: "rt-1",
		Adapter:   adapter,
		Process:   fake,
		Spawn:     process.Command{Executable: "fake"},
	})
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	running := sess.runningForTest()
	go func() {
		running.EmitStdout([]byte("<riido_log>프로젝트 go"))
		running.EmitStdout([]byte(".mod 작성중<end>"))
		running.EmitStdout([]byte("DONE"))
		running.EmitExit(0, nil)
	}()
	res := waitResult(t, sess, 2*time.Second)
	if res.Status != agentbridge.ResultCompleted {
		t.Fatalf("result: %+v", res)
	}
	events := drainEvents(t, sess, time.Second)
	for _, ev := range events {
		if ev.Kind == agentbridge.EventProgress && ev.Text == "프로젝트 go.mod 작성중" {
			return
		}
	}
	t.Fatalf("missing progress event in %+v", events)
}

func TestSessionAutoApprovalWritesProviderInput(t *testing.T) {
	fake := process.NewFake()
	running := process.NewFakeRunning()
	fake.NextRunning = running
	parser := &recordingParser{}
	adapter := &recordingAdapter{
		name:   "fake",
		parser: parser,
		translateFn: func(raw agentbridge.RawEvent) ([]agentbridge.Event, []agentbridge.Command, error) {
			switch string(raw.Bytes) {
			case "ASK":
				return []agentbridge.Event{{
					Kind: agentbridge.EventToolApprovalNeeded,
					Tool: agentbridge.ToolRef{ID: "tool-1", Kind: "read", ProviderRequestID: "req-1"},
				}}, nil, nil
			case "DONE":
				return []agentbridge.Event{{Kind: agentbridge.EventResult, Result: agentbridge.Result{Status: agentbridge.ResultCompleted}}}, nil, nil
			default:
				return nil, nil, nil
			}
		},
		inputFn: func(cmd agentbridge.Command) ([]byte, error) {
			if cmd.Kind != agentbridge.CommandApproveTool || cmd.ToolID != "tool-1" || cmd.ProviderRequestID != "req-1" {
				t.Fatalf("unexpected provider input command: %+v", cmd)
			}
			return []byte("approve:req-1\n"), nil
		},
	}

	sess, err := Start(context.Background(), Config{
		TaskID:      "task-approval",
		RuntimeID:   "rt-1",
		Adapter:     adapter,
		Process:     fake,
		Spawn:       process.Command{Executable: "fake"},
		AutoApprove: func(tool agentbridge.ToolRef) bool { return tool.Kind == "read" },
	})
	if err != nil {
		t.Fatalf("Start: %v", err)
	}

	running.EmitStdout([]byte("ASK"))
	select {
	case got := <-running.StdinRecv():
		if string(got) != "approve:req-1\n" {
			t.Fatalf("provider input = %q", got)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("provider input was not written")
	}

	go func() {
		running.EmitStdout([]byte("DONE"))
		running.EmitExit(0, nil)
	}()
	res := waitResult(t, sess, 2*time.Second)
	if res.Status != agentbridge.ResultCompleted {
		t.Fatalf("result: %+v", res)
	}
	_ = drainEvents(t, sess, time.Second)
}
