package supervisor

import (
	"context"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
	"github.com/teamswyg/riido-daemon/internal/process"
)

func TestSupervisorStopCancelsInFlightClaim(t *testing.T) {
	source := newBlockingClaimSource()
	actor := startClaimCancellationSupervisor(t, source, "daemon-claim-cancel")
	expectSignal(t, source.started, "claim did not start")

	stopCtx, cancel := context.WithTimeout(context.Background(), 250*time.Millisecond)
	defer cancel()
	if err := actor.Stop(stopCtx); err != nil {
		t.Fatalf("Stop should cancel in-flight claim: %v", err)
	}
	expectImmediateSignal(t, source.canceled, "claim context was not canceled")
}

func startClaimCancellationSupervisor(
	t *testing.T,
	source controlplane.TaskSourcePort,
	daemonID string,
) *Actor {
	t.Helper()
	actor, err := New(Config{
		DaemonID:       daemonID,
		Runtime:        startRuntime(t, process.NewFake()),
		Source:         source,
		Reporter:       newReporterProbe(),
		PollEvery:      10 * time.Millisecond,
		HeartbeatEvery: time.Hour,
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if err := actor.Start(context.Background()); err != nil {
		t.Fatalf("Start: %v", err)
	}
	return actor
}
