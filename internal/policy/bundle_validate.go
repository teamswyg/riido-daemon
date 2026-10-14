package policy

import (
	"errors"
	"fmt"
	"strings"
)

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
