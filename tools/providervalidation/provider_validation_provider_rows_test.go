package providervalidation

import (
	"strings"
	"testing"
)

func TestProviderValidationProviderRows(t *testing.T) {
	ctx := loadProviderValidationContext(t)
	for _, provider := range ctx.manifest.Providers {
		assertProviderRow(t, provider, ctx.docText)
	}
	for _, provider := range []string{"claude", "codex", "openclaw", "cursor"} {
		if _, ok := ctx.providers[provider]; !ok {
			t.Fatalf("missing provider row %q", provider)
		}
	}
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
