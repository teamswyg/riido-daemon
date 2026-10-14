package providervalidation

import (
	"strings"
	"testing"
)

func TestProviderValidationCodexFullAccessHarness(t *testing.T) {
	ctx := loadProviderValidationContext(t)
	assertCodexFullAccessHarness(t, ctx.providers["codex"], ctx.docText, ctx.securityText, ctx.runtimeText, ctx.migrationText)
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
	assertCodexMustNotClaim(t, row)
	combined := docText + "\n" + securityText + "\n" + runtimeText + "\n" + migrationText
	for _, needle := range []string{
		"Provider full-access/trusted modes are not assumed from provider defaults or\ncaller arguments",
		"daemon-owned full-access runtime selection",
		"Codex adapter 가 danger-full-access envelope 만 생성하고 그 위험을 Riido harness 가\n관리한다",
		"not a provider default, caller-provided default, or\n  hidden fallback",
		"Other providers should follow the same full-access/trusted-runtime\nmeta model only through provider-specific SSOT",
	} {
		if !strings.Contains(combined, needle) {
			t.Fatalf("docs must preserve Codex full-access harness decision %q", needle)
		}
	}
}
