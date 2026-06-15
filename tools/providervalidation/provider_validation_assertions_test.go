package providervalidation

import (
	"slices"
	"strings"
	"testing"
)

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
