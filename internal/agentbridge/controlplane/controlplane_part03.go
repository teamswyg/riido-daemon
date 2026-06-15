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
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
	"github.com/teamswyg/riido-daemon/pkg/util/fileutil"
)

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
