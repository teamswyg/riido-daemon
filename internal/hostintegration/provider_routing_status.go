package hostintegration

// ProviderRoutingStatus is the server-facing provider status vocabulary shared
// by C11 metadata collection and C10 routing/sync API contracts.
type ProviderRoutingStatus string

const (
	ProviderRoutingAvailable     ProviderRoutingStatus = "available"
	ProviderRoutingLoginRequired ProviderRoutingStatus = "login-required"
	ProviderRoutingUnsupported   ProviderRoutingStatus = "unsupported"
	ProviderRoutingStoreBlocked  ProviderRoutingStatus = "store-blocked"
)

func (s ProviderRoutingStatus) Valid() bool {
	switch s {
	case ProviderRoutingAvailable,
		ProviderRoutingLoginRequired,
		ProviderRoutingUnsupported,
		ProviderRoutingStoreBlocked:
		return true
	default:
		return false
	}
}

func providerRoutingStatusFromToolStatus(status ServerFacingToolStatus) ProviderRoutingStatus {
	if status.ProviderAvailable {
		return ProviderRoutingAvailable
	}
	if status.ProviderLoginStatus == ToolLoginRequired {
		return ProviderRoutingLoginRequired
	}
	return ProviderRoutingUnsupported
}
