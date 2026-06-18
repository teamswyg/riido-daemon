package providervalidation

import "testing"

func assertCodexMustNotClaim(t *testing.T, row providerEvidence) {
	t.Helper()
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
}
