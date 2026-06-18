package controlplane

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func readOnlyFileReportRecords(t *testing.T, dir string) []FileReportRecord {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected one report file, got %d", len(entries))
	}
	body, err := os.ReadFile(filepath.Join(dir, entries[0].Name()))
	if err != nil {
		t.Fatal(err)
	}
	return decodeFileReportRecords(t, body)
}

func decodeFileReportRecords(t *testing.T, body []byte) []FileReportRecord {
	t.Helper()
	dec := json.NewDecoder(bytes.NewReader(body))
	var records []FileReportRecord
	for {
		var rec FileReportRecord
		err := dec.Decode(&rec)
		if errors.Is(err, io.EOF) {
			return records
		}
		if err != nil {
			t.Fatal(err)
		}
		records = append(records, rec)
	}
}

func assertStartedFileReportRecord(t *testing.T, record FileReportRecord) {
	t.Helper()
	if record.Type != "started" || record.TaskID != "task-1" {
		t.Fatalf("started record: %+v", record)
	}
}

func assertEventFileReportRecord(t *testing.T, record FileReportRecord) {
	t.Helper()
	if record.Event == nil || record.Event.Text != "hello" {
		t.Fatalf("event record: %+v", record)
	}
}

func assertResultFileReportRecord(t *testing.T, record FileReportRecord) {
	t.Helper()
	if record.Result == nil || record.Result.Status != agentbridge.ResultCompleted {
		t.Fatalf("result record: %+v", record)
	}
}
