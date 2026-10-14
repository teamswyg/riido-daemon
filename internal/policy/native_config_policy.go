package policy

// NativeConfigHookSurface is one provider-native hook materialization surface
// covered by security.md T-CFG. It is not an unsafe bypass, but it still needs
// a policy decision before C6 writes hook settings/scripts into a task workdir.
type NativeConfigHookSurface string

const (
	NativeConfigHookClaudeCommandAudit NativeConfigHookSurface = "claude:command-hooks:audit"
)

// NativeConfigHookInput is the policy question: may C6 materialize this
// provider-native hook surface into the task workdir?
type NativeConfigHookInput struct {
	TrustTier     TrustTier
	Surface       NativeConfigHookSurface
	BundleAllows  bool
	PolicyVersion string
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
