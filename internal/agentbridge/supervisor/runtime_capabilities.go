package supervisor

import "github.com/teamswyg/riido-daemon/internal/agentbridge/runtimeactor"

func runtimeCapabilityMaps(caps []runtimeactor.Capability) (map[string]bool, map[string]string) {
	boolCaps := map[string]bool{}
	attrs := map[string]string{}
	for _, capability := range caps {
		runtimeBoolCapabilities(boolCaps, capability)
		runtimeCapabilityAttrs(attrs, capability)
	}
	return boolCaps, attrs
}

func runtimeBoolCapabilities(out map[string]bool, capability runtimeactor.Capability) {
	prefix := runtimeCapabilityPrefix(capability)
	out[prefix+"available"] = capability.Available
	out[prefix+"requires_experimental_opt_in"] = capability.RequiresExperimentalOptIn
	out[prefix+"supports_streaming"] = capability.SupportsStreaming
	out[prefix+"supports_resume"] = capability.SupportsResume
	out[prefix+"supports_system"] = capability.SupportsSystem
	out[prefix+"supports_max_turns"] = capability.SupportsMaxTurns
	out[prefix+"supports_mcp"] = capability.SupportsMCP
	out[prefix+"supports_tool_hooks"] = capability.SupportsToolHooks
	out[prefix+"supports_usage"] = capability.SupportsUsage
	out[prefix+"supports_file_events"] = capability.SupportsFileEvents
	out[prefix+"supports_worktree"] = capability.SupportsWorktree
}

func runtimeCapabilityAttrs(out map[string]string, capability runtimeactor.Capability) {
	prefix := runtimeCapabilityPrefix(capability)
	out[prefix+"compatibility_status"] = capability.CompatibilityStatus
	out[prefix+"capability_fingerprint"] = capability.CapabilityFingerprint
	out[prefix+"protocol_kind"] = capability.ProtocolKind
	out[prefix+"protocol_version"] = capability.ProtocolVersion
	out[prefix+"adapter_id"] = capability.AdapterID
	out[prefix+"adapter_version"] = capability.AdapterVersion
	out[prefix+"provider_version"] = capability.Version
}

func runtimeCapabilityPrefix(capability runtimeactor.Capability) string {
	return "provider." + capability.Provider + "."
}
