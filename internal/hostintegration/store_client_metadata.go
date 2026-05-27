package hostintegration

import (
	"fmt"
	"strings"

	"github.com/teamswyg/riido-contracts/provider/capability"
)

// ServerFacingClientMetadata is the C11 privacy-filtered status snapshot that
// the daemon may send to C10. It carries routing inputs, not local host paths.
type ServerFacingClientMetadata struct {
	DistributionChannel DistributionChannel          `json:"distribution_channel"`
	AppVersion          string                       `json:"app_version,omitempty"`
	Providers           []ServerFacingProviderStatus `json:"providers"`
}

// ServerFacingProviderStatus is the C10-safe provider routing subset nested
// under ServerFacingClientMetadata.
type ServerFacingProviderStatus struct {
	ProviderKind        capability.ProviderKind `json:"provider_kind"`
	ProviderAvailable   bool                    `json:"provider_available"`
	ProviderLoginStatus ToolLoginStatus         `json:"provider_login_status"`
	RoutingStatus       ProviderRoutingStatus   `json:"routing_status"`
}

// ServerFacingClientMetadataInput collects the local C11 facts that can be
// projected into the server-facing metadata boundary.
type ServerFacingClientMetadataInput struct {
	Channel    DistributionChannel
	AppVersion string
	Registry   *ExternalToolRegistry
}

// BuildServerFacingClientMetadata returns a deterministic C10-safe projection
// of provider availability for the current store/distribution channel.
func BuildServerFacingClientMetadata(input ServerFacingClientMetadataInput) (ServerFacingClientMetadata, error) {
	if !input.Channel.Valid() {
		return ServerFacingClientMetadata{}, fmt.Errorf("unknown distribution channel %q", input.Channel)
	}
	metadata := ServerFacingClientMetadata{
		DistributionChannel: input.Channel,
		AppVersion:          strings.TrimSpace(input.AppVersion),
	}
	for _, record := range input.Registry.Records() {
		status, err := record.ServerFacingStatus(input.Channel, "")
		if err != nil {
			return ServerFacingClientMetadata{}, err
		}
		metadata.Providers = append(metadata.Providers, ServerFacingProviderStatus{
			ProviderKind:        status.ProviderKind,
			ProviderAvailable:   status.ProviderAvailable,
			ProviderLoginStatus: status.ProviderLoginStatus,
			RoutingStatus:       providerRoutingStatusFromToolStatus(status),
		})
	}
	return metadata, nil
}

// ServerFacingProviderKinds returns the provider kinds present in the metadata
// in the same deterministic order as Providers.
func (m ServerFacingClientMetadata) ServerFacingProviderKinds() []capability.ProviderKind {
	if len(m.Providers) == 0 {
		return nil
	}
	providers := make([]capability.ProviderKind, 0, len(m.Providers))
	for _, provider := range m.Providers {
		providers = append(providers, provider.ProviderKind)
	}
	return providers
}
