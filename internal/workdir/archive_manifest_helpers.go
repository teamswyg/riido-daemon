package workdir

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

func readArchiveRecord(path string) (ArchiveRecord, error) {
	body, err := os.ReadFile(path)
	if err != nil {
		return ArchiveRecord{}, fmt.Errorf("workdir: read archive manifest: %w", err)
	}
	var record ArchiveRecord
	if err := json.Unmarshal(body, &record); err != nil {
		return ArchiveRecord{}, fmt.Errorf("workdir: decode archive manifest: %w", err)
	}
	return record, nil
}

func cleanupEligible(record ArchiveRecord, cutoff time.Time) bool {
	if record.SchemaVersion != ArchiveRecordSchemaVersion {
		return false
	}
	if record.RetentionMode != RetentionModeKeepInPlace {
		return false
	}
	if record.ArchivedAt.IsZero() {
		return false
	}
	return record.ArchivedAt.UTC().Before(cutoff)
}
