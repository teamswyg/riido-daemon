package providervalidation

import (
	"strings"
	"testing"
)

func TestProviderValidationOpenClawLimits(t *testing.T) {
	ctx := loadProviderValidationContext(t)
	assertOpenClawLimits(t, ctx.providers["openclaw"], ctx.docText, ctx.runtimeText)
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
	assertOpenClawEvidenceLimits(t, row)
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
