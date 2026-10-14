package hostintegration

import (
	"errors"
	"fmt"
)

// DefaultLocalIPCEndpoint selects the local-only endpoint shape for a channel
// and host OS. It does not open sockets or named pipes.
func DefaultLocalIPCEndpoint(in LocalIPCEndpointInput) (LocalIPCEndpoint, error) {
	if !in.Channel.Valid() {
		return LocalIPCEndpoint{}, fmt.Errorf("unknown distribution channel %q", in.Channel)
	}
	if !in.HostOS.Valid() {
		return LocalIPCEndpoint{}, fmt.Errorf("unknown host OS %q", in.HostOS)
	}
	if !in.Owner.Valid() {
		return LocalIPCEndpoint{}, fmt.Errorf("unknown local IPC owner %q", in.Owner)
	}
	if err := validateEndpointRoot(in); err != nil {
		return LocalIPCEndpoint{}, err
	}
	name, err := normalizedEndpointName(in.Name)
	if err != nil {
		return LocalIPCEndpoint{}, err
	}

	switch in.HostOS {
	case HostOSDarwin:
		if in.Channel == DistributionChannelMSIXSideload || in.Channel == DistributionChannelMSIXStore {
			return LocalIPCEndpoint{}, errors.New("msix channels require windows host OS")
		}
		return LocalIPCEndpoint{
			Channel:      in.Channel,
			HostOS:       in.HostOS,
			EndpointKind: LocalIPCEndpointUnixSocket,
			Path:         joinHostPath(in.HostOS, in.AppDataRoot.Path, name+".sock"),
			Owner:        in.Owner,
		}, nil
	case HostOSWindows:
		if in.Channel == DistributionChannelMacAppStore || in.Channel == DistributionChannelDeveloperID {
			return LocalIPCEndpoint{}, errors.New("mac distribution channels require darwin host OS")
		}
		return LocalIPCEndpoint{
			Channel:      in.Channel,
			HostOS:       in.HostOS,
			EndpointKind: LocalIPCEndpointNamedPipe,
			Path:         namedPipePath(in.Channel, in.Owner, name),
			Owner:        in.Owner,
		}, nil
	default:
		return LocalIPCEndpoint{}, fmt.Errorf("unsupported host OS %q", in.HostOS)
	}
}
