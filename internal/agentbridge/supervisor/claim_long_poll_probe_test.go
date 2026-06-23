package supervisor

import (
	"context"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
)

type claimLongPollProbe struct {
	claims chan claimLongPollRecord
}

type claimLongPollRecord struct {
	runtimeID string
	longPoll  bool
}

func newClaimLongPollProbe() *claimLongPollProbe {
	return &claimLongPollProbe{claims: make(chan claimLongPollRecord, 4)}
}

func (s *claimLongPollProbe) RegisterRuntime(context.Context, controlplane.RuntimeRegistration) error {
	return nil
}

func (s *claimLongPollProbe) DeregisterRuntime(context.Context, string) error { return nil }

func (s *claimLongPollProbe) Heartbeat(context.Context, controlplane.RuntimeHeartbeat) error {
	return nil
}

func (s *claimLongPollProbe) ClaimTask(ctx context.Context, runtimeID string) (*bridge.TaskRequest, error) {
	s.claims <- claimLongPollRecord{runtimeID: runtimeID, longPoll: controlplane.ClaimLongPollEnabled(ctx)}
	return nil, nil
}

func (s *claimLongPollProbe) WatchCancellation(context.Context, string) (<-chan error, error) {
	return make(chan error), nil
}

func (s *claimLongPollProbe) expectClaim(t *testing.T) claimLongPollRecord {
	t.Helper()
	select {
	case claim := <-s.claims:
		return claim
	case <-time.After(time.Second):
		t.Fatal("claim was not observed")
		return claimLongPollRecord{}
	}
}
