package runtimeactor

import providercap "github.com/teamswyg/riido-contracts/provider/capability"

func claudeCapabilityProfile() capabilityProfile {
	return capabilityProfile{
		protocolKind:              providercap.ProtocolClaudeStreamJSON,
		protocolMaturity:          providercap.ProtocolMaturityStable,
		eventStreamFormat:         providercap.EventStreamFormatNDJSON,
		supportsPermissionControl: true,
		exposesUnsafeBypass:       true,
		supportsWorktree:          true,
		defaultSandboxMode:        "unknown",
		defaultApprovalPolicy:     "on-request",
	}
}

func codexCapabilityProfile() capabilityProfile {
	return capabilityProfile{
		protocolKind:              providercap.ProtocolCodexAppServer,
		protocolMaturity:          providercap.ProtocolMaturityExperimental,
		eventStreamFormat:         providercap.EventStreamFormatJSONRPCNotifications,
		supportsPermissionControl: true,
		exposesUnsafeBypass:       true,
		supportsApprovalProtocol:  true,
		supportsWorktree:          true,
		defaultSandboxMode:        "workspace-write",
		defaultApprovalPolicy:     "on-request",
	}
}

func openClawCapabilityProfile() capabilityProfile {
	return capabilityProfile{
		protocolKind:          providercap.ProtocolOpenClawAgentJSON,
		protocolMaturity:      providercap.ProtocolMaturityExperimental,
		eventStreamFormat:     providercap.EventStreamFormatNDJSON,
		supportsWorktree:      false,
		defaultSandboxMode:    "unknown",
		defaultApprovalPolicy: "unknown",
	}
}

func cursorCapabilityProfile() capabilityProfile {
	return capabilityProfile{
		protocolKind:          providercap.ProtocolCursorAgentStreamJSON,
		protocolMaturity:      providercap.ProtocolMaturityExperimental,
		eventStreamFormat:     providercap.EventStreamFormatNDJSON,
		exposesUnsafeBypass:   true,
		supportsWorktree:      true,
		defaultSandboxMode:    "unknown",
		defaultApprovalPolicy: "unknown",
	}
}
