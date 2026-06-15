// Package policy implements the C7 Security / Policy decision helpers.
//
// It owns small, pure policy decisions that adjacent contexts consult before
// turning a risk surface into concrete provider flags. It does not spawn
// provider processes, mutate task state, or inspect provider capabilities; C4
// / C5 / C6 execute the decisions this package returns.
package policy

// TrustTier describes the runtime isolation level from
// docs/20-domain/security.md §1.
type TrustTier string

const (
	TrustTierHost               TrustTier = "Host"
	TrustTierIsolatedContainer  TrustTier = "IsolatedContainer"
	TrustTierEphemeralVM        TrustTier = "EphemeralVM"
	TrustTierCIControlledRunner TrustTier = "CIControlledRunner"
	TrustTierUnknown            TrustTier = "Unknown"
)

const DefaultLocalPolicyBundleVersion = "policy-bundle.local.v0"

// UnsafeBypassSurface is one concrete provider surface covered by
// security.md §5.
type UnsafeBypassSurface string

const (
	UnsafeBypassClaudePermissions UnsafeBypassSurface = "claude:bypassPermissions"
	UnsafeBypassCursorYolo        UnsafeBypassSurface = "cursor:--yolo"
	UnsafeBypassCodexYolo         UnsafeBypassSurface = "codex:--yolo"
	UnsafeBypassCodexDangerBypass UnsafeBypassSurface = "codex:--dangerously-bypass-approvals-and-sandbox"
)

// NativeConfigHookSurface is one provider-native hook materialization surface
// covered by security.md T-CFG. It is not an unsafe bypass, but it still needs
// a policy decision before C6 writes hook settings/scripts into a task workdir.
type NativeConfigHookSurface string

const (
	NativeConfigHookClaudeCommandAudit NativeConfigHookSurface = "claude:command-hooks:audit"
)

// NativeConfigFileSurface is one provider-native config-home/file
// materialization surface covered by security.md T-CFG. It is not an unsafe
// bypass, but it still decides whether C6 may write provider settings that
// redirect a provider CLI to task-scoped config.
type NativeConfigFileSurface string

const (
	NativeConfigFileCodexTaskScopedHome NativeConfigFileSurface = "codex:config-home:task-scoped"
)

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

// UnsafeBypassInput is the policy question: may this runtime activate a
// provider unsafe-bypass surface?
type UnsafeBypassInput struct {
	TrustTier     TrustTier
	Surface       UnsafeBypassSurface
	BundleAllows  bool
	PolicyVersion string
}

// NativeConfigHookInput is the policy question: may C6 materialize this
// provider-native hook surface into the task workdir?
type NativeConfigHookInput struct {
	TrustTier     TrustTier
	Surface       NativeConfigHookSurface
	BundleAllows  bool
	PolicyVersion string
}

// NativeConfigFileInput is the policy question: may C6 materialize this
// provider-native config-home/file surface into the task workdir?
type NativeConfigFileInput struct {
	TrustTier     TrustTier
	Surface       NativeConfigFileSurface
	BundleAllows  bool
	PolicyVersion string
}

// ToolUseInput is the G-S3 policy question: what should happen before a
// provider tool call with this risk surface starts?
type ToolUseInput struct {
	TrustTier              TrustTier
	Surface                ToolUseSurface
	BundleAllows           bool
	HumanApprovalAvailable bool
	PolicyVersion          string
}

// Decision is the stable result shape returned by policy helpers.
type Decision struct {
	Allowed bool
	Code    string
	Reason  string
}

// ToolUseDecision is the stable branch result returned by ToolUseSecurityGate.
type ToolUseDecision struct {
	Action ToolUseAction
	Code   string
	Reason string
}

// EvaluateUnsafeBypass implements security.md §5:
// Host/Unknown always deny; isolated tiers require an explicit bundle allow.
func EvaluateUnsafeBypass(input UnsafeBypassInput) Decision {
	tier := normalizeTrustTier(input.TrustTier)
	switch tier {
	case TrustTierHost:
		return deny("UNSAFE_BYPASS_ON_HOST", "unsafe bypass is forbidden on Host trust tier")
	case TrustTierUnknown:
		return deny("UNSAFE_BYPASS_UNKNOWN_TRUST_TIER", "unsafe bypass requires a verified isolated trust tier")
	case TrustTierIsolatedContainer, TrustTierEphemeralVM, TrustTierCIControlledRunner:
		if !input.BundleAllows {
			return deny("UNSAFE_BYPASS_NOT_IN_POLICY_BUNDLE", "unsafe bypass is not explicitly allowed by the active policy bundle")
		}
		return Decision{Allowed: true, Code: "ALLOWED", Reason: "unsafe bypass allowed by isolated trust tier and active policy bundle"}
	default:
		return deny("UNSAFE_BYPASS_UNKNOWN_TRUST_TIER", "unsafe bypass requires a known trust tier")
	}
}

// EvaluateNativeConfigHook implements the executable T-CFG hook rule:
// a known runtime tier and explicit policy-bundle allow are both required.
func EvaluateNativeConfigHook(input NativeConfigHookInput) Decision {
	tier := normalizeTrustTier(input.TrustTier)
	switch tier {
	case TrustTierUnknown:
		return deny("NATIVE_CONFIG_HOOK_UNKNOWN_TRUST_TIER", "native config hooks require a known runtime trust tier")
	case TrustTierHost, TrustTierIsolatedContainer, TrustTierEphemeralVM, TrustTierCIControlledRunner:
		if !input.BundleAllows {
			return deny("NATIVE_CONFIG_HOOK_NOT_IN_POLICY_BUNDLE", "native config hook surface is not explicitly allowed by the active policy bundle")
		}
		return Decision{Allowed: true, Code: "ALLOWED", Reason: "native config hook surface allowed by active policy bundle"}
	default:
		return deny("NATIVE_CONFIG_HOOK_UNKNOWN_TRUST_TIER", "native config hooks require a known runtime trust tier")
	}
}

// EvaluateNativeConfigFile implements the executable T-CFG file rule:
// a known runtime tier and explicit policy-bundle allow are both required.
func EvaluateNativeConfigFile(input NativeConfigFileInput) Decision {
	tier := normalizeTrustTier(input.TrustTier)
	switch tier {
	case TrustTierUnknown:
		return deny("NATIVE_CONFIG_FILE_UNKNOWN_TRUST_TIER", "native config files require a known runtime trust tier")
	case TrustTierHost, TrustTierIsolatedContainer, TrustTierEphemeralVM, TrustTierCIControlledRunner:
		if !input.BundleAllows {
			return deny("NATIVE_CONFIG_FILE_NOT_IN_POLICY_BUNDLE", "native config file surface is not explicitly allowed by the active policy bundle")
		}
		return Decision{Allowed: true, Code: "ALLOWED", Reason: "native config file surface allowed by active policy bundle"}
	default:
		return deny("NATIVE_CONFIG_FILE_UNKNOWN_TRUST_TIER", "native config files require a known runtime trust tier")
	}
}

// EvaluateToolUse implements the current executable subset of security.md G-S3:
// known tiers and explicit bundle allow can proceed; missing allow can only
// continue through a human approval branch; Unknown trust tier always blocks.
func EvaluateToolUse(input ToolUseInput) ToolUseDecision {
	if !isKnownToolUseSurface(input.Surface) {
		return interruptAndBlock("TOOL_USE_UNKNOWN_SURFACE", "tool use surface is not known to the active policy bundle")
	}
	tier := normalizeTrustTier(input.TrustTier)
	switch tier {
	case TrustTierUnknown:
		return interruptAndBlock("TOOL_USE_UNKNOWN_TRUST_TIER", "tool use requires a known runtime trust tier")
	case TrustTierHost, TrustTierIsolatedContainer, TrustTierEphemeralVM, TrustTierCIControlledRunner:
		if input.BundleAllows {
			return ToolUseDecision{Action: ToolUseActionAllow, Code: "ALLOWED", Reason: "tool use surface allowed by active policy bundle"}
		}
		if input.HumanApprovalAvailable {
			return ToolUseDecision{Action: ToolUseActionRequireApproval, Code: "TOOL_USE_REQUIRES_APPROVAL", Reason: "tool use surface is not explicitly allowed by the active policy bundle"}
		}
		return interruptAndBlock("TOOL_USE_NOT_IN_POLICY_BUNDLE", "tool use surface is not explicitly allowed and no approval path is available")
	default:
		return interruptAndBlock("TOOL_USE_UNKNOWN_TRUST_TIER", "tool use requires a known runtime trust tier")
	}
}

func normalizeTrustTier(tier TrustTier) TrustTier {
	if tier == "" {
		return TrustTierUnknown
	}
	return tier
}

func deny(code, reason string) Decision {
	return Decision{Allowed: false, Code: code, Reason: reason}
}

func interruptAndBlock(code, reason string) ToolUseDecision {
	return ToolUseDecision{Action: ToolUseActionInterruptAndBlock, Code: code, Reason: reason}
}
