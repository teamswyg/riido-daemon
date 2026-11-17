package codex

import (
	"context"
	"testing"
	"testing/synctest"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/session"
)

type idleApprovalResolver func(context.Context, string, agentbridge.ToolRef) (agentbridge.ToolApprovalResolution, error)

func (f idleApprovalResolver) ResolveToolApproval(ctx context.Context, id string, tool agentbridge.ToolRef) (agentbridge.ToolApprovalResolution, error) {
	return f(ctx, id, tool)
}

func TestOwnedApprovalDuringIdleReadStillWaitsForUser(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		requested, decision := make(chan struct{}), make(chan struct{})
		sess, child := startOwnedSessionWithConfig(t, waitIdle, 0, func(cfg *session.Config) {
			cfg.ToolApprovalResolver = idleApprovalResolver(func(ctx context.Context, _ string, _ agentbridge.ToolRef) (agentbridge.ToolApprovalResolution, error) {
				close(requested)
				select {
				case <-decision:
					return agentbridge.ToolApprovalResolution{Approved: true}, nil
				case <-ctx.Done():
					return agentbridge.ToolApprovalResolution{}, ctx.Err()
				}
			})
		})
		receiveIdleRead(t, child)
		raw := makeServerRequest(99, "item/commandExecution/requestApproval", map[string]any{
			"threadId": "parent", "turnId": "parent-turn", "itemId": "command", "command": "pwd",
		})
		child.EmitStdout(append(raw.Bytes, '\n'))
		select {
		case <-requested:
		case <-time.After(time.Second):
			t.Fatal("approval was lost during recovery")
		}
		time.Sleep(24 * time.Hour)
		select {
		case result := <-sess.Result():
			t.Fatalf("approval timed out: %+v", result)
		default:
		}
		close(decision)
		select {
		case <-child.StdinRecv():
		case <-time.After(time.Second):
			t.Fatal("approval was not sent")
		}
		emitIdleCompletion(t, child)
		if result := <-sess.Result(); result.Status != agentbridge.ResultCompleted {
			t.Fatalf("result = %+v", result)
		}
	})
}
