package providervalidation

import (
	"strings"
	"testing"
)

func TestProviderValidationRuntimeAndMigrationDocsMentionSchedulingGate(t *testing.T) {
	ctx := loadProviderValidationContext(t)
	for _, needle := range []string{
		"RIID-4901",
		"provider-validation-matrix.riido.json",
		"`supports_worktree=false`",
		"`required_surfaces=[worktree]`",
		"`MISSING_REQUIRED_SURFACE:worktree`",
	} {
		if !strings.Contains(ctx.runtimeText, needle) {
			t.Fatalf("provider-runtime doc must mention %q", needle)
		}
		if !strings.Contains(ctx.migrationText, needle) {
			t.Fatalf("daemon migration doc must mention %q", needle)
		}
	}
}
