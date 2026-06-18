package workdir

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Archive writes the local keep-in-place archive manifest for a terminal run.
func (a *FSAdapter) Archive(ws Workspace, req ArchiveRequest) (ArchiveRecord, error) {
	if err := validateArchiveWorkspace(ws); err != nil {
		return ArchiveRecord{}, err
	}
	status := strings.TrimSpace(req.ResultStatus)
	if status == "" {
		return ArchiveRecord{}, errors.New("workdir: archive result status is required")
	}
	record := ArchiveRecord{
		SchemaVersion: ArchiveRecordSchemaVersion,
		WorkdirPath:   ws.Workdir,
		ArchiveURI:    localFileURI(ws.Root),
		RetentionMode: RetentionModeKeepInPlace,
		ResultStatus:  status,
		ArchivedAt:    archiveTime(req.ArchivedAt),
	}
	if err := os.MkdirAll(ws.Root, 0o755); err != nil {
		return ArchiveRecord{}, fmt.Errorf("workdir: mkdir archive root: %w", err)
	}
	if err := writeJSONAtomic(filepath.Join(ws.Root, "archive.json"), record); err != nil {
		return ArchiveRecord{}, fmt.Errorf("workdir: write archive manifest: %w", err)
	}
	return record, nil
}

func validateArchiveWorkspace(ws Workspace) error {
	if strings.TrimSpace(ws.Root) == "" {
		return errors.New("workdir: workspace root is required")
	}
	if strings.TrimSpace(ws.Workdir) == "" {
		return errors.New("workdir: workdir path is required")
	}
	return nil
}

func archiveTime(t time.Time) time.Time {
	if t.IsZero() {
		return time.Now().UTC()
	}
	return t.UTC()
}
