package policy

// NativeConfigFileSurface is one provider-native config-home/file
// materialization surface covered by security.md T-CFG. It is not an unsafe
// bypass, but it still decides whether C6 may write provider settings that
// redirect a provider CLI to task-scoped config.
type NativeConfigFileSurface string

const (
	NativeConfigFileCodexTaskScopedHome NativeConfigFileSurface = "codex:config-home:task-scoped"
)

// NativeConfigFileInput is the policy question: may C6 materialize this
// provider-native config-home/file surface into the task workdir?
type NativeConfigFileInput struct {
	TrustTier     TrustTier
	Surface       NativeConfigFileSurface
	BundleAllows  bool
	PolicyVersion string
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
