package hostintegration

import (
	"fmt"
	"strings"

	"github.com/teamswyg/riido-contracts/provider/capability"
)

// ServerFacingToolStatus is the privacy-filtered subset of a registration row
// that may cross into C10. It intentionally has no executable path, workspace
// path, token, or provider secret fields.
type ServerFacingToolStatus struct {
	DistributionChannel DistributionChannel     `json:"distribution_channel"`
	AppVersion          string                  `json:"app_version,omitempty"`
	ProviderKind        capability.ProviderKind `json:"provider_kind"`
	ProviderAvailable   bool                    `json:"provider_available"`
	ProviderLoginStatus ToolLoginStatus         `json:"provider_login_status"`
}

// ServerFacingStatus returns the C10-safe projection of the local registration
// row. Do not add path-like fields here; distribution-host-integration.md §7 is
// the SSOT for this boundary.
func (r ExternalToolRecord) ServerFacingStatus(channel DistributionChannel, appVersion string) (ServerFacingToolStatus, error) {
	if !channel.Valid() {
		return ServerFacingToolStatus{}, fmt.Errorf("unknown distribution channel %q", channel)
	}
	if err := r.Validate(); err != nil {
		return ServerFacingToolStatus{}, err
	}
	return ServerFacingToolStatus{
		DistributionChannel: channel,
		AppVersion:          strings.TrimSpace(appVersion),
		ProviderKind:        r.Provider,
		ProviderAvailable:   r.ProviderAvailable(),
		ProviderLoginStatus: r.LoginStatus,
	}, nil
}
