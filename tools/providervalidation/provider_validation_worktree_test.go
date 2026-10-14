package providervalidation

import "testing"

func TestProviderValidationWorktreeProviders(t *testing.T) {
	providers := loadProviderValidationContext(t).providers
	assertWorktreeProvider(t, providers["claude"])
	assertWorktreeProvider(t, providers["codex"])
	assertWorktreeProvider(t, providers["cursor"])
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
