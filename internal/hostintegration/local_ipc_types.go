package hostintegration

// LocalIPCEndpointKind is the OS-level local-only transport family.
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

// LocalIPCEndpointInput is supplied after resolving an approved app data root.
type LocalIPCEndpointInput struct {
	Channel     DistributionChannel
	HostOS      HostOS
	AppDataRoot AppDataRoot
	Owner       LocalIPCOwner
	Name        string
}

func (kind LocalIPCEndpointKind) Valid() bool {
	switch kind {
	case LocalIPCEndpointUnixSocket, LocalIPCEndpointNamedPipe:
		return true
	default:
		return false
	}
}

func (owner LocalIPCOwner) Valid() bool {
	switch owner {
	case LocalIPCOwnerStoreApp, LocalIPCOwnerHelper:
		return true
	default:
		return false
	}
}
