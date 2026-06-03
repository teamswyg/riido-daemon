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
	runtimeDocPath := filepath.Join(root, "docs", "20-domain", "provider-runtime.md")
	migrationDocPath := filepath.Join(root, "docs", "migration", "daemon.md")

	manifest := loadManifest(t, manifestPath)
	docText := readText(t, docPath)
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
	for _, needle := range []string{
		"provider-validation-matrix.riido.json",
		"`PASS`",
		"SaaS completed thread alone is not filesystem side-effect evidence",
		"`supports_worktree=false`",
		"`MISSING_REQUIRED_SURFACE:worktree`",
	} {
		if !strings.Contains(docText, needle) {
			t.Fatalf("integration matrix doc must mention %q", needle)
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

func assertOpenClawLimits(t *testing.T, row providerEvidence, docText string, runtimeText string) {
	t.Helper()
	if row.WorktreeSupport != "unsupported" {
		t.Fatalf("OpenClaw worktree_support = %q, want unsupported", row.WorktreeSupport)
	}
	for _, needle := range []string{
		"ResultCompleted with non-empty provider output",
		"deterministic provider-safe session id",
		"executable path that passed OpenClaw Detect",
	} {
		if !hasString(row.PassEvidence, needle) {
			t.Fatalf("OpenClaw pass evidence missing %q: %+v", needle, row.PassEvidence)
		}
	}
	for _, needle := range []string{
		"SaaS completed thread alone does not prove filesystem side effect",
		"Runtime capability still reports supports_worktree=false",
	} {
		if !hasString(row.NegativeOrLimitedEvidence, needle) {
			t.Fatalf("OpenClaw limited evidence missing %q: %+v", needle, row.NegativeOrLimitedEvidence)
		}
	}
	if !hasString(row.RequiredSchedulingGate, "required_surfaces=[worktree] -> MISSING_REQUIRED_SURFACE:worktree") {
		t.Fatalf("OpenClaw scheduling gate missing: %+v", row.RequiredSchedulingGate)
	}
	for _, needle := range []string{
		"OpenClaw supports daemon-selected worktree",
		"SaaS completed thread proves filesystem side effect",
		"OpenClaw text completion is enough for worktree-required tasks",
	} {
		if !hasString(row.MustNotClaim, needle) {
			t.Fatalf("OpenClaw must_not_claim missing %q: %+v", needle, row.MustNotClaim)
		}
	}
	for _, needle := range []string{
		"`supports_worktree=false`",
		"`MISSING_REQUIRED_SURFACE:worktree`",
		"SaaS completion alone must not be treated as filesystem side-effect evidence",
	} {
		if !strings.Contains(docText+"\n"+runtimeText, needle) {
			t.Fatalf("docs must preserve OpenClaw limitation %q", needle)
		}
	}
}

func hasString(items []string, want string) bool {
	for _, item := range items {
		if item == want {
			return true
		}
	}
	return false
}

func hasAny(items []string, wants ...string) bool {
	for _, want := range wants {
		if hasString(items, want) {
			return true
		}
	}
	return false
}

type providerValidationManifest struct {
	SchemaVersion   string             `json:"schema_version"`
	ID              string             `json:"id"`
	RiidoTask       string             `json:"riido_task"`
	HumanDoc        string             `json:"human_doc"`
	SourceDocuments []string           `json:"source_documents"`
	GlobalRules     []string           `json:"global_rules"`
	Providers       []providerEvidence `json:"providers"`
}

type providerEvidence struct {
	Provider                  string   `json:"provider"`
	DisplayName               string   `json:"display_name"`
	RuntimeKind               string   `json:"runtime_kind"`
	Executable                string   `json:"executable"`
	DeterministicCI           []string `json:"deterministic_ci"`
	OptInIntegration          string   `json:"opt_in_integration"`
	WorktreeSupport           string   `json:"worktree_support"`
	PassEvidence              []string `json:"pass_evidence"`
	NegativeOrLimitedEvidence []string `json:"negative_or_limited_evidence,omitempty"`
	RequiredSchedulingGate    []string `json:"required_scheduling_gate,omitempty"`
	SkipBeforeRun             []string `json:"skip_before_run"`
	LatestEvidence            []string `json:"latest_evidence"`
	MustNotClaim              []string `json:"must_not_claim"`
}
