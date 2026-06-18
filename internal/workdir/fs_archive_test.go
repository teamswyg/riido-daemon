package workdir

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestArchiveWritesKeepInPlaceManifest(t *testing.T) {
	a, ws := preparedTestWorkspace(t, "run-1")
	archivedAt := time.Date(2026, 5, 24, 1, 2, 3, 4, time.UTC)
	record, err := a.Archive(ws, ArchiveRequest{ResultStatus: "completed", ArchivedAt: archivedAt})
	if err != nil {
		t.Fatalf("Archive: %v", err)
	}
	if record.SchemaVersion != ArchiveRecordSchemaVersion ||
		record.RetentionMode != RetentionModeKeepInPlace ||
		record.WorkdirPath != ws.Workdir ||
		!strings.HasPrefix(record.ArchiveURI, "file://") {
		t.Fatalf("archive record = %+v", record)
	}
	assertArchiveManifest(t, ws, archivedAt)
}

func assertArchiveManifest(t *testing.T, ws Workspace, archivedAt time.Time) {
	t.Helper()
	bytes, err := os.ReadFile(filepath.Join(ws.Root, "archive.json"))
	if err != nil {
		t.Fatalf("read archive manifest: %v", err)
	}
	var decoded ArchiveRecord
	if err := json.Unmarshal(bytes, &decoded); err != nil {
		t.Fatalf("decode archive manifest: %v", err)
	}
	if decoded.SchemaVersion != ArchiveRecordSchemaVersion ||
		decoded.RetentionMode != RetentionModeKeepInPlace ||
		decoded.ResultStatus != "completed" ||
		!decoded.ArchivedAt.Equal(archivedAt) {
		t.Fatalf("archive manifest = %+v", decoded)
	}
}
