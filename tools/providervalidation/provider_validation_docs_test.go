package providervalidation

import (
	"strings"
	"testing"
)

func TestProviderValidationIntegrationDocMentionsMatrixEvidence(t *testing.T) {
	docText := loadProviderValidationContext(t).docText
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
}

func TestProviderValidationSecurityDocKeepsFullAccessHarnessSSOT(t *testing.T) {
	securityText := loadProviderValidationContext(t).securityText
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
}
