package policy

// ToolUseSurface is one provider tool-use risk surface covered by
// security.md G-S3. It describes the policy risk, not a provider-specific
// raw tool name.
type ToolUseSurface string

const (
	ToolUseNetworkEgress      ToolUseSurface = "tool:network-egress"
	ToolUseProtectedPathWrite ToolUseSurface = "tool:protected-path-write"
	ToolUseSecretExposure     ToolUseSurface = "tool:secret-exposure"
	ToolUseDestructiveCommand ToolUseSurface = "tool:destructive-command"
)

// ToolUseAction is the current executable branch set for G-S3.
type ToolUseAction string

const (
	ToolUseActionAllow             ToolUseAction = "allow"
	ToolUseActionRequireApproval   ToolUseAction = "require-approval"
	ToolUseActionInterruptAndBlock ToolUseAction = "interrupt-and-block"
)

// ToolUseInput is the G-S3 policy question: what should happen before a
// provider tool call with this risk surface starts?
type ToolUseInput struct {
	TrustTier              TrustTier
	Surface                ToolUseSurface
	BundleAllows           bool
	HumanApprovalAvailable bool
	PolicyVersion          string
}

// ToolUseDecision is the stable branch result returned by ToolUseSecurityGate.
type ToolUseDecision struct {
	Action ToolUseAction
	Code   string
	Reason string
}
