package supervisor

import (
	"context"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/runtimeactor"
	"github.com/teamswyg/riido-daemon/internal/process"
)

type blockingIdlePollSource struct{ *claimLongPollProbe }

func (s blockingIdlePollSource) ClaimTask(ctx context.Context, _ string) (*bridge.TaskRequest, error) {
	if controlplane.ClaimLongPollEnabled(ctx) {
		<-ctx.Done()
		return nil, ctx.Err()
	}
	return nil, nil
}

func TestSupervisorActiveDeliveryDoesNotWaitForIdlePoll(t *testing.T) {
	for _, queued := range []bool{false, true} {
		t.Run(map[bool]string{false: "in-flight", true: "queued-message"}[queued], func(t *testing.T) {
			rt := startNamedRuntime(t, process.NewFake(), "rt-idle", "codex")
			a := &Actor{cfg: Config{Source: blockingIdlePollSource{newClaimLongPollProbe()}}, mailbox: make(chan envelope, 1)}
			inFlight := map[string]*runningTask{}
			if queued {
				a.mailbox <- envelope{taskResult: &taskResultMsg{taskID: "finished"}}
			} else {
				inFlight["preparing-or-running"] = &runningTask{}
			}
			ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
			defer cancel()
			start := time.Now()
			a.claimAvailable(ctx, []*runtimeactor.Actor{rt}, inFlight)
			t.Logf("claim wave duration: %s", time.Since(start))
			if ctx.Err() != nil {
				t.Fatal("pending delivery waited for the idle long-poll timeout")
			}
		})
	}
}
