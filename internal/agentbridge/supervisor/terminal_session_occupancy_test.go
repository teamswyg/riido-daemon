package supervisor

import (
	"context"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/executionproof"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/runtimeactor"
	"github.com/teamswyg/riido-daemon/internal/process"
)

type terminalCapacitySource struct {
	*controlplane.MemorySource
	available bool
}

func (s *terminalCapacitySource) ClaimTask(ctx context.Context, _ string) (*bridge.TaskRequest, error) {
	s.available = controlplane.ClaimCapacityAvailable(ctx)
	return nil, nil
}

func TestTerminalReportRetryDoesNotOccupyQuiescentSession(t *testing.T) {
	for _, observation := range []string{"exited", "not-started", "running", "unknown", "uncovered", "pending-host-work"} {
		t.Run(observation, func(t *testing.T) {
			scope := executionproof.NewScope(observation != "uncovered", nil)
			if observation != "not-started" {
				attempt, err := scope.Tracker().BeginInvocation()
				if err != nil {
					t.Fatal(err)
				}
				attempt.Started()
				if observation != "running" {
					attempt.Exited(observation != "unknown")
				}
			}
			if observation == "pending-host-work" {
				work, err := scope.Tracker().BeginHostWork()
				if err != nil {
					t.Fatal(err)
				}
				defer work.Finished()
			}
			ctx, cancel := context.WithCancel(executionproof.ContextWithScope(t.Context(), scope))
			defer cancel()
			reporter := newTerminalRetryReporter(2)
			source := &terminalCapacitySource{MemorySource: controlplane.NewMemorySource()}
			actor := &Actor{cfg: Config{Reporter: reporter, Source: source}}
			runtime := startRuntime(t, &process.Fake{})
			task := &runningTask{taskID: "old", provider: "codex", nativeSessionID: "same-session", runtimeID: "rt-local", ctx: ctx, cancel: cancel}
			inFlight := map[string]*runningTask{task.taskID: task}
			if actor.finishTaskWithResult(t.Context(), inFlight, task, agentbridge.Result{Status: agentbridge.ResultCompleted, SessionID: "same-session", Output: "completed answer"}) {
				t.Fatal("report must remain pending after delivery failure")
			}
			wantBusy := observation != "exited" && observation != "not-started"
			if got := nativeSessionInUse(inFlight, "reply", "codex", "same-session"); got != wantBusy {
				t.Errorf("same conversation reply blocked=%v, want %v", got, wantBusy)
			}
			actor.claimAvailable(t.Context(), []*runtimeactor.Actor{runtime}, inFlight)
			if source.available == wantBusy {
				t.Errorf("claim capacity available=%v, want %v", source.available, !wantBusy)
			}
			actor.retryTerminalReports(t.Context(), inFlight)
			if inFlight[task.taskID] != task || task.terminalResult.Status != agentbridge.ResultCompleted {
				t.Fatal("pending original result lost or rewritten")
			}
			if !actor.retryTerminalReports(t.Context(), inFlight) || len(inFlight) != 0 {
				t.Fatal("original result was not delivered after transport recovered")
			}
		})
	}
}
