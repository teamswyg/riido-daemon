package policy

import (
	"time"
)

const BundleSchemaVersion = "riido-policy-bundle.v1"

// PolicyBundle is the executable subset of security.md §2. It is intentionally
// small today: C7 owns the decision data, while C4/C5/C6 execute decisions.
type PolicyBundle struct {
	SchemaVersion     string                        `json:"schema_version"`
	Version           string                        `json:"version"`
	EffectiveSince    time.Time                     `json:"effective_since"`
	SupersededAt      *time.Time                    `json:"superseded_at,omitempty"`
	TrustTierPolicies map[TrustTier]TrustTierPolicy `json:"trust_tier_policies"`
}

type TrustTierPolicy struct {
	AllowedSurfaces AllowedSurfaceSet `json:"allowed_surfaces"`
}

type AllowedSurfaceSet struct {
	UnsafeBypass      []UnsafeBypassSurface     `json:"unsafe_bypass,omitempty"`
	NativeConfigHooks []NativeConfigHookSurface `json:"native_config_hooks,omitempty"`
	NativeConfigFiles []NativeConfigFileSurface `json:"native_config_files,omitempty"`
	ToolUse           []ToolUseSurface          `json:"tool_use,omitempty"`
}
