package runtimeactor

import providercap "github.com/teamswyg/riido-contracts/provider/capability"

type capabilityProfile struct {
	protocolKind              providercap.ProtocolKind
	protocolMaturity          providercap.ProtocolMaturity
	eventStreamFormat         providercap.EventStreamFormat
	supportsPermissionControl bool
	exposesUnsafeBypass       bool
	supportsApprovalProtocol  bool
	supportsFileEvents        bool
	supportsWorktree          bool
	defaultSandboxMode        string
	defaultApprovalPolicy     string
}
