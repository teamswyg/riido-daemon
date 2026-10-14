package runtimeactor

import (
	providercap "github.com/teamswyg/riido-contracts/provider/capability"
	providercatalog "github.com/teamswyg/riido-contracts/provider/catalog"
)

func profileForProvider(provider string) capabilityProfile {
	providerKind := providercatalog.Normalize(provider)
	switch {
	case providercatalog.IsClaudeFamily(provider):
		return claudeCapabilityProfile()
	case providerKind == providercatalog.KindCodex:
		return codexCapabilityProfile()
	case providerKind == providercatalog.KindOpenClaw:
		return openClawCapabilityProfile()
	case providerKind == providercatalog.KindCursor:
		return cursorCapabilityProfile()
	default:
		return unknownCapabilityProfile(providerKind)
	}
}

func unknownCapabilityProfile(providerKind providercatalog.Kind) capabilityProfile {
	return capabilityProfile{
		protocolKind:          providercap.ProtocolKind(string(providerKind) + "-unknown"),
		protocolMaturity:      providercap.ProtocolMaturityUnknown,
		eventStreamFormat:     providercap.EventStreamFormatUnknown,
		defaultSandboxMode:    "unknown",
		defaultApprovalPolicy: "unknown",
	}
}
