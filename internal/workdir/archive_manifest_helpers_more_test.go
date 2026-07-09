package workdir

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCleanupEligibleRequiresValidKeepInPlaceArchiveBeforeCutoff(t *testing.T) {
	cutoff := time.Date(2026, 7, 9, 12, 0, 0, 0, time.UTC)
	base := ArchiveRecord{
		SchemaVersion: ArchiveRecordSchemaVersion,
		RetentionMode: RetentionModeKeepInPlace,
		ArchivedAt:    cutoff.Add(-time.Second),
	}
	if !cleanupEligible(base, cutoff) {
		t.Fatal("expected valid old keep-in-place archive to be eligible")
	}
	for name, record := range map[string]ArchiveRecord{
		"wrong schema":    withArchiveSchema(base, "old"),
		"wrong retention": withArchiveRetention(base, "external"),
		"zero time":       withArchiveTime(base, time.Time{}),
		"equal cutoff":    withArchiveTime(base, cutoff),
		"newer archive":   withArchiveTime(base, cutoff.Add(time.Second)),
	} {
		if cleanupEligible(record, cutoff) {
			t.Fatalf("%s should not be cleanup eligible: %#v", name, record)
		}
	}
}

func TestReadArchiveRecordRejectsMissingAndMalformedManifest(t *testing.T) {
	if _, err := readArchiveRecord(filepath.Join(t.TempDir(), "archive.json")); err == nil {
		t.Fatal("expected missing archive manifest error")
	}
	path := filepath.Join(t.TempDir(), "archive.json")
	if err := os.WriteFile(path, []byte(`{"schema_version":`), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := readArchiveRecord(path); err == nil {
		t.Fatal("expected malformed archive manifest error")
	}
}

func withArchiveSchema(in ArchiveRecord, schema string) ArchiveRecord {
	in.SchemaVersion = schema
	return in
}

func withArchiveRetention(in ArchiveRecord, retention string) ArchiveRecord {
	in.RetentionMode = retention
	return in
}

func withArchiveTime(in ArchiveRecord, archivedAt time.Time) ArchiveRecord {
	in.ArchivedAt = archivedAt
	return in
}
