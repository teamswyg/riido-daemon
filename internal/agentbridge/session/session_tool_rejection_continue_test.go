package session

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestSessionNativeToolRejectionContinuesOwnedRun(t *testing.T) {
	for _, scenario := range []string{"denied", "resolver error", "resolver absent"} {
		t.Run(scenario, func(t *testing.T) {
			adapter := continuingRejectionAdapter(t)
			started := startRecordingSession(t, "assignment-1", adapter, func(cfg *Config) {
				cfg.AutoApprove = func(agentbridge.ToolRef) bool {
					t.Error("native confirmation must never be auto approved")
					return true
				}
				if scenario != "resolver absent" {
					cfg.ToolApprovalResolver = resolverFunc(func(context.Context, string, agentbridge.ToolRef) (agentbridge.ToolApprovalResolution, error) {
						if scenario == "resolver error" {
							return agentbridge.ToolApprovalResolution{}, context.DeadlineExceeded
						}
						return agentbridge.ToolApprovalResolution{Reason: "explicit denial"}, nil
					})
				}
			})
			t.Cleanup(func() { started.sess.Cancel(nil) })
			started.running.EmitStdout([]byte("ASK"))
			assertRejectionProviderInput(t, started.running)
			emitDone(started.running)
			if res := waitResult(t, started.sess, time.Second); res.Status != agentbridge.ResultCompleted {
				t.Fatalf("denied tool aborted remaining work: %+v", res)
			}
			events := drainEvents(t, started.sess, time.Second)
			if !hasWarningText(events, "tool approval rejected; continuing without this tool") {
				t.Fatal("missing tool rejection warning")
			}
			if scenario == "resolver error" {
				found := false
				for _, event := range events {
					found = found || event.Err == "tool approval resolver failed: context deadline exceeded"
				}
				if !found {
					t.Fatal("resolver failure was replaced with generic unavailable reason")
				}
			}
		})
	}
}

func TestSessionNativeRejectionDeliveryFailureStillBlocks(t *testing.T) {
	for _, scenario := range []string{"build", "empty", "write"} {
		t.Run(scenario, func(t *testing.T) {
			adapter := continuingRejectionAdapter(t)
			if scenario != "write" {
				adapter.inputFn = func(cmd agentbridge.Command) ([]byte, error) {
					assertToolApprovalCommand(t, cmd, agentbridge.CommandRejectTool)
					if scenario == "build" {
						return nil, errors.New("native rejection could not be built")
					}
					return nil, nil
				}
			}
			started := startRecordingSession(t, "assignment-1", adapter, nil)
			if scenario == "write" {
				for range 64 {
					if err := started.running.WriteStdin([]byte("occupied")); err != nil {
						t.Fatal(err)
					}
				}
			}
			started.running.EmitStdout([]byte("ASK"))
			if res := waitResult(t, started.sess, time.Second); res.Status != agentbridge.ResultBlocked {
				t.Fatalf("unconfirmed rejection must remain blocked: %+v", res)
			}
			events := drainEvents(t, started.sess, time.Second)
			warning := map[string]string{"build": "provider approval command build failed", "empty": "provider rejection command has no input", "write": "provider approval command write failed"}[scenario]
			if !hasWarningText(events, warning) {
				t.Fatalf("missing rejection delivery failure: %s", warning)
			}
		})
	}
}

func continuingRejectionAdapter(t *testing.T) *recordingAdapter {
	t.Helper()
	adapter := approvalCommandAdapter(t, "mcp_confirmation", agentbridge.CommandRejectTool)
	translate := adapter.translateFn
	adapter.translateFn = func(raw agentbridge.RawEvent) ([]agentbridge.Event, []agentbridge.Command, error) {
		events, commands, err := translate(raw)
		if len(events) == 1 && events[0].Kind == agentbridge.EventToolApprovalNeeded {
			events[0].Tool.RequireHuman = true
			events[0].Tool.ContinueAfterRejection = true
		}
		return events, commands, err
	}
	return adapter
}
