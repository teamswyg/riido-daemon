// Package controlplane defines the ports that the agent daemon uses to talk to
// whatever supplies tasks and consumes results.
//
// The daemon core depends only on these interfaces. Specific remote SaaS,
// task DB, or local projection adapters live in separate packages and plug in.
//
// In-tree adapters provided now:
//   - MemorySource / MemoryReporter: RAM-only, for tests and offline mode.
//   - FileQueueSource: JSON task files, claim receipts, and runtime registry files in a directory.
//   - FileReporter: task-scoped JSONL receipts in a directory.
//
// Not part of this package:
//   - supervisor polling / runtime selection.
//   - SaaS assignment HTTP polling and event sync adapters.
//   - task DB / project / mwsd-backed adapters.
package controlplane

import (
	"context"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
)

// TaskSourcePort is how the daemon pulls work in.
//
// ClaimTask returns nil, nil when no work is available. Concrete
// adapters MUST be safe to call from a single owning goroutine; the
// daemon does not need them to support concurrent claim from multiple
// goroutines at this layer.
type TaskSourcePort interface {
	RegisterRuntime(ctx context.Context, rt RuntimeRegistration) error
	DeregisterRuntime(ctx context.Context, runtimeID string) error
	Heartbeat(ctx context.Context, hb RuntimeHeartbeat) error
	ClaimTask(ctx context.Context, runtimeID string) (*bridge.TaskRequest, error)
	WatchCancellation(ctx context.Context, taskID string) (<-chan error, error)
}

// TaskReporterPort is how the daemon reports progress and outcome.
type TaskReporterPort interface {
	StartTask(ctx context.Context, taskID string) error
	ReportEvent(ctx context.Context, taskID string, ev agentbridge.Event) error
	CompleteTask(ctx context.Context, taskID string, res agentbridge.Result) error
}
