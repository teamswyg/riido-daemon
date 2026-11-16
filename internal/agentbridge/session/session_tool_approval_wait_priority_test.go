package session

import (
	"context"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/process"
)

func TestSessionResolvedApprovalHonorsReadyTerminationBeforeProviderInput(t *testing.T) {
	for _, tc := range []struct {
		name   string
		status agentbridge.ResultStatus
		reason string
	}{
		{"context", agentbridge.ResultCancelled, "context canceled"},
		{"session", agentbridge.ResultCancelled, "context canceled"},
		{"hard timeout", agentbridge.ResultTimeout, "hard timeout"},
		{"provider exit", agentbridge.ResultAborted, "provider disconnected while waiting for tool approval"},
		{"terminal request", agentbridge.ResultCancelled, "run stopped"},
		{"already terminal", agentbridge.ResultCancelled, "already cancelled"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			proc := process.NewFakeRunning()
			sess := &Session{events: make(chan agentbridge.Event, 10), cancel: make(chan cancelRequest, 1)}
			r := newSessionRunner(ctx, Config{Now: time.Now, Adapter: approvalAdapter(t, "patch_apply")}, sess, proc)
			defer r.stopTimers()
			switch tc.name {
			case "context":
				cancel()
			case "session":
				sess.cancel <- cancelRequest{ctx: ctx, cause: context.Canceled}
			case "hard timeout":
				ready := make(chan time.Time, 1)
				ready <- time.Now()
				r.hardC = ready
			case "already terminal":
				r.state.Terminal = true
				r.state.Result = agentbridge.Result{Status: tc.status, Error: tc.reason}
			case "provider exit":
				exited := make(chan process.ExitStatus, 1)
				exited <- process.ExitStatus{Code: 1}
				r.exitedCh = exited
			case "terminal request":
				sess.terminal = make(chan terminalRequest, 1)
				sess.terminal <- terminalRequest{ctx: ctx, result: agentbridge.Result{Status: tc.status, Error: tc.reason}}
			}
			tool := agentbridge.ToolRef{ID: "tool-1", Kind: "patch_apply", ProviderRequestID: "req-1"}
			// This is the sender boundary after select has chosen a ready resolver result.
			r.executeResolvedToolApproval(tool, agentbridge.ToolApprovalResolution{Approved: true})
			select {
			case frame := <-proc.StdinRecv():
				t.Fatalf("sent approval despite ready termination: %q", frame)
			default:
			}
			if !r.state.Terminal || r.state.Result.Status != tc.status || r.state.Result.Error != tc.reason {
				t.Fatalf("termination: %+v", r.state.Result)
			}
		})
	}
}
