package hostintegration

import (
	"errors"
	"fmt"
	"strings"
	"unicode"
)

// LocalIPCEndpointKind is the OS-level local-only transport family. C11 owns
// the mapping; C1-C10 should only consume provider-neutral endpoint facts.
type LocalIPCEndpointKind string

const (
	LocalIPCEndpointUnixSocket LocalIPCEndpointKind = "unix-socket"
	LocalIPCEndpointNamedPipe  LocalIPCEndpointKind = "named-pipe"
)

// LocalIPCOwner records which packaged role owns an endpoint.
type LocalIPCOwner string

const (
	LocalIPCOwnerStoreApp LocalIPCOwner = "store-app"
	LocalIPCOwnerHelper   LocalIPCOwner = "helper"
)

// LocalIPCEndpoint is the C11 local-only IPC endpoint descriptor.
type LocalIPCEndpoint struct {
	Channel      DistributionChannel
	HostOS       HostOS
	EndpointKind LocalIPCEndpointKind
	Path         string
	Owner        LocalIPCOwner
}

// LocalIPCEndpointInput is supplied by a C11 adapter after it has resolved the
// channel-approved app data root.
type LocalIPCEndpointInput struct {
	Channel     DistributionChannel
	HostOS      HostOS
	AppDataRoot AppDataRoot
	Owner       LocalIPCOwner
	Name        string
}

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

// Valid reports whether kind is one of the SSOT-defined endpoint kinds.
func (kind LocalIPCEndpointKind) Valid() bool {
	switch kind {
	case LocalIPCEndpointUnixSocket, LocalIPCEndpointNamedPipe:
		return true
	default:
		return false
	}
}

// Valid reports whether owner is one of the SSOT-defined endpoint owners.
func (owner LocalIPCOwner) Valid() bool {
	switch owner {
	case LocalIPCOwnerStoreApp, LocalIPCOwnerHelper:
		return true
	default:
		return false
	}
}

func validateEndpointRoot(in LocalIPCEndpointInput) error {
	if strings.TrimSpace(in.AppDataRoot.Path) == "" {
		return errors.New("local IPC endpoint requires app data root")
	}
	if in.AppDataRoot.Channel != in.Channel {
		return fmt.Errorf("app data root channel %q does not match endpoint channel %q", in.AppDataRoot.Channel, in.Channel)
	}
	if in.AppDataRoot.HostOS != in.HostOS {
		return fmt.Errorf("app data root host OS %q does not match endpoint host OS %q", in.AppDataRoot.HostOS, in.HostOS)
	}
	return nil
}

func normalizedEndpointName(raw string) (string, error) {
	name := strings.TrimSpace(raw)
	if name == "" {
		name = "riido"
	}
	for _, r := range name {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' || r == '_' || r == '.' {
			continue
		}
		return "", fmt.Errorf("invalid local IPC endpoint name %q", raw)
	}
	if strings.ContainsAny(name, `/\:`) {
		return "", fmt.Errorf("invalid local IPC endpoint name %q", raw)
	}
	return name, nil
}

func namedPipePath(channel DistributionChannel, owner LocalIPCOwner, name string) string {
	return `\\.\pipe\riido-` + string(channel) + "-" + string(owner) + "-" + name
}
