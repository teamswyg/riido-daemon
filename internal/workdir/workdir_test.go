package workdir

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/teamswyg/riido-contracts/ir"
)

func TestInjectClaudeWritesCLAUDEmd(t *testing.T) {
	root := t.TempDir()
	a := NewFSAdapter(root)
	ws, err := a.Prepare(TaskID{Workspace: "ws-1", Task: "task-A"})
	if err != nil {
		t.Fatalf("Prepare: %v", err)
	}
	if err := a.InjectRuntimeConfig(ws, RuntimeConfig{
		Provider: "claude",
		Identity: "Agent: tester (id: t-1)",
		CLICatalog: []string{
			"riido task list",
			"riido api status",
		},
		HardRules: []string{
			"Use --output json always.",
		},
		Workflow: "default",
	}); err != nil {
		t.Fatalf("InjectRuntimeConfig: %v", err)
	}

	path := filepath.Join(ws.Workdir, "CLAUDE.md")
	bytes, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read CLAUDE.md: %v", err)
	}
	content := string(bytes)
	for _, want := range []string{
		"Agent: tester (id: t-1)",
		"riido task list",
		"Use --output json always.",
		"workflow: default",
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("CLAUDE.md missing %q:\n%s", want, content)
		}
	}
}

func TestInjectCodexWritesAGENTSmd(t *testing.T) {
	root := t.TempDir()
	a := NewFSAdapter(root)
	ws, _ := a.Prepare(TaskID{Workspace: "ws-1", Task: "task-B"})
	if err := a.InjectRuntimeConfig(ws, RuntimeConfig{
		Provider:                   "codex",
		ProtocolKind:               "codex-app-server",
		TelemetryContractPlacement: "prompt",
		Identity:                   "id",
		Workflow:                   "quick-create",
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(ws.Workdir, "AGENTS.md")); err != nil {
		t.Fatalf("AGENTS.md missing: %v", err)
	}
	if _, err := os.Stat(filepath.Join(ws.NativeConfig, "AGENTS.md")); err != nil {
		t.Fatalf("native-config AGENTS.md copy missing: %v", err)
	}
	if _, err := os.Stat(filepath.Join(ws.Workdir, "CLAUDE.md")); err == nil {
		t.Fatalf("codex must not create CLAUDE.md")
	}

	manifest := readNativeConfigManifest(t, filepath.Join(ws.Workdir, NativeConfigManifestPath))
	if manifest.SchemaVersion != NativeConfigManifestSchemaVersion {
		t.Fatalf("manifest schema = %q", manifest.SchemaVersion)
	}
	if manifest.ProviderKind != "codex" ||
		manifest.ProtocolKind != "codex-app-server" ||
		manifest.PrimaryInstructionFile != "AGENTS.md" ||
		manifest.ManifestFile != NativeConfigManifestPath ||
		manifest.HookMode != NativeConfigHookModeInstructionOnly ||
		manifest.ConfigHomeDir != "" ||
		manifest.TelemetryContractPlacement != "prompt" ||
		manifest.Workflow != "quick-create" {
		t.Fatalf("manifest = %+v", manifest)
	}
	if len(manifest.ProviderSettingsFiles) != 0 {
		t.Fatalf("manifest provider settings files = %+v", manifest.ProviderSettingsFiles)
	}
	for _, want := range []string{"AGENTS.md", NativeConfigManifestPath} {
		if !containsString(manifest.GeneratedFiles, want) {
			t.Fatalf("manifest generated files missing %q: %+v", want, manifest.GeneratedFiles)
		}
		if _, err := os.Stat(filepath.Join(ws.NativeConfig, filepath.FromSlash(want))); err != nil {
			t.Fatalf("native-config copy missing %s: %v", want, err)
		}
	}
}

func TestInjectCodexCanApplyConfigHomePolicy(t *testing.T) {
	root := t.TempDir()
	a := NewFSAdapter(root)
	ws, _ := a.Prepare(TaskID{Workspace: "ws-1", Task: "task-codex-no-home"})
	if err := a.InjectRuntimeConfig(ws, RuntimeConfig{
		Provider:             "codex",
		ProtocolKind:         "codex-app-server",
		NativeConfigHomeMode: NativeConfigHomeModeDisabled,
	}); err != nil {
		t.Fatal(err)
	}

	manifest := readNativeConfigManifest(t, filepath.Join(ws.Workdir, NativeConfigManifestPath))
	if manifest.ProviderKind != "codex" || manifest.ConfigHomeDir != "" {
		t.Fatalf("manifest = %+v", manifest)
	}
	for _, blocked := range []string{".codex/config.toml"} {
		if containsString(manifest.ProviderSettingsFiles, blocked) || containsString(manifest.GeneratedFiles, blocked) {
			t.Fatalf("manifest must not include blocked config home artifact %q: %+v", blocked, manifest)
		}
		if _, err := os.Stat(filepath.Join(ws.Workdir, filepath.FromSlash(blocked))); !os.IsNotExist(err) {
			t.Fatalf("workdir config home artifact %s should be absent, stat err=%v", blocked, err)
		}
		if _, err := os.Stat(filepath.Join(ws.NativeConfig, filepath.FromSlash(blocked))); !os.IsNotExist(err) {
			t.Fatalf("native-config config home artifact %s should be absent, stat err=%v", blocked, err)
		}
	}
	if _, err := os.Stat(filepath.Join(ws.Workdir, "AGENTS.md")); err != nil {
		t.Fatalf("primary instruction file should remain: %v", err)
	}
}

func TestInjectOpenClawAndCursorRemainInstructionOnly(t *testing.T) {
	for _, provider := range []string{"openclaw", "cursor"} {
		t.Run(provider, func(t *testing.T) {
			root := t.TempDir()
			a := NewFSAdapter(root)
			ws, err := a.Prepare(TaskID{Workspace: "ws-1", Task: "task-" + provider})
			if err != nil {
				t.Fatal(err)
			}
			if err := a.InjectRuntimeConfig(ws, RuntimeConfig{
				Provider:                   provider,
				ProtocolKind:               provider + "-protocol",
				TelemetryContractPlacement: "prompt",
				Identity:                   "id",
				Workflow:                   "default",
			}); err != nil {
				t.Fatal(err)
			}

			manifest := readNativeConfigManifest(t, filepath.Join(ws.Workdir, NativeConfigManifestPath))
			if manifest.ProviderKind != provider ||
				manifest.ProtocolKind != provider+"-protocol" ||
				manifest.PrimaryInstructionFile != "AGENTS.md" ||
				manifest.ManifestFile != NativeConfigManifestPath ||
				manifest.HookMode != NativeConfigHookModeInstructionOnly ||
				manifest.ConfigHomeDir != "" ||
				manifest.TelemetryContractPlacement != "prompt" ||
				manifest.Workflow != "default" ||
				len(manifest.ProviderSettingsFiles) != 0 ||
				len(manifest.HookFiles) != 0 {
				t.Fatalf("manifest = %+v", manifest)
			}
			if len(manifest.GeneratedFiles) != 2 ||
				!containsString(manifest.GeneratedFiles, "AGENTS.md") ||
				!containsString(manifest.GeneratedFiles, NativeConfigManifestPath) {
				t.Fatalf("generated files = %+v", manifest.GeneratedFiles)
			}
			for _, want := range manifest.GeneratedFiles {
				if _, err := os.Stat(filepath.Join(ws.Workdir, filepath.FromSlash(want))); err != nil {
					t.Fatalf("workdir generated file %s missing: %v", want, err)
				}
				if _, err := os.Stat(filepath.Join(ws.NativeConfig, filepath.FromSlash(want))); err != nil {
					t.Fatalf("native-config generated file %s missing: %v", want, err)
				}
			}
			for _, blocked := range []string{
				".cursor/settings.json",
				".cursor/rules",
				".openclaw/settings.json",
				".openclaw/config.json",
			} {
				if _, err := os.Stat(filepath.Join(ws.Workdir, filepath.FromSlash(blocked))); !os.IsNotExist(err) {
					t.Fatalf("provider-native artifact %s should be absent from workdir, stat err=%v", blocked, err)
				}
				if _, err := os.Stat(filepath.Join(ws.NativeConfig, filepath.FromSlash(blocked))); !os.IsNotExist(err) {
					t.Fatalf("provider-native artifact %s should be absent from native-config, stat err=%v", blocked, err)
				}
			}
		})
	}
}

func TestInjectClaudeWritesSettingsAndHook(t *testing.T) {
	root := t.TempDir()
	a := NewFSAdapter(root)
	ws, err := a.Prepare(TaskID{Workspace: "ws-1", Task: "task-claude"})
	if err != nil {
		t.Fatal(err)
	}
	if err := a.InjectRuntimeConfig(ws, RuntimeConfig{Provider: "claude", ProtocolKind: "claude-stream-json"}); err != nil {
		t.Fatal(err)
	}
	manifest := readNativeConfigManifest(t, filepath.Join(ws.Workdir, NativeConfigManifestPath))
	if manifest.HookMode != NativeConfigHookModeClaudeCommandHooks {
		t.Fatalf("hook mode = %q", manifest.HookMode)
	}
	for _, want := range []string{".claude/settings.json", ".riido/hooks/claude-audit-hook.sh"} {
		if !containsString(manifest.GeneratedFiles, want) {
			t.Fatalf("manifest generated files missing %q: %+v", want, manifest.GeneratedFiles)
		}
	}
	if !containsString(manifest.ProviderSettingsFiles, ".claude/settings.json") {
		t.Fatalf("provider settings files = %+v", manifest.ProviderSettingsFiles)
	}
	if !containsString(manifest.HookFiles, ".riido/hooks/claude-audit-hook.sh") {
		t.Fatalf("hook files = %+v", manifest.HookFiles)
	}
	settings, err := os.ReadFile(filepath.Join(ws.Workdir, ".claude", "settings.json"))
	if err != nil {
		t.Fatalf("read claude settings: %v", err)
	}
	if !strings.Contains(string(settings), `"PreToolUse"`) ||
		!strings.Contains(string(settings), `"PostToolUse"`) ||
		!strings.Contains(string(settings), "${CLAUDE_PROJECT_DIR}/.riido/hooks/claude-audit-hook.sh") {
		t.Fatalf("claude settings missing hook config:\n%s", settings)
	}
	hookPath := filepath.Join(ws.Workdir, ".riido", "hooks", "claude-audit-hook.sh")
	info, err := os.Stat(hookPath)
	if err != nil {
		t.Fatalf("hook script missing: %v", err)
	}
	if info.Mode().Perm() != 0o755 {
		t.Fatalf("hook script mode = %v", info.Mode().Perm())
	}
	if _, err := os.Stat(filepath.Join(ws.NativeConfig, ".claude", "settings.json")); err != nil {
		t.Fatalf("native config claude settings copy missing: %v", err)
	}
	if _, err := os.Stat(filepath.Join(ws.NativeConfig, ".riido", "hooks", "claude-audit-hook.sh")); err != nil {
		t.Fatalf("native config hook copy missing: %v", err)
	}
}

func TestInjectClaudeCanApplyInstructionOnlyHookPolicy(t *testing.T) {
	root := t.TempDir()
	a := NewFSAdapter(root)
	ws, err := a.Prepare(TaskID{Workspace: "ws-1", Task: "task-claude-no-hooks"})
	if err != nil {
		t.Fatal(err)
	}
	if err := a.InjectRuntimeConfig(ws, RuntimeConfig{
		Provider:       "claude",
		ProtocolKind:   "claude-stream-json",
		NativeHookMode: NativeConfigHookModeInstructionOnly,
	}); err != nil {
		t.Fatal(err)
	}
	manifest := readNativeConfigManifest(t, filepath.Join(ws.Workdir, NativeConfigManifestPath))
	if manifest.HookMode != NativeConfigHookModeInstructionOnly {
		t.Fatalf("hook mode = %q", manifest.HookMode)
	}
	for _, blocked := range []string{".claude/settings.json", ".riido/hooks/claude-audit-hook.sh"} {
		if containsString(manifest.GeneratedFiles, blocked) {
			t.Fatalf("manifest generated files must not include blocked hook artifact %q: %+v", blocked, manifest.GeneratedFiles)
		}
		if _, err := os.Stat(filepath.Join(ws.Workdir, filepath.FromSlash(blocked))); !os.IsNotExist(err) {
			t.Fatalf("workdir hook artifact %s should be absent, stat err=%v", blocked, err)
		}
		if _, err := os.Stat(filepath.Join(ws.NativeConfig, filepath.FromSlash(blocked))); !os.IsNotExist(err) {
			t.Fatalf("native-config hook artifact %s should be absent, stat err=%v", blocked, err)
		}
	}
	if _, err := os.Stat(filepath.Join(ws.Workdir, "CLAUDE.md")); err != nil {
		t.Fatalf("primary instruction file should remain: %v", err)
	}
}

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

func TestProviderConfigFilenameRegistry(t *testing.T) {
	for _, tc := range []struct {
		provider string
		want     string
	}{
		{"claude", "CLAUDE.md"},
		{"codex", "AGENTS.md"},
		{"openclaw", "AGENTS.md"},
		{"cursor", "AGENTS.md"},
	} {
		got := ProviderConfigFilename(tc.provider)
		if got != tc.want {
			t.Fatalf("%s: want %q, got %q", tc.provider, tc.want, got)
		}
		plan := ProviderConfigPlan(tc.provider)
		if plan.PrimaryInstructionFile != tc.want ||
			plan.ManifestFile != NativeConfigManifestPath ||
			plan.HookMode == "" {
			t.Fatalf("%s plan = %+v", tc.provider, plan)
		}
	}
	if got := ProviderConfigFilename("unknown"); got != "AGENTS.md" {
		t.Fatalf("unknown provider should fall back to AGENTS.md, got %q", got)
	}
}

func TestInjectRefusesPathTraversal(t *testing.T) {
	root := t.TempDir()
	a := NewFSAdapter(root)
	ws, _ := a.Prepare(TaskID{Workspace: "ws-1", Task: "task-Z"})
	// A malicious provider name with path traversal must NOT escape the workdir.
	err := a.InjectRuntimeConfig(ws, RuntimeConfig{Provider: "../etc"})
	if err == nil {
		t.Fatalf("expected error for path-traversal provider")
	}
}

func readNativeConfigManifest(t *testing.T, path string) NativeConfigManifest {
	t.Helper()
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read native config manifest: %v", err)
	}
	var manifest NativeConfigManifest
	if err := json.Unmarshal(body, &manifest); err != nil {
		t.Fatalf("decode native config manifest: %v", err)
	}
	return manifest
}

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func TestInjectRendersWorkdirGuidanceWhenSet(t *testing.T) {
	root := t.TempDir()
	a := NewFSAdapter(root)

	ws, _ := a.Prepare(TaskID{Workspace: "ws-1", Task: "task-guidance"})
	if err := a.InjectRuntimeConfig(ws, RuntimeConfig{
		Provider:        "codex",
		WorkdirGuidance: "The working directory `/x/workdir` is empty: it has no source repository.",
		Workflow:        "default",
	}); err != nil {
		t.Fatalf("InjectRuntimeConfig: %v", err)
	}
	content, err := os.ReadFile(filepath.Join(ws.Workdir, "AGENTS.md"))
	if err != nil {
		t.Fatalf("read AGENTS.md: %v", err)
	}
	for _, want := range []string{"## Working directory", "no source repository"} {
		if !strings.Contains(string(content), want) {
			t.Fatalf("AGENTS.md missing %q:\n%s", want, string(content))
		}
	}

	// Absent guidance -> no section.
	ws2, _ := a.Prepare(TaskID{Workspace: "ws-1", Task: "task-none"})
	if err := a.InjectRuntimeConfig(ws2, RuntimeConfig{Provider: "codex", Workflow: "default"}); err != nil {
		t.Fatalf("InjectRuntimeConfig: %v", err)
	}
	c2, _ := os.ReadFile(filepath.Join(ws2.Workdir, "AGENTS.md"))
	if strings.Contains(string(c2), "## Working directory") {
		t.Fatalf("Working directory section should be absent when guidance is empty:\n%s", string(c2))
	}
}
