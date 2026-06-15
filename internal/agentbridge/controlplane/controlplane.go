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
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
	"github.com/teamswyg/riido-daemon/pkg/util/fileutil"
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

func (r *FileReporter) appendRecord(ctx context.Context, rec FileReportRecord) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	if rec.TaskID == "" {
		return controlPlaneErrorf(ErrControlPlaneInput, "file-reporter.append", "empty taskID")
	}
	path := r.reportPath(rec.TaskID)
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return controlPlaneWrapf(ErrControlPlaneReporter, "file-reporter.append", err, "open report file")
	}
	if err := json.NewEncoder(f).Encode(rec); err != nil {
		_ = f.Close()
		return controlPlaneWrapf(ErrControlPlaneReporter, "file-reporter.append", err, "encode report record")
	}
	if err := f.Close(); err != nil {
		return controlPlaneWrapf(ErrControlPlaneReporter, "file-reporter.append", err, "close report file")
	}
	return nil
}

func (r *FileReporter) reportPath(taskID string) string {
	sum := sha256.Sum256([]byte(taskID))
	return filepath.Join(r.dir, fmt.Sprintf("%x.jsonl", sum[:]))
}

// ----- FileQueueSource -----

// FileQueueSource reads JSON-encoded TaskRequest files from a directory
// and writes runtime registry/heartbeat records under dir/runtimes/.
// Each successful ClaimTask atomically moves the top-level task file
// into dir/claims/ and replaces it with a claim receipt, so the same
// task is not replayed even if multiple daemon processes poll the same
// local queue. Useful for batch testing and for ad-hoc CLI-driven queues.
type FileQueueSource struct {
	dir string
	now func() time.Time
}

func NewFileQueueSource(dir string) (*FileQueueSource, error) {
	info, err := os.Stat(dir)
	if err != nil {
		return nil, controlPlaneWrapf(ErrControlPlaneQueue, "file-queue.new", err, "stat queue dir")
	}
	if !info.IsDir() {
		return nil, controlPlaneErrorf(ErrControlPlaneQueue, "file-queue.new", "%s is not a directory", dir)
	}
	return &FileQueueSource{dir: dir, now: time.Now}, nil
}

func (s *FileQueueSource) RegisterRuntime(ctx context.Context, rt RuntimeRegistration) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	if rt.RuntimeID == "" {
		return controlPlaneErrorf(ErrControlPlaneRuntime, "file-queue.register-runtime", "empty RuntimeID")
	}
	rec := RegisteredRuntime{
		RuntimeRegistration: rt,
		LastHeartbeat:       s.now().UTC(),
	}
	return s.writeRuntimeRecord(rec)
}

func (s *FileQueueSource) DeregisterRuntime(ctx context.Context, runtimeID string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	if runtimeID == "" {
		return controlPlaneErrorf(ErrControlPlaneRuntime, "file-queue.deregister-runtime", "empty RuntimeID")
	}
	if err := os.Remove(s.runtimePath(runtimeID)); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return controlPlaneWrapf(ErrControlPlaneRegistry, "file-queue.deregister-runtime", err, "deregister runtime")
	}
	return nil
}

func (s *FileQueueSource) Heartbeat(ctx context.Context, hb RuntimeHeartbeat) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	if hb.RuntimeID == "" {
		return controlPlaneErrorf(ErrControlPlaneRuntime, "file-queue.heartbeat", "empty RuntimeID")
	}
	path := s.runtimePath(hb.RuntimeID)
	body, err := os.ReadFile(path)
	if err != nil {
		return controlPlaneWrapf(ErrControlPlaneRegistry, "file-queue.heartbeat", err, "read runtime registry")
	}
	rec, err := parseRuntimeRecord(body)
	if err != nil {
		return err
	}
	rec.LastHeartbeat = s.now().UTC()
	applyHeartbeat(&rec.RuntimeRegistration, hb)
	return s.writeRuntimeRecord(rec)
}

func (s *FileQueueSource) ClaimTask(ctx context.Context, runtimeID string) (*bridge.TaskRequest, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		return nil, controlPlaneWrapf(ErrControlPlaneQueue, "file-queue.claim-task", err, "read queue dir")
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name() < entries[j].Name() })
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".json" {
			continue
		}
		path := filepath.Join(s.dir, e.Name())
		body, err := os.ReadFile(path)
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				continue
			}
			return nil, controlPlaneWrapf(ErrControlPlaneQueue, "file-queue.claim-task", err, "read task file")
		}
		var req bridge.TaskRequest
		if err := json.Unmarshal(body, &req); err != nil {
			return nil, controlPlaneWrapf(ErrControlPlaneQueue, "file-queue.claim-task", err, "parse %s", path)
		}
		available, ok, err := s.runtimeProviderAvailable(runtimeID, string(req.Provider))
		if err != nil {
			return nil, err
		}
		if ok && !available {
			continue
		}
		claimPath, err := s.moveTaskToClaim(path, runtimeID)
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				continue // raced with another claim
			}
			return nil, err
		}
		rec := FileClaimRecord{
			SchemaVersion: FileClaimRecordSchemaVersion,
			TaskID:        req.ID,
			RuntimeID:     runtimeID,
			SourceFile:    e.Name(),
			ClaimedAt:     s.now().UTC(),
			Task:          req,
		}
		if err := fileutil.WriteJSONAtomic(claimPath, rec); err != nil {
			return nil, controlPlaneWrapf(ErrControlPlanePersistence, "file-queue.claim-task", err, "write claim receipt")
		}
		return &req, nil
	}
	return nil, nil
}

func (s *FileQueueSource) moveTaskToClaim(path, runtimeID string) (string, error) {
	claimsDir := filepath.Join(s.dir, "claims")
	if err := os.MkdirAll(claimsDir, 0o755); err != nil {
		return "", controlPlaneWrapf(ErrControlPlanePersistence, "file-queue.move-claim", err, "create claims dir")
	}
	runtimeHash := sha256.Sum256([]byte(runtimeID))
	tmp, err := os.CreateTemp(claimsDir, fmt.Sprintf("%020d-%x-*.json", s.now().UTC().UnixNano(), runtimeHash[:4]))
	if err != nil {
		return "", controlPlaneWrapf(ErrControlPlanePersistence, "file-queue.move-claim", err, "reserve claim path")
	}
	claimPath := tmp.Name()
	if err := tmp.Close(); err != nil {
		_ = os.Remove(claimPath)
		return "", controlPlaneWrapf(ErrControlPlanePersistence, "file-queue.move-claim", err, "close claim path reservation")
	}
	if err := os.Remove(claimPath); err != nil {
		return "", controlPlaneWrapf(ErrControlPlanePersistence, "file-queue.move-claim", err, "release claim path reservation")
	}
	if err := os.Rename(path, claimPath); err != nil {
		return "", controlPlaneWrapf(ErrControlPlanePersistence, "file-queue.move-claim", err, "rename task to claim")
	}
	return claimPath, nil
}

func (s *FileQueueSource) runtimeProviderAvailable(runtimeID, provider string) (bool, bool, error) {
	provider = strings.TrimSpace(provider)
	if runtimeID == "" || provider == "" {
		return true, false, nil
	}
	body, err := os.ReadFile(s.runtimePath(runtimeID))
	if errors.Is(err, fs.ErrNotExist) {
		return true, false, nil
	}
	if err != nil {
		return false, false, controlPlaneWrapf(ErrControlPlaneRegistry, "file-queue.runtime-provider-available", err, "read runtime registry")
	}
	rec, err := parseRuntimeRecord(body)
	if err != nil {
		return false, false, err
	}
	key := "provider." + provider + ".available"
	if available, ok := rec.Capabilities[key]; ok {
		return available, true, nil
	}
	for capabilityKey := range rec.Capabilities {
		if strings.HasPrefix(capabilityKey, "provider.") && strings.HasSuffix(capabilityKey, ".available") {
			return false, true, nil
		}
	}
	return true, false, nil
}

func (s *FileQueueSource) WatchCancellation(_ context.Context, _ string) (<-chan error, error) {
	// File queue has no out-of-band cancel channel; return a closed
	// channel so the caller can range over it without blocking.
	ch := make(chan error)
	close(ch)
	return ch, nil
}

func (s *FileQueueSource) runtimePath(runtimeID string) string {
	sum := sha256.Sum256([]byte(runtimeID))
	return filepath.Join(s.dir, "runtimes", fmt.Sprintf("%x.json", sum[:]))
}

func parseRuntimeRecord(body []byte) (RegisteredRuntime, error) {
	var rec RegisteredRuntime
	if err := json.Unmarshal(body, &rec); err != nil {
		return RegisteredRuntime{}, controlPlaneWrapf(ErrControlPlaneRegistry, "runtime-registry.parse", err, "parse runtime registry")
	}
	return rec, nil
}

func applyHeartbeat(reg *RuntimeRegistration, hb RuntimeHeartbeat) {
	if hb.RuntimeID != "" {
		reg.RuntimeID = hb.RuntimeID
	}
	if hb.DeviceName != "" {
		reg.DeviceName = hb.DeviceName
	}
	reg.UptimeSeconds = hb.UptimeSeconds
	reg.SlotLimit = hb.SlotLimit
	reg.SlotsInUse = hb.SlotsInUse
	reg.RunningTaskIDs = append([]string(nil), hb.RunningTaskIDs...)
	sort.Strings(reg.RunningTaskIDs)
}

func (s *FileQueueSource) writeRuntimeRecord(rec RegisteredRuntime) error {
	if err := fileutil.WriteJSONAtomic(s.runtimePath(rec.RuntimeID), rec); err != nil {
		return controlPlaneWrapf(ErrControlPlaneRegistry, "file-queue.write-runtime", err, "write runtime registry")
	}
	return nil
}
