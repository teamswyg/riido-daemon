package policy

import (
	"strings"
	"testing"
)

func TestParsePolicyBundleRejectsUnknownTierNativeConfigHook(t *testing.T) {
	errText := parseBundleSurfaceError(t, "Unknown", `"native_config_hooks": ["claude:command-hooks:audit"]`)
	if !strings.Contains(errText, "Unknown") {
		t.Fatalf("expected Unknown native config hook rejection, got %s", errText)
	}
}

func TestParsePolicyBundleRejectsUnknownTierNativeConfigFile(t *testing.T) {
	errText := parseBundleSurfaceError(t, "Unknown", `"native_config_files": ["codex:config-home:task-scoped"]`)
	if !strings.Contains(errText, "Unknown") {
		t.Fatalf("expected Unknown native config file rejection, got %s", errText)
	}
}

func TestParsePolicyBundleRejectsUnknownTierToolUse(t *testing.T) {
	errText := parseBundleSurfaceError(t, "Unknown", `"tool_use": ["tool:network-egress"]`)
	if !strings.Contains(errText, "Unknown") {
		t.Fatalf("expected Unknown tool use rejection, got %s", errText)
	}
}
