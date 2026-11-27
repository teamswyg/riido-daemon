package session

import (
	"context"
	"errors"
	"testing"
	"testing/synctest"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestSessionResolverDelayedDecisionRemainsExplicitAndFailClosed(t *testing.T) {
	for _, tc := range []struct {
		name       string
		resolution agentbridge.ToolApprovalResolution
		err        error
		command    agentbridge.CommandKind
		status     agentbridge.ResultStatus
		reason     string
	}{
		{"approve", agentbridge.ToolApprovalResolution{Approved: true}, nil, agentbridge.CommandApproveTool, agentbridge.ResultCompleted, ""},
		{"deny", agentbridge.ToolApprovalResolution{Reason: "explicit denial"}, nil, agentbridge.CommandRejectTool, agentbridge.ResultBlocked, "explicit denial"},
		{"resolver error", agentbridge.ToolApprovalResolution{}, errors.New("resolver unavailable"), agentbridge.CommandRejectTool, agentbridge.ResultBlocked, "tool approval resolver failed: resolver unavailable"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				requested := make(chan agentbridge.ToolRef, 1)
				decision := make(chan struct{})
				adapter := approvalCommandAdapter(t, "mcp_confirmation", tc.command)
				translate := adapter.translateFn
				adapter.translateFn = func(raw agentbridge.RawEvent) ([]agentbridge.Event, []agentbridge.Command, error) {
					events, commands, err := translate(raw)
					if len(events) == 1 && events[0].Kind == agentbridge.EventToolApprovalNeeded {
						events[0].Tool.RequireHuman = true
					}
					return events, commands, err
				}
				started := startRecordingSession(t, "assignment-1", adapter, func(cfg *Config) {
					cfg.SemanticIdle = 100 * time.Millisecond
					cfg.AutoApprove = func(agentbridge.ToolRef) bool { return true }
					cfg.ToolApprovalResolver = resolverFunc(func(ctx context.Context, executionID string, tool agentbridge.ToolRef) (agentbridge.ToolApprovalResolution, error) {
						if executionID != "assignment-1" || tool.ID != "tool-1" || tool.ProviderRequestID != "req-1" || !tool.RequireHuman {
							t.Errorf("resolver identity: execution=%s tool=%+v", executionID, tool)
						}
						requested <- tool
						select {
						case <-decision:
							return tc.resolution, tc.err
						case <-ctx.Done():
							return agentbridge.ToolApprovalResolution{}, ctx.Err()
						}
					})
				})
				t.Cleanup(func() { started.sess.Cancel(nil) })
				started.running.EmitStdout([]byte("ASK"))
				<-requested
				time.Sleep(200 * time.Millisecond)
				synctest.Wait()
				select {
				case res := <-started.sess.Result():
					t.Fatalf("ended before explicit decision: %+v", res)
				default:
				}
				select {
				case frame := <-started.running.StdinRecv():
					t.Fatalf("sent provider input before explicit decision: %q", frame)
				default:
				}
				close(decision)
				frame := <-started.running.StdinRecv()
				if tc.resolution.Approved && string(frame) != "approve:req-1\n" || !tc.resolution.Approved && string(frame) != "reject:req-1\n" {
					t.Fatalf("decision frame: %q", frame)
				}
				if tc.resolution.Approved {
					emitDone(started.running)
				}
				res := waitResult(t, started.sess, time.Second)
				if res.Status != tc.status || res.Error != tc.reason {
					t.Fatalf("result: %+v", res)
				}
				_ = drainEvents(t, started.sess, time.Second)
			})
		})
	}
}
