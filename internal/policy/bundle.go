package policy

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"
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

func DefaultLocalPolicyBundle() PolicyBundle {
	return PolicyBundle{
		SchemaVersion:  BundleSchemaVersion,
		Version:        DefaultLocalPolicyBundleVersion,
		EffectiveSince: time.Date(2026, 5, 27, 0, 0, 0, 0, time.UTC),
		TrustTierPolicies: map[TrustTier]TrustTierPolicy{
			TrustTierHost: {
				AllowedSurfaces: AllowedSurfaceSet{
					NativeConfigHooks: []NativeConfigHookSurface{
						NativeConfigHookClaudeCommandAudit,
					},
					NativeConfigFiles: []NativeConfigFileSurface{
						NativeConfigFileCodexTaskScopedHome,
					},
				},
			},
		},
	}
}

func LoadPolicyBundleFile(path string) (PolicyBundle, error) {
	trimmed := strings.TrimSpace(path)
	if trimmed == "" {
		return PolicyBundle{}, errors.New("policy: bundle path is required")
	}
	data, err := os.ReadFile(trimmed)
	if err != nil {
		return PolicyBundle{}, fmt.Errorf("policy: load bundle %s: %w", trimmed, err)
	}
	bundle, err := ParsePolicyBundleJSON(data)
	if err != nil {
		return PolicyBundle{}, fmt.Errorf("policy: load bundle %s: %w", trimmed, err)
	}
	return bundle, nil
}

func ParsePolicyBundleJSON(data []byte) (PolicyBundle, error) {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	var bundle PolicyBundle
	if err := dec.Decode(&bundle); err != nil {
		return PolicyBundle{}, fmt.Errorf("parse policy bundle: %w", err)
	}
	var extra any
	if err := dec.Decode(&extra); !errors.Is(err, io.EOF) {
		return PolicyBundle{}, errors.New("parse policy bundle: trailing JSON value")
	}
	if err := bundle.Validate(); err != nil {
		return PolicyBundle{}, err
	}
	return bundle, nil
}

func (b PolicyBundle) Validate() error {
	if b.SchemaVersion != BundleSchemaVersion {
		return fmt.Errorf("policy: schema_version must be %q", BundleSchemaVersion)
	}
	if strings.TrimSpace(b.Version) == "" {
		return errors.New("policy: version is required")
	}
	if b.EffectiveSince.IsZero() {
		return errors.New("policy: effective_since is required")
	}
	if b.SupersededAt != nil && !b.SupersededAt.After(b.EffectiveSince) {
		return errors.New("policy: superseded_at must be after effective_since")
	}
	if b.TrustTierPolicies == nil {
		return errors.New("policy: trust_tier_policies is required")
	}
	for tier, tierPolicy := range b.TrustTierPolicies {
		if !isKnownTrustTier(tier) {
			return fmt.Errorf("policy: unknown trust tier %q", tier)
		}
		if err := validateAllowedSurfaces(tier, tierPolicy.AllowedSurfaces); err != nil {
			return err
		}
	}
	return nil
}

func (b PolicyBundle) AllowsUnsafeBypass(tier TrustTier, surface UnsafeBypassSurface) bool {
	tierPolicy, ok := b.TrustTierPolicies[normalizeTrustTier(tier)]
	if !ok {
		return false
	}
	return slices.Contains(tierPolicy.AllowedSurfaces.UnsafeBypass, surface)
}

func (b PolicyBundle) AllowsNativeConfigHook(tier TrustTier, surface NativeConfigHookSurface) bool {
	tierPolicy, ok := b.TrustTierPolicies[normalizeTrustTier(tier)]
	if !ok {
		return false
	}
	return slices.Contains(tierPolicy.AllowedSurfaces.NativeConfigHooks, surface)
}

func (b PolicyBundle) AllowsNativeConfigFile(tier TrustTier, surface NativeConfigFileSurface) bool {
	tierPolicy, ok := b.TrustTierPolicies[normalizeTrustTier(tier)]
	if !ok {
		return false
	}
	return slices.Contains(tierPolicy.AllowedSurfaces.NativeConfigFiles, surface)
}

func (b PolicyBundle) AllowsToolUse(tier TrustTier, surface ToolUseSurface) bool {
	tierPolicy, ok := b.TrustTierPolicies[normalizeTrustTier(tier)]
	if !ok {
		return false
	}
	return slices.Contains(tierPolicy.AllowedSurfaces.ToolUse, surface)
}

func EvaluateUnsafeBypassWithBundle(bundle PolicyBundle, input UnsafeBypassInput) Decision {
	input.BundleAllows = bundle.AllowsUnsafeBypass(input.TrustTier, input.Surface)
	input.PolicyVersion = bundle.Version
	return EvaluateUnsafeBypass(input)
}

func EvaluateNativeConfigHookWithBundle(bundle PolicyBundle, input NativeConfigHookInput) Decision {
	input.BundleAllows = bundle.AllowsNativeConfigHook(input.TrustTier, input.Surface)
	input.PolicyVersion = bundle.Version
	return EvaluateNativeConfigHook(input)
}

func EvaluateNativeConfigFileWithBundle(bundle PolicyBundle, input NativeConfigFileInput) Decision {
	input.BundleAllows = bundle.AllowsNativeConfigFile(input.TrustTier, input.Surface)
	input.PolicyVersion = bundle.Version
	return EvaluateNativeConfigFile(input)
}

func EvaluateToolUseWithBundle(bundle PolicyBundle, input ToolUseInput) ToolUseDecision {
	input.BundleAllows = bundle.AllowsToolUse(input.TrustTier, input.Surface)
	input.PolicyVersion = bundle.Version
	return EvaluateToolUse(input)
}

func validateAllowedSurfaces(tier TrustTier, surfaces AllowedSurfaceSet) error {
	seenUnsafe := map[UnsafeBypassSurface]bool{}
	for _, surface := range surfaces.UnsafeBypass {
		if !isKnownUnsafeBypassSurface(surface) {
			return fmt.Errorf("policy: unknown unsafe bypass surface %q", surface)
		}
		if seenUnsafe[surface] {
			return fmt.Errorf("policy: duplicate unsafe bypass surface %q", surface)
		}
		seenUnsafe[surface] = true
		if tier == TrustTierHost || tier == TrustTierUnknown {
			return fmt.Errorf("policy: trust tier %s cannot allow unsafe bypass surface %q", tier, surface)
		}
	}
	seenHooks := map[NativeConfigHookSurface]bool{}
	for _, surface := range surfaces.NativeConfigHooks {
		if !isKnownNativeConfigHookSurface(surface) {
			return fmt.Errorf("policy: unknown native config hook surface %q", surface)
		}
		if seenHooks[surface] {
			return fmt.Errorf("policy: duplicate native config hook surface %q", surface)
		}
		seenHooks[surface] = true
		if tier == TrustTierUnknown {
			return fmt.Errorf("policy: trust tier %s cannot allow native config hook surface %q", tier, surface)
		}
	}
	seenFiles := map[NativeConfigFileSurface]bool{}
	for _, surface := range surfaces.NativeConfigFiles {
		if !isKnownNativeConfigFileSurface(surface) {
			return fmt.Errorf("policy: unknown native config file surface %q", surface)
		}
		if seenFiles[surface] {
			return fmt.Errorf("policy: duplicate native config file surface %q", surface)
		}
		seenFiles[surface] = true
		if tier == TrustTierUnknown {
			return fmt.Errorf("policy: trust tier %s cannot allow native config file surface %q", tier, surface)
		}
	}
	seenToolUse := map[ToolUseSurface]bool{}
	for _, surface := range surfaces.ToolUse {
		if !isKnownToolUseSurface(surface) {
			return fmt.Errorf("policy: unknown tool use surface %q", surface)
		}
		if seenToolUse[surface] {
			return fmt.Errorf("policy: duplicate tool use surface %q", surface)
		}
		seenToolUse[surface] = true
		if tier == TrustTierUnknown {
			return fmt.Errorf("policy: trust tier %s cannot allow tool use surface %q", tier, surface)
		}
	}
	return nil
}

func isKnownTrustTier(tier TrustTier) bool {
	switch tier {
	case TrustTierHost, TrustTierIsolatedContainer, TrustTierEphemeralVM, TrustTierCIControlledRunner, TrustTierUnknown:
		return true
	default:
		return false
	}
}

func isKnownUnsafeBypassSurface(surface UnsafeBypassSurface) bool {
	switch surface {
	case UnsafeBypassClaudePermissions, UnsafeBypassCursorYolo, UnsafeBypassCodexYolo, UnsafeBypassCodexDangerBypass, UnsafeBypassCodexDangerAccess:
		return true
	default:
		return false
	}
}

func isKnownNativeConfigHookSurface(surface NativeConfigHookSurface) bool {
	switch surface {
	case NativeConfigHookClaudeCommandAudit:
		return true
	default:
		return false
	}
}

func isKnownNativeConfigFileSurface(surface NativeConfigFileSurface) bool {
	switch surface {
	case NativeConfigFileCodexTaskScopedHome:
		return true
	default:
		return false
	}
}

func isKnownToolUseSurface(surface ToolUseSurface) bool {
	switch surface {
	case ToolUseNetworkEgress, ToolUseProtectedPathWrite, ToolUseSecretExposure, ToolUseDestructiveCommand:
		return true
	default:
		return false
	}
}
