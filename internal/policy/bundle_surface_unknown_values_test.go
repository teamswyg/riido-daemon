package policy

import (
	"strings"
	"testing"
)

func TestParsePolicyBundleRejectsUnknownSurface(t *testing.T) {
	errText := parseBundleSurfaceError(t, "EphemeralVM", `"unsafe_bypass": ["cursor:--ghost"]`)
	if !strings.Contains(errText, "unknown unsafe bypass surface") {
		t.Fatalf("expected unknown surface rejection, got %s", errText)
	}
}

func TestParsePolicyBundleRejectsUnknownNativeConfigHook(t *testing.T) {
	errText := parseBundleSurfaceError(t, "Host", `"native_config_hooks": ["claude:command-hooks:blocking"]`)
	if !strings.Contains(errText, "unknown native config hook surface") {
		t.Fatalf("expected unknown native config hook rejection, got %s", errText)
	}
}

func TestParsePolicyBundleRejectsUnknownNativeConfigFile(t *testing.T) {
	errText := parseBundleSurfaceError(t, "Host", `"native_config_files": ["codex:config-home:global"]`)
	if !strings.Contains(errText, "unknown native config file surface") {
		t.Fatalf("expected unknown native config file rejection, got %s", errText)
	}
}

func TestParsePolicyBundleRejectsUnknownToolUseSurface(t *testing.T) {
	errText := parseBundleSurfaceError(t, "Host", `"tool_use": ["tool:teleport"]`)
	if !strings.Contains(errText, "unknown tool use surface") {
		t.Fatalf("expected unknown tool use surface rejection, got %s", errText)
	}
}
