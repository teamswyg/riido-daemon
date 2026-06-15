package controlplane

import (
	"context"
	"os"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
)

func (s *MemorySource) DeregisterRuntime(_ context.Context, runtimeID string) error {
	if _, ok := s.runtimes[runtimeID]; !ok {
		return controlPlaneErrorf(ErrControlPlaneRuntime, "memory.deregister-runtime", "unknown runtime %q", runtimeID)
	}
	delete(s.runtimes, runtimeID)
	return nil
}

func (s *MemorySource) Heartbeat(_ context.Context, hb RuntimeHeartbeat) error {
	r, ok := s.runtimes[hb.RuntimeID]
	if !ok {
		return controlPlaneErrorf(ErrControlPlaneRuntime, "memory.heartbeat", "heartbeat for unknown runtime %q", hb.RuntimeID)
	}
	r.LastHeartbeat = s.now()
	applyHeartbeat(&r.RuntimeRegistration, hb)
	return nil
}

func (s *MemorySource) ClaimTask(_ context.Context, _ string) (*bridge.TaskRequest, error) {
	if len(s.queue) == 0 {
		return nil, nil
	}
	req := s.queue[0]
	s.queue = s.queue[1:]
	return &req, nil
}

func (s *MemorySource) WatchCancellation(_ context.Context, taskID string) (<-chan error, error) {
	if taskID == "" {
		return nil, controlPlaneErrorf(ErrControlPlaneInput, "memory.watch-cancellation", "empty taskID")
	}
	ch := make(chan error, 1)
	s.cancelChs[taskID] = ch
	return ch, nil
}

// Cancel delivers a cancellation cause to a previously watched task.
// If no watcher exists, the cause is dropped (no-op).
func (s *MemorySource) Cancel(taskID string, cause error) {
	ch, ok := s.cancelChs[taskID]
	if !ok {
		return
	}
	select {
	case ch <- cause:
	default:
	}
}

// ----- MemoryReporter -----

type TaskRecord struct {
	Started bool
	Events  []agentbridge.Event
	Result  agentbridge.Result
}

// MemoryReporter stores per-task evidence in RAM, indexable by task id.
// Single-goroutine ownership — no mutex.
type MemoryReporter struct {
	records map[string]*TaskRecord
}

func NewMemoryReporter() *MemoryReporter {
	return &MemoryReporter{records: map[string]*TaskRecord{}}
}

func (r *MemoryReporter) record(taskID string) *TaskRecord {
	rec, ok := r.records[taskID]
	if !ok {
		rec = &TaskRecord{}
		r.records[taskID] = rec
	}
	return rec
}

func (r *MemoryReporter) StartTask(_ context.Context, taskID string) error {
	r.record(taskID).Started = true
	return nil
}

func (r *MemoryReporter) ReportEvent(_ context.Context, taskID string, ev agentbridge.Event) error {
	r.record(taskID).Events = append(r.record(taskID).Events, ev)
	return nil
}

func (r *MemoryReporter) CompleteTask(_ context.Context, taskID string, res agentbridge.Result) error {
	r.record(taskID).Result = res
	return nil
}

// Recorded returns a snapshot of the task's record. If the task is
// unknown, an empty record is returned (not nil).
func (r *MemoryReporter) Recorded(taskID string) TaskRecord {
	if rec, ok := r.records[taskID]; ok {
		out := *rec
		out.Events = append([]agentbridge.Event(nil), rec.Events...)
		return out
	}
	return TaskRecord{}
}

// ----- FileReporter -----

// FileReportRecord is one JSONL record written by FileReporter.
type FileReportRecord struct {
	Type   string              `json:"type"`
	TaskID string              `json:"task_id"`
	At     time.Time           `json:"at"`
	Event  *agentbridge.Event  `json:"event,omitempty"`
	Result *agentbridge.Result `json:"result,omitempty"`
}

// FileClaimRecordSchemaVersion is the local file queue claim receipt schema.
const FileClaimRecordSchemaVersion = "riido-file-queue-claim.v1"

// FileClaimRecord is written under queue/claims/ when FileQueueSource
// atomically claims a top-level task JSON file.
type FileClaimRecord struct {
	SchemaVersion string             `json:"schema_version"`
	TaskID        string             `json:"task_id"`
	RuntimeID     string             `json:"runtime_id"`
	SourceFile    string             `json:"source_file"`
	ClaimedAt     time.Time          `json:"claimed_at"`
	Task          bridge.TaskRequest `json:"task"`
}

// FileReporter appends task progress and terminal results to per-task
// JSONL files. Like the other in-tree control-plane adapters, it is
// owned by the SupervisorActor goroutine; no mutex is required here.
type FileReporter struct {
	dir string
	now func() time.Time
}

func NewFileReporter(dir string) (*FileReporter, error) {
	if dir == "" {
		return nil, controlPlaneErrorf(ErrControlPlaneInput, "file-reporter.new", "empty report dir")
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, controlPlaneWrapf(ErrControlPlaneReporter, "file-reporter.new", err, "create report dir")
	}
	return &FileReporter{dir: dir, now: time.Now}, nil
}

func (r *FileReporter) StartTask(ctx context.Context, taskID string) error {
	return r.appendRecord(ctx, FileReportRecord{Type: "started", TaskID: taskID, At: r.now().UTC()})
}

func (r *FileReporter) ReportEvent(ctx context.Context, taskID string, ev agentbridge.Event) error {
	return r.appendRecord(ctx, FileReportRecord{Type: "event", TaskID: taskID, At: r.now().UTC(), Event: &ev})
}

func (r *FileReporter) CompleteTask(ctx context.Context, taskID string, res agentbridge.Result) error {
	return r.appendRecord(ctx, FileReportRecord{Type: "result", TaskID: taskID, At: r.now().UTC(), Result: &res})
}
