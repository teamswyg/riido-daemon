package workdir

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestProviderConfigPlanUsesGeneratedCatalog(t *testing.T) {
	if NativeConfigPlanSchemaVersion != "riido-native-config-plan.v1" {
		t.Fatalf("native config plan schema = %q", NativeConfigPlanSchemaVersion)
	}
	claude := ProviderConfigPlan(" Claude ")
	if claude.ProviderKind != "claude" ||
		claude.PrimaryInstructionFile != "CLAUDE.md" ||
		claude.HookMode != NativeConfigHookModeClaudeCommandHooks ||
		!containsString(claude.ProviderSettingsFiles, ".claude/settings.json") ||
		!containsString(claude.HookFiles, ".riido/hooks/claude-audit-hook.sh") {
		t.Fatalf("claude plan = %+v", claude)
	}
	codex := ProviderConfigPlan("codex")
	if codex.ConfigHomeDir != "" || len(codex.ProviderSettingsFiles) != 0 {
		t.Fatalf("codex plan = %+v", codex)
	}
	cursor := ProviderConfigPlan("cursor")
	if cursor.ProviderKind != "cursor" ||
		cursor.PrimaryInstructionFile != "AGENTS.md" ||
		cursor.HookMode != NativeConfigHookModeInstructionOnly ||
		cursor.ConfigHomeDir != "" ||
		len(cursor.ProviderSettingsFiles) != 0 ||
		len(cursor.HookFiles) != 0 {
		t.Fatalf("cursor plan = %+v", cursor)
	}
	openclaw := ProviderConfigPlan("openclaw")
	if openclaw.ProviderKind != "openclaw" ||
		openclaw.PrimaryInstructionFile != "AGENTS.md" ||
		openclaw.HookMode != NativeConfigHookModeInstructionOnly ||
		openclaw.ConfigHomeDir != "" ||
		len(openclaw.ProviderSettingsFiles) != 0 ||
		len(openclaw.HookFiles) != 0 {
		t.Fatalf("openclaw plan = %+v", openclaw)
	}
	unknown := ProviderConfigPlan("unknown-provider")
	if unknown.ProviderKind != "unknown-provider" || unknown.PrimaryInstructionFile != "AGENTS.md" || unknown.ManifestFile != NativeConfigManifestPath {
		t.Fatalf("unknown plan = %+v", unknown)
	}
}

func TestPrepareCreatesTreeAndEnforcesWorkspaceID(t *testing.T) {
	root := t.TempDir()
	a := NewFSAdapter(root)
	// Empty workspace must be rejected (spec §6.1 "workspace_id 필수").
	_, err := a.Prepare(TaskID{Task: "task-1"})
	if err == nil {
		t.Fatal("expected error for empty workspace id")
	}
	ws, err := a.Prepare(TaskID{Workspace: "ws-1", Task: "task-1", Run: "run-1"})
	if err != nil {
		t.Fatal(err)
	}
	wantRoot := filepath.Join(root, "ws-1", "tasks", "task-1", "runs", "run-1")
	if ws.Root != wantRoot {
		t.Fatalf("workspace root = %q, want %q", ws.Root, wantRoot)
	}
	// Each SSOT layout directory must exist.
	for _, sub := range []string{"workdir", "output", "logs", "artifacts", "native-config", "ir"} {
		info, err := os.Stat(filepath.Join(ws.Root, sub))
		if err != nil {
			t.Fatalf("expected %s subdir: %v", sub, err)
		}
		if !info.IsDir() {
			t.Fatalf("%s should be a directory", sub)
		}
	}
	// .gc_meta.json must include workspace_id + task_id.
	meta, err := os.ReadFile(filepath.Join(ws.Root, ".gc_meta.json"))
	if err != nil {
		t.Fatalf("missing gc meta: %v", err)
	}
	for _, want := range []string{`"workspace_id":"ws-1"`, `"task_id":"task-1"`, `"run_id":"run-1"`} {
		if !strings.Contains(string(meta), want) {
			t.Fatalf("gc meta missing %q:\n%s", want, meta)
		}
	}
}

func TestArchiveWritesKeepInPlaceManifest(t *testing.T) {
	root := t.TempDir()
	a := NewFSAdapter(root)
	ws, err := a.Prepare(TaskID{Workspace: "ws-1", Task: "task-1", Run: "run-1"})
	if err != nil {
		t.Fatal(err)
	}
	archivedAt := time.Date(2026, 5, 24, 1, 2, 3, 4, time.UTC)
	record, err := a.Archive(ws, ArchiveRequest{
		ResultStatus: "completed",
		ArchivedAt:   archivedAt,
	})
	if err != nil {
		t.Fatalf("Archive: %v", err)
	}
	if record.SchemaVersion != ArchiveRecordSchemaVersion {
		t.Fatalf("schema version = %q", record.SchemaVersion)
	}
	if record.RetentionMode != RetentionModeKeepInPlace {
		t.Fatalf("retention mode = %q", record.RetentionMode)
	}
	if record.WorkdirPath != ws.Workdir {
		t.Fatalf("workdir path = %q, want %q", record.WorkdirPath, ws.Workdir)
	}
	if !strings.HasPrefix(record.ArchiveURI, "file://") {
		t.Fatalf("archive uri = %q", record.ArchiveURI)
	}

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
