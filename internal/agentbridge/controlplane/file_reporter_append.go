package controlplane

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
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
