package hostintegration

import (
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
)

const PrivacyMetadataAllowlistSchemaVersion = "riido-privacy-metadata-allowlist.v1"

const (
	PrivacySurfaceServerFacingClientMetadata = "c11-server-facing-client-metadata"
	PrivacySurfaceProviderStatusSyncRequest  = "c10-provider-status-sync-request"
)

//go:embed privacy_metadata_allowlist.riido.json
var privacyMetadataAllowlistFS embed.FS

// PrivacyMetadataAllowlist is the executable C10/C11 policy artifact that
// keeps store metadata, public privacy policy, and server request fields aligned.
type PrivacyMetadataAllowlist struct {
	SchemaVersion string                         `json:"schema_version"`
	Surfaces      []PrivacyMetadataSurfacePolicy `json:"surfaces"`
}

// PrivacyMetadataSurfacePolicy describes one JSON boundary where local daemon
// facts may cross toward C10.
type PrivacyMetadataSurfacePolicy struct {
	ID                 string   `json:"id"`
	OwnerContext       string   `json:"owner_context"`
	AllowedJSONPaths   []string `json:"allowed_json_paths"`
	ForbiddenJSONPaths []string `json:"forbidden_json_paths"`
}

// LoadPrivacyMetadataAllowlist loads and validates the checked-in policy
// artifact. The artifact is intentionally data, not code, so review metadata can
// cite it directly.
func LoadPrivacyMetadataAllowlist() (PrivacyMetadataAllowlist, error) {
	data, err := privacyMetadataAllowlistFS.ReadFile("privacy_metadata_allowlist.riido.json")
	if err != nil {
		return PrivacyMetadataAllowlist{}, err
	}
	var allowlist PrivacyMetadataAllowlist
	if err := json.Unmarshal(data, &allowlist); err != nil {
		return PrivacyMetadataAllowlist{}, fmt.Errorf("decode privacy metadata allowlist: %w", err)
	}
	if err := allowlist.Validate(); err != nil {
		return PrivacyMetadataAllowlist{}, err
	}
	return allowlist, nil
}

func (a PrivacyMetadataAllowlist) Validate() error {
	var errs []error
	if a.SchemaVersion != PrivacyMetadataAllowlistSchemaVersion {
		errs = append(errs, fmt.Errorf("unknown privacy metadata allowlist schema %q", a.SchemaVersion))
	}
	seen := map[string]struct{}{}
	for i, surface := range a.Surfaces {
		if surface.ID == "" {
			errs = append(errs, fmt.Errorf("surfaces[%d].id is required", i))
		}
		if _, ok := seen[surface.ID]; ok {
			errs = append(errs, fmt.Errorf("surfaces[%d].id duplicates %s", i, surface.ID))
		}
		seen[surface.ID] = struct{}{}
		if surface.OwnerContext == "" {
			errs = append(errs, fmt.Errorf("surfaces[%d].owner_context is required", i))
		}
		if len(surface.AllowedJSONPaths) == 0 {
			errs = append(errs, fmt.Errorf("surfaces[%d].allowed_json_paths is required", i))
		}
		forbidden := map[string]struct{}{}
		for _, path := range surface.ForbiddenJSONPaths {
			if path == "" {
				errs = append(errs, fmt.Errorf("surfaces[%d].forbidden_json_paths contains empty path", i))
				continue
			}
			forbidden[path] = struct{}{}
		}
		for _, path := range surface.AllowedJSONPaths {
			if path == "" {
				errs = append(errs, fmt.Errorf("surfaces[%d].allowed_json_paths contains empty path", i))
				continue
			}
			if _, ok := forbidden[path]; ok {
				errs = append(errs, fmt.Errorf("surfaces[%d] allows forbidden path %s", i, path))
			}
		}
	}
	for _, required := range []string{PrivacySurfaceServerFacingClientMetadata, PrivacySurfaceProviderStatusSyncRequest} {
		if _, ok := seen[required]; !ok {
			errs = append(errs, fmt.Errorf("missing privacy metadata surface %s", required))
		}
	}
	return errors.Join(errs...)
}

func (a PrivacyMetadataAllowlist) Surface(id string) (PrivacyMetadataSurfacePolicy, bool) {
	for _, surface := range a.Surfaces {
		if surface.ID == id {
			return surface, true
		}
	}
	return PrivacyMetadataSurfacePolicy{}, false
}

func (s PrivacyMetadataSurfacePolicy) Allows(path string) bool {
	return slices.Contains(s.AllowedJSONPaths, path)
}

func (s PrivacyMetadataSurfacePolicy) Forbids(path string) bool {
	return slices.Contains(s.ForbiddenJSONPaths, path)
}
