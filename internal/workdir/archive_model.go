package workdir

import "time"

// ArchiveRequest is the terminal-run input for Archive.
type ArchiveRequest struct {
	ResultStatus string
	ArchivedAt   time.Time
}

// ArchiveRecord is the local archive manifest written at run-root/archive.json.
type ArchiveRecord struct {
	SchemaVersion string    `json:"schema_version"`
	WorkdirPath   string    `json:"workdir_path"`
	ArchiveURI    string    `json:"archive_uri"`
	RetentionMode string    `json:"retention_mode"`
	ResultStatus  string    `json:"result_status"`
	ArchivedAt    time.Time `json:"archived_at"`
}

// CleanupRequest defines an explicit retention cleanup pass. The
// daemon supplies ArchivedBefore from its Factor 12 retention config.
type CleanupRequest struct {
	ArchivedBefore time.Time
	RemovedAt      time.Time
}

// CleanupRecord describes one run root removed by cleanup.
type CleanupRecord struct {
	RunRoot   string        `json:"run_root"`
	Archive   ArchiveRecord `json:"archive"`
	RemovedAt time.Time     `json:"removed_at"`
}

// CleanupResult summarizes an archived-run cleanup pass.
type CleanupResult struct {
	ScannedArchiveRecords int             `json:"scanned_archive_records"`
	Removed               []CleanupRecord `json:"removed"`
}
