package workdir

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/teamswyg/riido-contracts/ir"
)

func TestCleanupArchivedBeforeRemovesOnlyExpiredArchivedRuns(t *testing.T) {
	root := t.TempDir()
	a := NewFSAdapter(root)
	oldRun, err := a.Prepare(TaskID{Workspace: "ws-1", Task: "task-old", Run: "run-old"})
	if err != nil {
		t.Fatal(err)
	}
	freshRun, err := a.Prepare(TaskID{Workspace: "ws-1", Task: "task-fresh", Run: "run-fresh"})
	if err != nil {
		t.Fatal(err)
	}
	activeRun, err := a.Prepare(TaskID{Workspace: "ws-1", Task: "task-active", Run: "run-active"})
	if err != nil {
		t.Fatal(err)
	}

	cutoff := time.Date(2026, 5, 25, 12, 0, 0, 0, time.UTC)
	if _, err := a.Archive(oldRun, ArchiveRequest{
		ResultStatus: "completed",
		ArchivedAt:   cutoff.Add(-time.Hour),
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := a.Archive(freshRun, ArchiveRequest{
		ResultStatus: "failed",
		ArchivedAt:   cutoff.Add(time.Hour),
	}); err != nil {
		t.Fatal(err)
	}

	result, err := a.CleanupArchivedBefore(context.Background(), CleanupRequest{
		ArchivedBefore: cutoff,
		RemovedAt:      cutoff.Add(2 * time.Hour),
	})
	if err != nil {
		t.Fatalf("CleanupArchivedBefore: %v", err)
	}
	if result.ScannedArchiveRecords != 2 || len(result.Removed) != 1 {
		t.Fatalf("cleanup result = %+v", result)
	}
	if result.Removed[0].RunRoot != oldRun.Root || result.Removed[0].Archive.ResultStatus != "completed" {
		t.Fatalf("removed record = %+v", result.Removed[0])
	}
	if _, err := os.Stat(oldRun.Root); !os.IsNotExist(err) {
		t.Fatalf("old archived run should be removed, stat err=%v", err)
	}
	for _, keep := range []string{freshRun.Root, activeRun.Root} {
		if info, err := os.Stat(keep); err != nil || !info.IsDir() {
			t.Fatalf("run should remain %s: info=%+v err=%v", keep, info, err)
		}
	}
}

func TestCleanupArchivedBeforeRequiresCutoff(t *testing.T) {
	_, err := NewFSAdapter(t.TempDir()).CleanupArchivedBefore(context.Background(), CleanupRequest{})
	if err == nil {
		t.Fatal("expected error for empty cleanup cutoff")
	}
}

func TestComputeNativeConfigVersionIsDeterministicAndPolicyBound(t *testing.T) {
	root := t.TempDir()
	a := NewFSAdapter(root)
	ws, err := a.Prepare(TaskID{Workspace: "ws-1", Task: "task-1", Run: "run-1"})
	if err != nil {
		t.Fatal(err)
	}
	if err := a.InjectRuntimeConfig(ws, RuntimeConfig{
		Provider: "codex",
		Identity: "Agent: tester",
	}); err != nil {
		t.Fatal(err)
	}
	input := NativeConfigVersionInput{
		PolicyBundleVersion: "policy-bundle.test.v1",
		ProviderKind:        "codex",
		ProtocolKind:        "codex-app-server",
	}
	first, err := ComputeNativeConfigVersion(ws, input)
	if err != nil {
		t.Fatalf("ComputeNativeConfigVersion: %v", err)
	}
	second, err := ComputeNativeConfigVersion(ws, input)
	if err != nil {
		t.Fatal(err)
	}
	if first == "" || first != second {
		t.Fatalf("version should be deterministic: first=%q second=%q", first, second)
	}

	changedPolicy := input
	changedPolicy.PolicyBundleVersion = "policy-bundle.test.v2"
	policyVersion, err := ComputeNativeConfigVersion(ws, changedPolicy)
	if err != nil {
		t.Fatal(err)
	}
	if policyVersion == first {
		t.Fatal("version must change when policy bundle changes")
	}

	path := filepath.Join(ws.NativeConfig, "AGENTS.md")
	if err := os.WriteFile(path, []byte("changed"), 0o644); err != nil {
		t.Fatal(err)
	}
	changedContent, err := ComputeNativeConfigVersion(ws, input)
	if err != nil {
		t.Fatal(err)
	}
	if changedContent == first {
		t.Fatal("version must change when injected file content changes")
	}
}

func TestRunEventSinkAppendsJSONL(t *testing.T) {
	root := t.TempDir()
	a := NewFSAdapter(root)
	ws, err := a.Prepare(TaskID{Workspace: "ws-1", Task: "task-1", Run: "run-1"})
	if err != nil {
		t.Fatal(err)
	}
	sink, err := NewRunEventSink(ws)
	if err != nil {
		t.Fatal(err)
	}
	ev := ir.CanonicalEvent{
		EventID:             "event-1",
		OccurredAt:          time.Date(2026, 5, 25, 10, 0, 0, 0, time.UTC),
		EventSchemaVersion:  1,
		Scope:               ir.EventScopeTask,
		Type:                ir.EventTaskCreated,
		ActorKind:           ir.ActorDaemon,
		ActorID:             "daemon-1",
		RiidoDaemonVersion:  "riido-daemon.test.v1",
		PolicyBundleVersion: "policy-bundle.test.v1",
		TaskID:              "task-1",
		FSMVersion:          1,
	}
	ev2 := ev
	ev2.EventID = "event-2"
	if err := sink.AppendEvents(context.Background(), []ir.CanonicalEvent{ev, ev2}); err != nil {
		t.Fatalf("AppendEvents: %v", err)
	}
	body, err := os.ReadFile(sink.Path())
	if err != nil {
		t.Fatalf("read event log: %v", err)
	}
	dec := json.NewDecoder(bytes.NewReader(body))
	count := 0
	for {
		var got ir.CanonicalEvent
		err := dec.Decode(&got)
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		wantID := fmt.Sprintf("event-%d", count+1)
		if got.EventID != wantID || got.Type != ir.EventTaskCreated {
			t.Fatalf("event mismatch: %+v", got)
		}
		count++
	}
	if count != 2 {
		t.Fatalf("event count = %d, want 2", count)
	}
}
