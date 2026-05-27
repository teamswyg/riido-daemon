package controlplane

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
)

func TestTaskReportContextRoundTripFromMetadata(t *testing.T) {
	report, ok := TaskReportContextFromMetadata(map[string]string{
		MetadataRuntimeLeaseID:               "runtime-lease:t-1:3",
		MetadataRuntimeFencingToken:          "3",
		MetadataRuntimeCapabilityFingerprint: "fp-1",
	})
	if !ok {
		t.Fatal("TaskReportContextFromMetadata returned ok=false")
	}
	if report.RuntimeLeaseID != "runtime-lease:t-1:3" || report.RuntimeFencingToken != 3 || !report.RuntimeFencingTokenSet || report.RuntimeCapabilityFingerprint != "fp-1" {
		t.Fatalf("unexpected report context: %+v", report)
	}
	ctx := ContextWithTaskReport(context.Background(), report)
	got, ok := TaskReportContextFromContext(ctx)
	if !ok || got != report {
		t.Fatalf("context round trip = %+v, %v", got, ok)
	}
}

// --- TaskSourcePort: MemorySource ---

func TestMemorySourceClaimReturnsNoTaskWhenEmpty(t *testing.T) {
	src := NewMemorySource()
	req, err := src.ClaimTask(context.Background(), "rt-1")
	if err != nil {
		t.Fatalf("Claim: %v", err)
	}
	if req != nil {
		t.Fatalf("expected nil request, got %+v", req)
	}
}

func TestMemorySourceQueueAndClaim(t *testing.T) {
	src := NewMemorySource()
	src.Enqueue(bridge.TaskRequest{ID: "t-1", Provider: "claude", Prompt: "hi"})
	src.Enqueue(bridge.TaskRequest{ID: "t-2", Provider: "codex", Prompt: "yo"})

	first, _ := src.ClaimTask(context.Background(), "rt-1")
	second, _ := src.ClaimTask(context.Background(), "rt-1")
	third, _ := src.ClaimTask(context.Background(), "rt-1")

	if first == nil || first.ID != "t-1" {
		t.Fatalf("first: %+v", first)
	}
	if second == nil || second.ID != "t-2" {
		t.Fatalf("second: %+v", second)
	}
	if third != nil {
		t.Fatalf("expected empty queue, got %+v", third)
	}
}

func TestMemorySourceRegisterDeregisterHeartbeat(t *testing.T) {
	src := NewMemorySource()
	now := time.Now()
	src.now = func() time.Time { return now }

	reg := RuntimeRegistration{
		DaemonID:  "d-1",
		RuntimeID: "rt-1",
		Provider:  "claude",
	}
	if err := src.RegisterRuntime(context.Background(), reg); err != nil {
		t.Fatalf("Register: %v", err)
	}

	if rts := src.Registered(); len(rts) != 1 || rts[0].RuntimeID != "rt-1" {
		t.Fatalf("registered: %+v", rts)
	}

	now = now.Add(15 * time.Second)
	if err := src.Heartbeat(context.Background(), RuntimeHeartbeat{RuntimeID: "rt-1", SlotLimit: 2, SlotsInUse: 1, RunningTaskIDs: []string{"task-1"}}); err != nil {
		t.Fatalf("Heartbeat: %v", err)
	}
	hb := src.Registered()[0].LastHeartbeat
	if !hb.Equal(now) {
		t.Fatalf("heartbeat: %v", hb)
	}
	if got := src.Registered()[0].SlotsInUse; got != 1 {
		t.Fatalf("slots in use after heartbeat = %d", got)
	}

	if err := src.DeregisterRuntime(context.Background(), "rt-1"); err != nil {
		t.Fatalf("Deregister: %v", err)
	}
	if rts := src.Registered(); len(rts) != 0 {
		t.Fatalf("expected empty after deregister: %+v", rts)
	}
}

func TestMemorySourceWatchCancellation(t *testing.T) {
	src := NewMemorySource()
	src.Enqueue(bridge.TaskRequest{ID: "t-1", Provider: "claude"})
	_, _ = src.ClaimTask(context.Background(), "rt-1")

	ch, err := src.WatchCancellation(context.Background(), "t-1")
	if err != nil {
		t.Fatalf("Watch: %v", err)
	}

	src.Cancel("t-1", errors.New("user cancel"))
	select {
	case cause := <-ch:
		if cause == nil || cause.Error() != "user cancel" {
			t.Fatalf("cause: %v", cause)
		}
	case <-time.After(time.Second):
		t.Fatal("cancellation not delivered")
	}
}

// --- TaskReporterPort: MemoryReporter ---

func TestMemoryReporterRoundTrip(t *testing.T) {
	rep := NewMemoryReporter()
	if err := rep.StartTask(context.Background(), "t-1"); err != nil {
		t.Fatal(err)
	}
	if err := rep.ReportEvent(context.Background(), "t-1", agentbridge.Event{Kind: agentbridge.EventTextDelta, Text: "hi"}); err != nil {
		t.Fatal(err)
	}
	if err := rep.CompleteTask(context.Background(), "t-1", agentbridge.Result{Status: agentbridge.ResultCompleted, Output: "ok"}); err != nil {
		t.Fatal(err)
	}
	rec := rep.Recorded("t-1")
	if !rec.Started || len(rec.Events) != 1 || rec.Result.Status != agentbridge.ResultCompleted {
		t.Fatalf("record: %+v", rec)
	}
}

func TestFileReporterWritesJSONL(t *testing.T) {
	dir := t.TempDir()
	rep, err := NewFileReporter(dir)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 5, 24, 12, 0, 0, 0, time.UTC)
	rep.now = func() time.Time { return now }

	if err := rep.StartTask(context.Background(), "task-1"); err != nil {
		t.Fatal(err)
	}
	if err := rep.ReportEvent(context.Background(), "task-1", agentbridge.Event{Kind: agentbridge.EventTextDelta, Text: "hello"}); err != nil {
		t.Fatal(err)
	}
	if err := rep.CompleteTask(context.Background(), "task-1", agentbridge.Result{Status: agentbridge.ResultCompleted, Output: "done"}); err != nil {
		t.Fatal(err)
	}

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
	dec := json.NewDecoder(bytes.NewReader(body))
	var records []FileReportRecord
	for {
		var rec FileReportRecord
		err := dec.Decode(&rec)
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		records = append(records, rec)
	}
	if len(records) != 3 {
		t.Fatalf("records: %+v", records)
	}
	if records[0].Type != "started" || records[0].TaskID != "task-1" {
		t.Fatalf("started record: %+v", records[0])
	}
	if records[1].Event == nil || records[1].Event.Text != "hello" {
		t.Fatalf("event record: %+v", records[1])
	}
	if records[2].Result == nil || records[2].Result.Status != agentbridge.ResultCompleted {
		t.Fatalf("result record: %+v", records[2])
	}
}

func TestFileReporterCreatesDir(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "reports")
	if _, err := NewFileReporter(dir); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !info.IsDir() {
		t.Fatalf("report path is not a dir: %s", dir)
	}
}

// --- FileQueueSource ---

func TestFileQueueSourceReadsTasksFromDir(t *testing.T) {
	dir := t.TempDir()
	src, err := NewFileQueueSource(dir)
	if err != nil {
		t.Fatal(err)
	}

	for i, p := range []string{"claude", "codex"} {
		req := bridge.TaskRequest{ID: "f-" + strconv.Itoa(i), Provider: bridge.Provider(p), Prompt: "x"}
		body, _ := json.Marshal(req)
		path := filepath.Join(dir, req.ID+".json")
		if err := os.WriteFile(path, body, 0o644); err != nil {
			t.Fatal(err)
		}
	}

	got1, err := src.ClaimTask(context.Background(), "rt-1")
	if err != nil || got1 == nil {
		t.Fatalf("claim 1: %v / %+v", err, got1)
	}
	got2, _ := src.ClaimTask(context.Background(), "rt-1")
	got3, _ := src.ClaimTask(context.Background(), "rt-1")
	if got1.ID == got2.ID {
		t.Fatalf("claim should not return same task twice")
	}
	if got3 != nil {
		t.Fatalf("expected empty, got %+v", got3)
	}

	// Tasks must be removed from the top-level dir after claim (no replay).
	if remaining := countTopLevelJSON(t, dir); remaining != 0 {
		t.Fatalf("file source should consume top-level tasks, got %d remaining", remaining)
	}
	claims := readClaimRecords(t, filepath.Join(dir, "claims"))
	if len(claims) != 2 {
		t.Fatalf("claim receipts = %+v", claims)
	}
	for _, rec := range claims {
		if rec.SchemaVersion != FileClaimRecordSchemaVersion || rec.RuntimeID != "rt-1" || rec.TaskID == "" {
			t.Fatalf("claim receipt mismatch: %+v", rec)
		}
	}
}

func TestFileQueueSourceSkipsUnavailableProviderForRuntime(t *testing.T) {
	dir := t.TempDir()
	src, err := NewFileQueueSource(dir)
	if err != nil {
		t.Fatal(err)
	}
	if err := src.RegisterRuntime(context.Background(), RuntimeRegistration{
		DaemonID:  "daemon-1",
		RuntimeID: "rt-claude",
		Provider:  "multi",
		Capabilities: map[string]bool{
			"provider.claude.available": true,
			"provider.codex.available":  false,
		},
	}); err != nil {
		t.Fatal(err)
	}
	if err := src.RegisterRuntime(context.Background(), RuntimeRegistration{
		DaemonID:  "daemon-2",
		RuntimeID: "rt-codex",
		Provider:  "multi",
		Capabilities: map[string]bool{
			"provider.claude.available": false,
			"provider.codex.available":  true,
		},
	}); err != nil {
		t.Fatal(err)
	}
	for _, req := range []bridge.TaskRequest{
		{ID: "codex-task", Provider: "codex", Prompt: "x"},
		{ID: "claude-task", Provider: "claude", Prompt: "x"},
	} {
		body, _ := json.Marshal(req)
		if err := os.WriteFile(filepath.Join(dir, req.ID+".json"), body, 0o644); err != nil {
			t.Fatal(err)
		}
	}

	first, err := src.ClaimTask(context.Background(), "rt-claude")
	if err != nil {
		t.Fatal(err)
	}
	if first == nil || first.ID != "claude-task" {
		t.Fatalf("rt-claude should skip codex and claim claude, got %+v", first)
	}
	if remaining := countTopLevelJSON(t, dir); remaining != 1 {
		t.Fatalf("codex task should remain for another runtime, remaining=%d", remaining)
	}

	second, err := src.ClaimTask(context.Background(), "rt-codex")
	if err != nil {
		t.Fatal(err)
	}
	if second == nil || second.ID != "codex-task" {
		t.Fatalf("rt-codex should claim remaining codex task, got %+v", second)
	}
	if claims := readClaimRecords(t, filepath.Join(dir, "claims")); len(claims) != 2 {
		t.Fatalf("claim receipts = %+v", claims)
	}
}

func TestFileQueueSourceClaimReceiptsDoNotOverwriteRepeatedFilenames(t *testing.T) {
	dir := t.TempDir()
	src, err := NewFileQueueSource(dir)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 5, 25, 1, 2, 3, 0, time.UTC)
	src.now = func() time.Time { return now }

	writeTaskFile := func(id string) {
		t.Helper()
		req := bridge.TaskRequest{ID: id, Provider: "codex", Prompt: "x"}
		body, _ := json.Marshal(req)
		if err := os.WriteFile(filepath.Join(dir, "task.json"), body, 0o644); err != nil {
			t.Fatal(err)
		}
	}

	writeTaskFile("first")
	if got, err := src.ClaimTask(context.Background(), "rt-1"); err != nil || got == nil || got.ID != "first" {
		t.Fatalf("first claim: %v / %+v", err, got)
	}
	writeTaskFile("second")
	if got, err := src.ClaimTask(context.Background(), "rt-1"); err != nil || got == nil || got.ID != "second" {
		t.Fatalf("second claim: %v / %+v", err, got)
	}

	claims := readClaimRecords(t, filepath.Join(dir, "claims"))
	if len(claims) != 2 {
		t.Fatalf("expected two claim receipts, got %+v", claims)
	}
	if claims[0].TaskID == claims[1].TaskID {
		t.Fatalf("claim receipts were overwritten: %+v", claims)
	}
}

func TestFileQueueSourceWritesRuntimeRegistry(t *testing.T) {
	dir := t.TempDir()
	src, err := NewFileQueueSource(dir)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 5, 24, 10, 0, 0, 0, time.UTC)
	src.now = func() time.Time { return now }

	reg := RuntimeRegistration{
		DaemonID:  "daemon-1",
		RuntimeID: "rt-1",
		Provider:  "multi",
		Capabilities: map[string]bool{
			"provider.codex.available": true,
		},
		DeviceName: "mac-mini",
		StartedAt:  now.Add(-time.Minute),
	}
	if err := src.RegisterRuntime(context.Background(), reg); err != nil {
		t.Fatalf("RegisterRuntime: %v", err)
	}

	rec := readFileRuntimeRecord(t, src.runtimePath("rt-1"))
	if rec.RuntimeID != "rt-1" || rec.DaemonID != "daemon-1" || rec.Provider != "multi" {
		t.Fatalf("runtime record identity: %+v", rec)
	}
	if !rec.LastHeartbeat.Equal(now) {
		t.Fatalf("last heartbeat = %v, want %v", rec.LastHeartbeat, now)
	}
	if !rec.Capabilities["provider.codex.available"] {
		t.Fatalf("runtime capabilities missing: %+v", rec.Capabilities)
	}

	now = now.Add(30 * time.Second)
	if err := src.Heartbeat(context.Background(), RuntimeHeartbeat{RuntimeID: "rt-1", SlotLimit: 4, SlotsInUse: 2, RunningTaskIDs: []string{"task-b", "task-a"}}); err != nil {
		t.Fatalf("Heartbeat: %v", err)
	}
	rec = readFileRuntimeRecord(t, src.runtimePath("rt-1"))
	if !rec.LastHeartbeat.Equal(now) {
		t.Fatalf("heartbeat after update = %v, want %v", rec.LastHeartbeat, now)
	}
	if rec.SlotLimit != 4 || rec.SlotsInUse != 2 {
		t.Fatalf("slot heartbeat mismatch: %+v", rec)
	}
	if len(rec.RunningTaskIDs) != 2 || rec.RunningTaskIDs[0] != "task-a" || rec.RunningTaskIDs[1] != "task-b" {
		t.Fatalf("running task ids should be sorted: %+v", rec.RunningTaskIDs)
	}

	if err := src.DeregisterRuntime(context.Background(), "rt-1"); err != nil {
		t.Fatalf("DeregisterRuntime: %v", err)
	}
	if _, err := os.Stat(src.runtimePath("rt-1")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("runtime record should be removed, stat err=%v", err)
	}
}

func TestFileQueueSourceMissingDir(t *testing.T) {
	_, err := NewFileQueueSource("/nonexistent-path-xyz-9999")
	if err == nil {
		t.Fatal("expected error for missing dir")
	}
}

func readFileRuntimeRecord(t *testing.T, path string) RegisteredRuntime {
	t.Helper()
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read runtime record: %v", err)
	}
	var rec RegisteredRuntime
	if err := json.Unmarshal(body, &rec); err != nil {
		t.Fatalf("decode runtime record: %v", err)
	}
	return rec
}

func countTopLevelJSON(t *testing.T, dir string) int {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	count := 0
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if filepath.Ext(entry.Name()) == ".json" {
			count++
		}
	}
	return count
}

func readClaimRecords(t *testing.T, dir string) []FileClaimRecord {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read claims dir: %v", err)
	}
	records := make([]FileClaimRecord, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		body, err := os.ReadFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			t.Fatal(err)
		}
		var rec FileClaimRecord
		if err := json.Unmarshal(body, &rec); err != nil {
			t.Fatalf("decode claim record %s: %v", entry.Name(), err)
		}
		records = append(records, rec)
	}
	return records
}
