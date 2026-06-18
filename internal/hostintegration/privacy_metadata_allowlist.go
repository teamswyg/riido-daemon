package hostintegration

import (
	"embed"
	"encoding/json"
	"fmt"
)

//go:embed privacy_metadata_allowlist.riido.json
var privacyMetadataAllowlistFS embed.FS

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
