package hostintegration

import (
	"errors"
	"fmt"
)

func (a PrivacyMetadataAllowlist) Validate() error {
	var errs []error
	if a.SchemaVersion != PrivacyMetadataAllowlistSchemaVersion {
		errs = append(errs, fmt.Errorf("unknown privacy metadata allowlist schema %q", a.SchemaVersion))
	}
	seen := map[string]struct{}{}
	for i, surface := range a.Surfaces {
		errs = append(errs, validatePrivacySurface(i, surface, seen)...)
	}
	for _, required := range requiredPrivacyMetadataSurfaces() {
		if _, ok := seen[required]; !ok {
			errs = append(errs, fmt.Errorf("missing privacy metadata surface %s", required))
		}
	}
	return errors.Join(errs...)
}

func requiredPrivacyMetadataSurfaces() []string {
	return []string{
		PrivacySurfaceServerFacingClientMetadata,
		PrivacySurfaceProviderStatusSyncRequest,
	}
}
