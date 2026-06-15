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
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
)

const (
	MetadataRuntimeLeaseID               = "runtime_lease_id"
	MetadataRuntimeFencingToken          = "runtime_fencing_token"
	MetadataRuntimeCapabilityFingerprint = "runtime_capability_fingerprint"
	// MetadataTaskID preserves the logical Riido task id when a source uses
	// TaskRequest.ID as a run/execution id. Sources that already use
	// TaskRequest.ID as the logical task id do not need to set it.
	MetadataTaskID = "task_id"
)

// RuntimeRegistration is the payload the daemon hands to the control
// plane when announcing a local runtime.
type RuntimeModel struct {
	ModelID   string `json:"model_id"`
	Label     string `json:"label"`
	IsDefault bool   `json:"is_default"`
}

type RuntimeRegistration struct {
	DaemonID             string            `json:"daemon_id"`
	RuntimeID            string            `json:"runtime_id"`
	Provider             string            `json:"provider"`
	Executable           string            `json:"executable,omitempty"`
	Version              string            `json:"version,omitempty"`
	Capabilities         map[string]bool   `json:"capabilities,omitempty"`
	CapabilityAttributes map[string]string `json:"capability_attributes,omitempty"`
	DeviceName           string            `json:"device_name,omitempty"`
	Models               []RuntimeModel    `json:"models,omitempty"`
	StartedAt            time.Time         `json:"started_at"`
	UptimeSeconds        int64             `json:"uptime_seconds,omitempty"`
	SlotLimit            int               `json:"slot_limit,omitempty"`
	SlotsInUse           int               `json:"slots_in_use,omitempty"`
	RunningTaskIDs       []string          `json:"running_task_ids,omitempty"`
}

// RuntimeHeartbeat is the periodic liveness/capacity snapshot the
// daemon hands to a task source after registration.
type RuntimeHeartbeat struct {
	RuntimeID      string   `json:"runtime_id"`
	UptimeSeconds  int64    `json:"uptime_seconds,omitempty"`
	DeviceName     string   `json:"device_name,omitempty"`
	SlotLimit      int      `json:"slot_limit,omitempty"`
	SlotsInUse     int      `json:"slots_in_use,omitempty"`
	RunningTaskIDs []string `json:"running_task_ids,omitempty"`
}

// RegisteredRuntime is the control plane's view of a runtime, including
// the last heartbeat it observed.
type RegisteredRuntime struct {
	RuntimeRegistration
	LastHeartbeat time.Time `json:"last_heartbeat"`
}

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

// TaskReportContext carries claim-time lease metadata alongside reporter
// calls without widening every reporter method signature.
type TaskReportContext struct {
	RuntimeLeaseID               string
	RuntimeFencingToken          int64
	RuntimeFencingTokenSet       bool
	RuntimeCapabilityFingerprint string
}

type taskReportContextKey struct{}

// ContextWithTaskReport attaches claim-time lease metadata to reporter calls.
func ContextWithTaskReport(ctx context.Context, report TaskReportContext) context.Context {
	return context.WithValue(ctx, taskReportContextKey{}, report)
}

// TaskReportContextFromContext returns claim-time lease metadata attached to ctx.
func TaskReportContextFromContext(ctx context.Context) (TaskReportContext, bool) {
	report, ok := ctx.Value(taskReportContextKey{}).(TaskReportContext)
	return report, ok
}

// TaskReportContextFromMetadata extracts claim-time lease metadata from a task request.
func TaskReportContextFromMetadata(metadata map[string]string) (TaskReportContext, bool) {
	if len(metadata) == 0 {
		return TaskReportContext{}, false
	}
	report := TaskReportContext{
		RuntimeLeaseID:               strings.TrimSpace(metadata[MetadataRuntimeLeaseID]),
		RuntimeCapabilityFingerprint: strings.TrimSpace(metadata[MetadataRuntimeCapabilityFingerprint]),
	}
	if raw := strings.TrimSpace(metadata[MetadataRuntimeFencingToken]); raw != "" {
		token, err := strconv.ParseInt(raw, 10, 64)
		if err == nil {
			report.RuntimeFencingToken = token
			report.RuntimeFencingTokenSet = true
		}
	}
	return report, report.RuntimeLeaseID != "" || report.RuntimeFencingTokenSet || report.RuntimeCapabilityFingerprint != ""
}

// ----- MemorySource: in-memory queue + runtime registry -----

// MemorySource is the simplest TaskSourcePort: tasks live in a FIFO,
// runtimes in a map. Intended for tests, offline mode, and bootstrap.
//
// All state is owned by the calling goroutine — the source itself is
// NOT a separate actor. Callers (daemon main goroutine or a
// SupervisorActor) serialize access. We do not use sync.Mutex here.
type MemorySource struct {
	queue     []bridge.TaskRequest
	runtimes  map[string]*RegisteredRuntime
	cancelChs map[string]chan error
	now       func() time.Time
}

func NewMemorySource() *MemorySource {
	return &MemorySource{
		runtimes:  map[string]*RegisteredRuntime{},
		cancelChs: map[string]chan error{},
		now:       time.Now,
	}
}

// Enqueue appends a task to the internal queue (test/daemon helper).
func (s *MemorySource) Enqueue(req bridge.TaskRequest) {
	s.queue = append(s.queue, req)
}

// Registered returns a snapshot of registered runtimes sorted by id.
func (s *MemorySource) Registered() []RegisteredRuntime {
	out := make([]RegisteredRuntime, 0, len(s.runtimes))
	for _, r := range s.runtimes {
		out = append(out, *r)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].RuntimeID < out[j].RuntimeID })
	return out
}

func (s *MemorySource) RegisterRuntime(_ context.Context, rt RuntimeRegistration) error {
	if rt.RuntimeID == "" {
		return controlPlaneErrorf(ErrControlPlaneRuntime, "memory.register-runtime", "empty RuntimeID")
	}
	s.runtimes[rt.RuntimeID] = &RegisteredRuntime{
		RuntimeRegistration: rt,
		LastHeartbeat:       s.now(),
	}
	return nil
}
