package controlplane

import (
	"context"
	"os"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

// FileReportRecord is one JSONL record written by FileReporter.
type FileReportRecord struct {
	Type   string              `json:"type"`
	TaskID string              `json:"task_id"`
	At     time.Time           `json:"at"`
	Event  *agentbridge.Event  `json:"event,omitempty"`
	Result *agentbridge.Result `json:"result,omitempty"`
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
