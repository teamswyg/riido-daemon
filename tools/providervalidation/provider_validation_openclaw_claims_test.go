package providervalidation

import "testing"

func assertOpenClawEvidenceLimits(t *testing.T, row providerEvidence) {
	t.Helper()
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
}
