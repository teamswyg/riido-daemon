package providervalidation

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestProviderValidationMatrix(t *testing.T) {
	root := filepath.Join("..", "..")
	manifestPath := filepath.Join(root, "docs", "30-architecture", "provider-validation-matrix.riido.json")
	docPath := filepath.Join(root, "docs", "30-architecture", "integration-matrix.md")
	securityDocPath := filepath.Join(root, "docs", "20-domain", "security.md")
	runtimeDocPath := filepath.Join(root, "docs", "20-domain", "provider-runtime.md")
	migrationDocPath := filepath.Join(root, "docs", "migration", "daemon.md")

	manifest := loadManifest(t, manifestPath)
	docText := readText(t, docPath)
	securityText := readText(t, securityDocPath)
	runtimeText := readText(t, runtimeDocPath)
	migrationText := readText(t, migrationDocPath)

	if manifest.SchemaVersion != "riido-daemon-provider-validation-matrix.v1" {
		t.Fatalf("schema_version = %q", manifest.SchemaVersion)
	}
	if manifest.ID != "daemon-provider-validation-matrix" || manifest.RiidoTask != "RIID-4901" {
		t.Fatalf("manifest identity drifted: %+v", manifest)
	}
	if manifest.HumanDoc != "docs/30-architecture/integration-matrix.md" {
		t.Fatalf("human_doc = %q", manifest.HumanDoc)
	}
	for _, want := range []string{
		"docs/20-domain/security.md",
		"docs/20-domain/provider-runtime.md",
		"docs/30-architecture/integration-matrix.md",
	} {
		if !hasString(manifest.SourceDocuments, want) {
			t.Fatalf("source_documents must include %q: %+v", want, manifest.SourceDocuments)
		}
	}
	if !hasString(manifest.GlobalRules, "Provider full-access/trusted runtime modes must be explicit daemon-owned harness envelopes, never implicit provider defaults or caller-provided CustomArgs.") {
		t.Fatalf("global_rules must preserve full-access harness invariant: %+v", manifest.GlobalRules)
	}
	for _, needle := range []string{
		"provider-validation-matrix.riido.json",
		"`PASS`",
		"SaaS completed thread alone is not filesystem side-effect evidence",
		"`supports_worktree=false`",
		"`MISSING_REQUIRED_SURFACE:worktree`",
		"[`security.md`](../20-domain/security.md) §4.3",
	} {
		if !strings.Contains(docText, needle) {
			t.Fatalf("integration matrix doc must mention %q", needle)
		}
	}
	if strings.Count(securityText, "### 4.3 Provider full-access runtime harness") != 1 {
		t.Fatalf("security doc must expose exactly one full-access harness SSOT section")
	}
	if strings.Contains(securityText, "### 4.2 Provider full-access runtime harness") {
		t.Fatalf("security doc must not keep the old duplicate §4.2 full-access heading")
	}
	for _, needle := range []string{
		"Provider full-access runtime harness",
		"default 가 full-access",
		"default sandbox 가\ndanger-full-access",
		"Codex adapter 가 danger-full-access launch\nenvelope 만 생성",
		"codex --sandbox danger-full-access app-server --listen stdio://",
		"daemon 이 Codex 를 전권 host automation",
		"Claude / Cursor / OpenClaw 도 같은 메타 모델",
	} {
		if !strings.Contains(securityText, needle) {
			t.Fatalf("security doc must preserve full-access harness SSOT phrase %q", needle)
		}
	}
	for _, needle := range []string{
		"RIID-4901",
		"provider-validation-matrix.riido.json",
		"`supports_worktree=false`",
		"`required_surfaces=[worktree]`",
		"`MISSING_REQUIRED_SURFACE:worktree`",
	} {
		if !strings.Contains(runtimeText, needle) {
			t.Fatalf("provider-runtime doc must mention %q", needle)
		}
		if !strings.Contains(migrationText, needle) {
			t.Fatalf("daemon migration doc must mention %q", needle)
		}
	}

	providers := map[string]providerEvidence{}
	for _, provider := range manifest.Providers {
		if provider.Provider == "" {
			t.Fatalf("provider row has empty provider: %+v", provider)
		}
		if _, exists := providers[provider.Provider]; exists {
			t.Fatalf("duplicate provider row %q", provider.Provider)
		}
		providers[provider.Provider] = provider
		assertProviderRow(t, provider, docText)
	}
	for _, provider := range []string{"claude", "codex", "openclaw", "cursor"} {
		if _, ok := providers[provider]; !ok {
			t.Fatalf("missing provider row %q", provider)
		}
	}

	assertWorktreeProvider(t, providers["claude"])
	assertWorktreeProvider(t, providers["codex"])
	assertWorktreeProvider(t, providers["cursor"])
	assertCodexFullAccessHarness(t, providers["codex"], docText, securityText, runtimeText, migrationText)
	assertOpenClawLimits(t, providers["openclaw"], docText, runtimeText)
}

func loadManifest(t *testing.T, path string) providerValidationManifest {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read manifest: %v", err)
	}
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	var manifest providerValidationManifest
	if err := dec.Decode(&manifest); err != nil {
		t.Fatalf("decode manifest: %v", err)
	}
	return manifest
}

func readText(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(data)
}

func assertProviderRow(t *testing.T, row providerEvidence, docText string) {
	t.Helper()
	if row.DisplayName == "" || row.RuntimeKind == "" || row.Executable == "" {
		t.Fatalf("provider row missing display/runtime/executable: %+v", row)
	}
	if !strings.Contains(row.OptInIntegration, "AGENTBRIDGE_INTEGRATION=1") ||
		!strings.Contains(row.OptInIntegration, "./internal/provider/"+row.Provider) ||
		!strings.Contains(row.OptInIntegration, "TestIntegration") {
		t.Fatalf("provider %q opt-in integration command is incomplete: %q", row.Provider, row.OptInIntegration)
	}
	if len(row.DeterministicCI) == 0 || len(row.PassEvidence) == 0 || len(row.SkipBeforeRun) == 0 || len(row.MustNotClaim) == 0 {
		t.Fatalf("provider row must include CI/pass/skip/must_not_claim evidence: %+v", row)
	}
	if !strings.Contains(docText, row.DisplayName) || !strings.Contains(docText, row.Executable) {
		t.Fatalf("integration matrix doc must mention provider %q display/executable", row.Provider)
	}
}

func assertWorktreeProvider(t *testing.T, row providerEvidence) {
	t.Helper()
	if row.WorktreeSupport != "supported" {
		t.Fatalf("provider %q worktree_support = %q, want supported", row.Provider, row.WorktreeSupport)
	}
	if !hasString(row.PassEvidence, "ResultCompleted") ||
		!hasString(row.PassEvidence, "expected file artifact inside daemon-selected workdir") {
		t.Fatalf("provider %q must require completed result and daemon-selected workdir artifact: %+v", row.Provider, row.PassEvidence)
	}
	if hasAny(row.MustNotClaim, "SaaS completed thread proves filesystem side effect", "OpenClaw supports daemon-selected worktree") {
		t.Fatalf("provider %q has OpenClaw-only negative claim: %+v", row.Provider, row.MustNotClaim)
	}
}
