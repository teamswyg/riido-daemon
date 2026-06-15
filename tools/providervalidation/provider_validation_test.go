package providervalidation

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
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

func assertCodexFullAccessHarness(t *testing.T, row providerEvidence, docText, securityText, runtimeText, migrationText string) {
	t.Helper()
	if row.Provider != "codex" {
		t.Fatalf("Codex harness assertion called with provider %q", row.Provider)
	}
	for _, needle := range []string{
		"explicit daemon-owned codex --sandbox danger-full-access app-server --listen stdio:// launch shape",
		"caller sandbox/config/unsafe-bypass args are dropped with DroppedArgs evidence",
		"expected file artifact inside daemon-selected workdir",
	} {
		if !hasString(row.PassEvidence, needle) {
			t.Fatalf("Codex pass evidence missing %q: %+v", needle, row.PassEvidence)
		}
	}
	for _, needle := range []string{
		"Codex full-access came from provider default sandbox selection",
		"Codex sandbox selection came from caller CustomArgs or SaaS payload",
		"Codex task-scoped permission profile is active",
	} {
		if !hasString(row.MustNotClaim, needle) {
			t.Fatalf("Codex must_not_claim missing %q: %+v", needle, row.MustNotClaim)
		}
	}
	if !hasString(row.LatestEvidence, "RIID-4917-Codex-full-access-harness-policy") {
		t.Fatalf("Codex latest_evidence must include RIID-4917 harness policy: %+v", row.LatestEvidence)
	}
	for _, needle := range []string{
		"Provider full-access/trusted modes are not assumed from provider defaults or\ncaller arguments",
		"daemon-owned full-access runtime selection",
		"Codex adapter 가 danger-full-access envelope 만 생성하고 그 위험을 Riido harness 가\n관리한다",
		"not a provider default, caller-provided default, or\n  hidden fallback",
		"Other providers should follow the same full-access/trusted-runtime\nmeta model only through provider-specific SSOT",
	} {
		if !strings.Contains(docText+"\n"+securityText+"\n"+runtimeText+"\n"+migrationText, needle) {
			t.Fatalf("docs must preserve Codex full-access harness decision %q", needle)
		}
	}
}

func assertOpenClawLimits(t *testing.T, row providerEvidence, docText, runtimeText string) {
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
	return slices.Contains(items, want)
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
