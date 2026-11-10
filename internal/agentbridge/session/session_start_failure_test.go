package session

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/process"
)

type startupDiagnosticAdapter struct{ trackingAdapter }

func (*startupDiagnosticAdapter) Translate(raw agentbridge.RawEvent) ([]agentbridge.Event, []agentbridge.Command, error) {
	return []agentbridge.Event{
		{Kind: agentbridge.EventError, Err: string(raw.Bytes)},
		{Kind: agentbridge.EventToolApprovalNeeded},
		{Kind: agentbridge.EventResult, Result: agentbridge.Result{Status: agentbridge.ResultCompleted}},
	}, []agentbridge.Command{{Kind: agentbridge.CommandCancelProvider}}, nil
}

func TestStartupDiagnosticIsBoundedAndNeverResumesDriver(t *testing.T) {
	running := process.NewFakeRunning()
	running.EmitStdout([]byte(strings.Repeat("x", startupDiagnosticBytes*2)))
	driver := &fakeDriver{onStart: func(context.Context, ProtocolIO) error { return errors.New("pipe ended") }}
	approved := false
	sess, err := Start(context.Background(), Config{
		Adapter: &startupDiagnosticAdapter{}, Process: &process.Fake{NextRunning: running},
		Spawn: process.Command{Executable: "fake"}, ProtocolDriver: driver,
		AutoApprove: func(agentbridge.ToolRef) bool { approved = true; return true },
	})
	if err != nil {
		t.Fatal(err)
	}
	for range sess.Events() {
	}
	result := <-sess.Result()
	if result.Status != agentbridge.ResultFailed || len(result.Error) > 2200 || !strings.Contains(result.Error, "pipe ended") {
		t.Fatalf("unbounded or promoted startup: %+v", result)
	}
	if driver.rawCalls != 0 || approved {
		t.Fatal("startup diagnostics resumed protocol processing or approved a tool")
	}
}

func TestStartupDiagnosticPreservesTermination(t *testing.T) {
	for _, tc := range []struct {
		name       string
		hard, idle time.Duration
		want       agentbridge.ResultStatus
	}{
		{"cancel", 0, 0, agentbridge.ResultCancelled},
		{"hard", 30 * time.Millisecond, 0, agentbridge.ResultTimeout},
		{"idle", 0, 30 * time.Millisecond, agentbridge.ResultTimeout},
	} {
		t.Run(tc.name, func(t *testing.T) {
			entered := make(chan struct{})
			driver := &fakeDriver{onStart: func(context.Context, ProtocolIO) error { close(entered); return errors.New("pipe ended") }}
			sess, err := Start(context.Background(), Config{
				Adapter: &trackingAdapter{}, Process: process.NewFake(),
				Spawn: process.Command{Executable: "fake"}, ProtocolDriver: driver,
				HardTimeout: tc.hard, SemanticIdle: tc.idle,
			})
			if err != nil {
				t.Fatal(err)
			}
			<-entered
			if tc.name == "cancel" {
				sess.Cancel(nil)
			}
			select {
			case result := <-sess.Result():
				if result.Status != tc.want {
					t.Fatalf("status=%s want=%s", result.Status, tc.want)
				}
			case <-time.After(time.Second):
				t.Fatal("startup diagnostic blocked termination")
			}
		})
	}
}
