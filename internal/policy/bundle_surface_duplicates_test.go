package policy

import (
	"strings"
	"testing"
)

func TestParsePolicyBundleRejectsDuplicateNativeConfigFile(t *testing.T) {
	errText := parseBundleSurfaceError(t, "Host", `"native_config_files": [
		"codex:config-home:task-scoped",
		"codex:config-home:task-scoped"
	]`)
	if !strings.Contains(errText, "duplicate native config file surface") {
		t.Fatalf("expected duplicate native config file rejection, got %s", errText)
	}
}

func TestParsePolicyBundleRejectsDuplicateToolUseSurface(t *testing.T) {
	errText := parseBundleSurfaceError(t, "Host", `"tool_use": [
		"tool:network-egress",
		"tool:network-egress"
	]`)
	if !strings.Contains(errText, "duplicate tool use surface") {
		t.Fatalf("expected duplicate tool use rejection, got %s", errText)
	}
}
