package supervisor

import (
	"context"
	"sync"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/runtimeactor"
	"github.com/teamswyg/riido-daemon/internal/workdir"
)

type recoveryCancellationSource struct {
	*controlplane.MemorySource
	watched chan struct{}
	once    sync.Once
}

func (s *recoveryCancellationSource) WatchCancellation(ctx context.Context, id string) (<-chan error, error) {
	ch, err := s.MemorySource.WatchCancellation(ctx, id)
	if err == nil {
		s.once.Do(func() { close(s.watched) })
	}
	return ch, err
}

type recoveryBackoffReporter struct {
	*reporterProbe
	backoff chan context.Context
}

func (r *recoveryBackoffReporter) ReportEvent(ctx context.Context, id string, ev agentbridge.Event) error {
	if ev.RecoveryContext == nil || ev.RecoveryContext.State != "backoff" {
		return r.reporterProbe.ReportEvent(ctx, id, ev)
	}
	select {
	case r.backoff <- ctx:
	case <-ctx.Done():
		return ctx.Err()
	}
	// Do not advance merely because a delayed test consumer read an old event.
	// Cancellation must interrupt this real reporter call and its original claim.
	<-ctx.Done()
	return ctx.Err()
}

func startBackoffCancellationSupervisor(t *testing.T, req bridge.TaskRequest, rt *runtimeactor.Actor) (*recoveryBackoffReporter, *recoveryCancellationSource) {
	t.Helper()
	source := &recoveryCancellationSource{MemorySource: controlplane.NewMemorySource(), watched: make(chan struct{})}
	source.Enqueue(req)
	reporter := &recoveryBackoffReporter{reporterProbe: newReporterProbe(), backoff: make(chan context.Context)}
	reporter.events = make(chan agentbridge.Event, 100)
	startRoutingSupervisor(t, Config{DaemonID: "daemon", Source: source, Reporter: reporter, Runtimes: []*runtimeactor.Actor{rt}, Workdir: workdir.NewFSAdapter(t.TempDir())})
	return reporter, source
}
