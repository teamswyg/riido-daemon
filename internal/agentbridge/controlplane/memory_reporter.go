package controlplane

import (
	"context"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

type TaskRecord struct {
	Started bool
	Events  []agentbridge.Event
	Result  agentbridge.Result
}

// MemoryReporter stores per-task evidence in RAM, indexable by task id.
// Single-goroutine ownership -- no mutex.
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
