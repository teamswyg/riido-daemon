package policy

import "testing"

func TestDefaultLocalPolicyBundleAllowsOnlyClaudeNativeConfigSurface(t *testing.T) {
	bundle := DefaultLocalPolicyBundle()

	if err := bundle.Validate(); err != nil {
		t.Fatal(err)
	}
	if bundle.Version != DefaultLocalPolicyBundleVersion {
		t.Fatalf("default bundle version = %q", bundle.Version)
	}
	if !bundle.AllowsNativeConfigHook(TrustTierHost, NativeConfigHookClaudeCommandAudit) {
		t.Fatal("local default should allow Claude audit hook materialization")
	}
	if bundle.AllowsNativeConfigFile(TrustTierHost, NativeConfigFileCodexTaskScopedHome) {
		t.Fatal("local default must not materialize Codex task-scoped config home")
	}
	if bundle.AllowsUnsafeBypass(TrustTierHost, UnsafeBypassClaudePermissions) {
		t.Fatal("local default must not allow unsafe bypass on Host")
	}
	if bundle.AllowsToolUse(TrustTierHost, ToolUseNetworkEgress) {
		t.Fatal("local default must not silently allow tool-use risk surfaces")
	}
}
